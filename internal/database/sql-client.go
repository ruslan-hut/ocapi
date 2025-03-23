package database

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"ocapi/entity"
	"ocapi/internal/config"
	"sync"
	"time"
)

type MySql struct {
	db                  *sql.DB
	prefix              string
	structure           map[string]map[string]Column
	stmtGetProductByUID *sql.Stmt
	onceUIDStmt         sync.Once
}

func NewSQLClient(conf *config.Config) (*MySql, error) {
	if !conf.SQL.Enabled {
		return nil, nil
	}
	connectionURI := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		conf.SQL.UserName, conf.SQL.Password, conf.SQL.HostName, conf.SQL.Port, conf.SQL.Database)
	db, err := sql.Open("mysql", connectionURI)
	if err != nil {
		return nil, fmt.Errorf("sql connect: %w", err)
	}

	// try ping three times with 30 seconds interval; wait for database to start
	for i := 0; i < 3; i++ {
		if err = db.Ping(); err == nil {
			break
		}
		if i == 2 {
			return nil, fmt.Errorf("ping database: %w", err)
		}
		time.Sleep(30 * time.Second)
	}

	db.SetMaxOpenConns(50)           // макс. кол-во открытых соединений
	db.SetMaxIdleConns(10)           // макс. кол-во "неактивных" соединений в пуле
	db.SetConnMaxLifetime(time.Hour) // время жизни соединения

	sdb := &MySql{
		db:        db,
		prefix:    conf.SQL.Prefix,
		structure: make(map[string]map[string]Column),
	}

	if err = sdb.addColumnIfNotExists("product", "batch_uid", "VARCHAR(64) NOT NULL"); err != nil {
		return nil, err
	}
	if err = sdb.addColumnIfNotExists("category", "parent_uid", "VARCHAR(64) NOT NULL"); err != nil {
		return nil, err
	}
	if err = sdb.addColumnIfNotExists("category", "category_uid", "VARCHAR(64) NOT NULL"); err != nil {
		return nil, err
	}

	return sdb, nil
}

func (s *MySql) Close() {
	if s.stmtGetProductByUID != nil {
		_ = s.stmtGetProductByUID.Close()
	}
	_ = s.db.Close()
}

func (s *MySql) Stats() string {
	stats := s.db.Stats()
	return fmt.Sprintf("open: %d, inuse: %d, idle: %d",
		stats.OpenConnections,
		stats.InUse,
		stats.Idle)
}

func (s *MySql) ProductSearch(model string) ([]*entity.Product, error) {
	query := fmt.Sprintf(
		`SELECT 
					product_id,
					model,
					sku,
					status,
					stock_status_id,
					quantity,
					price,
					manufacturer_id,
					date_modified 
				FROM %sproduct 
				WHERE model=?`,
		s.prefix,
	)
	rows, err := s.db.Query(query, model)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var products []*entity.Product
	for rows.Next() {
		var product entity.Product
		if err = rows.Scan(
			&product.Id,
			&product.Model,
			&product.Sku,
			&product.Status,
			&product.StockStatusId,
			&product.Quantity,
			&product.Price,
			&product.ManufacturerId,
			&product.DateModified,
		); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		products = append(products, &product)
	}
	return products, nil
}

func (s *MySql) SaveProducts(productsData []*entity.ProductData) error {
	for _, productData := range productsData {
		productId, err := s.getProductByUID(productData.Uid)
		if err != nil {
			return fmt.Errorf("product search: %v", err)
		}

		if productId == 0 {
			err = s.addProduct(productData)
		} else {
			err = s.updateProduct(productId, productData)
		}

		if err != nil {
			return fmt.Errorf("product %s: %v", productData.Uid, err)
		}
	}
	return nil
}

func (s *MySql) SaveProductsDescription(productsDescData []*entity.ProductDescription) error {
	for _, productDescData := range productsDescData {

		productId, err := s.getProductByUID(productDescData.ProductUid)
		if err != nil {
			return fmt.Errorf("product search: %v", err)
		}

		if productId == 0 {
			return fmt.Errorf("product decription: uid %s not found", productDescData.ProductUid)
		}

		err = s.upsertProductDescription(productId, productDescData)
		if err != nil {
			return fmt.Errorf("product description %s: %v", productDescData.ProductUid, err)
		}
	}
	return nil
}

func (s *MySql) SaveCategories(categoriesData []*entity.CategoryData) error {
	for _, categoryData := range categoriesData {
		categoryId, err := s.getCategoryByUID(categoryData.CategoryUID)
		if err != nil {
			return fmt.Errorf("category search: %v", err)
		}
		category := entity.CategoryFromCategoryData(categoryData)
		category.CategoryId = categoryId

		err = s.updateCategory(category)

		if err != nil {
			return fmt.Errorf("category [%d] %s: %v", categoryId, categoryData.CategoryUID, err)
		}
	}
	return nil
}

func (s *MySql) SaveCategoriesDescription(categoriesDescData []*entity.CategoryDescriptionData) error {
	for _, categoryDescData := range categoriesDescData {
		categoryId, err := s.getCategoryByUID(categoryDescData.CategoryUid)
		if err != nil {
			return fmt.Errorf("category search: %v", err)
		}
		category := entity.CategoryDescriptionFromCategoryDescriptionData(categoryDescData)
		category.CategoryId = categoryId
		err = s.upsertCategoryDescription(category)

		if err != nil {
			return fmt.Errorf("category [%d] %s: %v", categoryId, categoryDescData.CategoryUid, err)
		}
	}
	return nil
}

func (s *MySql) UpdateProductImage(productUid string, image string) error {
	productId, err := s.getProductByUID(productUid)
	if err != nil {
		return fmt.Errorf("product search: %v", err)
	}

	if productId == 0 {
		return fmt.Errorf("product decription: uid %s not found", productUid)
	}

	query := fmt.Sprintf(
		`UPDATE %sproduct SET
				image = ?
			    WHERE product_id = ?`,
		s.prefix,
	)

	_, err = s.db.Exec(query,
		image,
		productId)
	if err != nil {
		return fmt.Errorf("update: %v", err)
	}

	return nil
}

func (s *MySql) updateProduct(productId int64, productData *entity.ProductData) error {

	product := entity.ProductFromProductData(productData)

	manufacturerId, err := s.getManufacturerId(productData.Manufacturer)
	if err != nil {
		return fmt.Errorf("manufacturer search: %v", err)
	}

	query := fmt.Sprintf(
		`UPDATE %sproduct SET
				sku = ?, 
				quantity = ?, 
                stock_status_id = ?,
				price = ?, 
				manufacturer_id = ?, 
				status = ?, 
				date_modified = ?,
                batch_uid = ?
			    WHERE product_id = ?`,
		s.prefix,
	)

	_, err = s.db.Exec(query,
		product.Sku,
		product.Quantity,
		product.StockStatusId,
		product.Price,
		manufacturerId,
		product.Status,
		time.Now(),
		product.BatchUID,
		productId)
	if err != nil {
		return fmt.Errorf("update: %v", err)
	}

	categoryId, err := s.getCategoryByUID(productData.CategoryUid)
	if err == nil {
		if err = s.addProductToCategory(productId, categoryId); err != nil {
			return fmt.Errorf("add to category: %v", err)
		}
	}

	return nil
}

func (s *MySql) addProduct(productData *entity.ProductData) error {

	product := entity.ProductFromProductData(productData)
	product.DateAdded = time.Now()

	manufacturerId, err := s.getManufacturerId(productData.Manufacturer)
	if err != nil {
		return fmt.Errorf("manufacturer search: %v", err)
	}
	product.ManufacturerId = manufacturerId

	// known columns
	userData := map[string]interface{}{
		"model":           product.Model,
		"sku":             product.Sku,
		"price":           product.Price,
		"manufacturer_id": product.ManufacturerId,
		"quantity":        product.Quantity,
		"status":          product.Status,
		"stock_status_id": product.StockStatusId,
		"minimum":         product.Minimum,
		"subtract":        product.Subtract,
		"date_available":  product.DateAvailable,
		"shipping":        product.Shipping,
		"tax_class_id":    product.TaxClassId,
		"weight_class_id": product.WeightClassId,
		"length_class_id": product.LengthClassId,
		"date_added":      product.DateAdded,
		"date_modified":   product.DateModified,
		"batch_uid":       product.BatchUID,
	}

	productId, err := s.insert("product", userData)
	if err != nil {
		return err
	}

	if err = s.addProductToStore(productId); err != nil {
		return err
	}

	categoryId, err := s.getCategoryByUID(productData.CategoryUid)
	if err == nil {
		if err = s.addProductToCategory(productId, categoryId); err != nil {
			return err
		}
	}

	if err = s.addProductToLayout(productId); err != nil {
		return err
	}

	// Insert SEO URL
	//seoURL := s.TransLit(product.Model)
	//seoURL = s.MetaURL(seoURL)
	//seoURL = fmt.Sprintf("%s", seoURL)
	//_, err = s.db.ExecContext(s.ctx, `
	//	INSERT INTO seo_url (store_id, language_id, query, keyword) VALUES
	//	('0', 1, ?, ?), ('0', 2, ?, ?), ('0', 3, ?, ?)`,
	//	fmt.Sprintf("product_id=%d", productId), seoURL,
	//	fmt.Sprintf("product_id=%d", productId), seoURL,
	//	fmt.Sprintf("product_id=%d", productId), seoURL)
	//if err != nil {
	//	return fmt.Errorf("failed to insert SEO URL: %v", err)
	//}

	//err = s.disActivateProducts(nowDate)
	//if err != nil {
	//	return fmt.Errorf("failed to disactivate products: %v", err)
	//}
	return nil
}

func (s *MySql) addProductToStore(productID int64) error {
	query := fmt.Sprintf(
		`INSERT INTO %sproduct_to_store (
				product_id,
				store_id)
			VALUES (?, ?)`,
		s.prefix)

	_, err := s.db.Exec(query,
		productID, 0)

	if err != nil {
		return fmt.Errorf("product to store: %v", err)
	}

	return nil
}

func (s *MySql) addProductToLayout(productID int64) error {
	query := fmt.Sprintf(
		`INSERT INTO %sproduct_to_layout (
				product_id,
				store_id,
                layout_id)
			VALUES (?, ?, ?)`,
		s.prefix)

	_, err := s.db.Exec(query,
		productID, 0, 0)

	if err != nil {
		return fmt.Errorf("product to layout: %v", err)
	}

	return nil
}

func (s *MySql) findProductDescription(productId, languageId int64) (*entity.ProductDescription, error) {
	query := fmt.Sprintf(
		`SELECT
					name,
					description
				FROM %sproduct_description 
				WHERE product_id=? AND language_id=? LIMIT 1`,
		s.prefix,
	)
	rows, err := s.db.Query(query, productId, languageId)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var productDesc entity.ProductDescription
		if err = rows.Scan(
			&productDesc.Name,
			&productDesc.Description,
		); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		return &productDesc, nil
	}

	return nil, nil
}

func (s *MySql) upsertProductDescription(productId int64, productDescription *entity.ProductDescription) error {
	var query string
	var err error

	desc, err := s.findProductDescription(productId, productDescription.LanguageId)
	if err != nil {
		return fmt.Errorf("lookup description: %v", err)
	}

	if desc != nil {
		if productDescription.Description == "" {
			query = fmt.Sprintf(
				`UPDATE %sproduct_description SET
					name = ?
			    WHERE product_id = ? AND language_id = ?`,
				s.prefix,
			)
			_, err = s.db.Exec(query,
				productDescription.Name,
				productId,
				productDescription.LanguageId)
		} else {
			query = fmt.Sprintf(
				`UPDATE %sproduct_description SET
					name = ?,
					description = ?
			    WHERE product_id = ? AND language_id = ?`,
				s.prefix,
			)
			_, err = s.db.Exec(query,
				productDescription.Name,
				productDescription.Description,
				productId,
				productDescription.LanguageId)
		}
		if err != nil {
			return fmt.Errorf("update: %v", err)
		}
	} else {

		userData := map[string]interface{}{
			"product_id":  productId,
			"language_id": productDescription.LanguageId,
			"name":        productDescription.Name,
			"description": productDescription.Description,
			"meta_title":  productDescription.Name,
		}

		_, err = s.insert("product_description", userData)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MySql) getProductByUID(uid string) (int64, error) {
	var initErr error
	s.onceUIDStmt.Do(func() {
		query := fmt.Sprintf(`SELECT product_id FROM %sproduct WHERE model=? LIMIT 1`, s.prefix)
		s.stmtGetProductByUID, initErr = s.db.Prepare(query)
	})
	if initErr != nil {
		return 0, initErr
	}
	var productId int64
	err := s.stmtGetProductByUID.QueryRow(uid).Scan(&productId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return productId, nil
}

func (s *MySql) getCategoryByUID(uid string) (int64, error) {
	query := fmt.Sprintf(
		`SELECT
					category_id
				FROM %scategory 
				WHERE category_uid=?
				LIMIT 1`,
		s.prefix,
	)
	rows, err := s.db.Query(query, uid)
	if err != nil {
		return 0, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var categoryId int64
		if err = rows.Scan(
			&categoryId,
		); err != nil {
			return 0, err
		}
		return categoryId, nil
	}

	userData := map[string]interface{}{
		"category_uid": uid,
		"date_added":   time.Now(),
	}
	categoryId, err := s.insert("category", userData)
	if err != nil {
		return 0, err
	}

	return categoryId, nil
}

func (s *MySql) updateCategory(category *entity.Category) error {
	query := fmt.Sprintf(
		`UPDATE %scategory SET
                        parent_id=?,
                        parent_uid=?,
                        top=?,
                        sort_order=?,
                        status=?,
                        date_modified=?
			    WHERE category_id = ?`,
		s.prefix,
	)
	_, err := s.db.Query(query,
		category.ParentId,
		category.ParentUID,
		category.Top,
		category.SortOrder,
		category.Status,
		category.DateModified,
		category.CategoryId)
	if err != nil {
		return fmt.Errorf("update: %v", err)
	}

	return nil
}

func (s *MySql) findCategoryDescription(categoryId, languageId int64) (*entity.CategoryDescription, error) {
	query := fmt.Sprintf(
		`SELECT
					name,
					description,
					meta_title,
					meta_description,
					meta_keyword
				FROM %scategory_description 
				WHERE category_id=? AND language_id=?`,
		s.prefix,
	)
	rows, err := s.db.Query(query, categoryId, languageId)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var categoryDesc entity.CategoryDescription
		if err = rows.Scan(
			&categoryDesc.Name,
			&categoryDesc.Description,
			&categoryDesc.MetaTitle,
			&categoryDesc.MetaDescription,
			&categoryDesc.MetaKeyword,
		); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		return &categoryDesc, nil
	}

	return nil, nil
}

func (s *MySql) upsertCategoryDescription(categoryDesc *entity.CategoryDescription) error {
	if categoryDesc.CategoryId == 0 {
		return fmt.Errorf("category id not provided")
	}
	if categoryDesc.LanguageId == 0 {
		return fmt.Errorf("language id not provided")
	}

	description, err := s.findCategoryDescription(categoryDesc.CategoryId, categoryDesc.LanguageId)
	if err != nil {
		return fmt.Errorf("lookup description: %v", err)
	}

	if description != nil {
		query := fmt.Sprintf(`UPDATE %scategory_description SET name=?, description=?
			    WHERE category_id=? AND language_id=?`, s.prefix,
		)
		_, err = s.db.Exec(query,
			categoryDesc.Name,
			categoryDesc.Description,
			categoryDesc.CategoryId,
			categoryDesc.LanguageId)
		if err != nil {
			return fmt.Errorf("update: %v", err)
		}
	} else {
		userData := map[string]interface{}{
			"category_id":      categoryDesc.CategoryId,
			"language_id":      categoryDesc.LanguageId,
			"name":             categoryDesc.Name,
			"description":      categoryDesc.Description,
			"meta_title":       categoryDesc.Name,
			"meta_description": categoryDesc.Name,
		}
		_, err = s.insert("category_description", userData)
		if err != nil {
			return err
		}
	}
	return nil
}

// Helper function to add product to category
func (s *MySql) addProductToCategory(productId, categoryId int64) error {

	query := fmt.Sprintf(`DELETE FROM %sproduct_to_category WHERE product_id=?`, s.prefix)
	_, err := s.db.Exec(query, productId)
	if err != nil {
		return fmt.Errorf("delete: %v", err)
	}

	query = fmt.Sprintf(`INSERT INTO %sproduct_to_category (
                        product_id,
                        category_id)
			VALUES (?, ?)`, s.prefix)
	_, err = s.db.Exec(query, productId, categoryId)
	if err != nil {
		return fmt.Errorf("insert: %v", err)
	}
	return nil
}

func (s *MySql) getManufacturerId(name string) (int64, error) {
	if name == "" {
		return 0, nil
	}
	query := fmt.Sprintf(
		`SELECT
					manufacturer_id
				FROM %smanufacturer 
				WHERE name=?
				LIMIT 1`,
		s.prefix,
	)
	rows, err := s.db.Query(query, name)
	if err != nil {
		return 0, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var manufacturerId int64
		if err = rows.Scan(manufacturerId); err != nil {
			return 0, err
		}
		return manufacturerId, nil
	}

	query = fmt.Sprintf(`INSERT INTO %smanufacturer (name) VALUES (?)`, s.prefix)
	res, err := s.db.Exec(query, name)
	if err != nil {
		return 0, fmt.Errorf("insert: %w", err)
	}

	manufacturerId, _ := res.LastInsertId()

	query = fmt.Sprintf(`INSERT INTO %smanufacturer_to_store (manufacturer_id,Store_id) VALUES (?,0)`, s.prefix)
	res, err = s.db.Exec(query, name)
	if err != nil {
		return 0, fmt.Errorf("insert: %w", err)
	}

	return manufacturerId, nil
}

func (s *MySql) ReadTable(table, filter string, limit int) (interface{}, error) {
	query := fmt.Sprintf("SELECT * FROM %s", table)
	if filter != "" {
		query = fmt.Sprintf("%s WHERE %s", query, filter)
	}
	if limit > 0 {
		query = fmt.Sprintf("%s LIMIT %d", query, limit)
	}
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		columnPointers := make([]interface{}, len(columns))
		columnValues := make([]interface{}, len(columns))
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}

		if err = rows.Scan(columnPointers...); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		rowMap := make(map[string]interface{})
		for i, colName := range columns {
			if str, ok := columnValues[i].([]byte); ok {
				decodedValue, err := base64.StdEncoding.DecodeString(string(str))
				if err != nil {
					rowMap[colName] = string(str)
				} else {
					rowMap[colName] = string(decodedValue)
				}
			} else {
				rowMap[colName] = columnValues[i]
			}
		}
		results = append(results, rowMap)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return results, nil
}

// Placeholder for the TransLit function
func (s *MySql) TransLit(input string) string {
	// Transliteration logic here
	return input
}

// Placeholder for the MetaURL function
func (s *MySql) MetaURL(input string) string {
	// Meta URL logic here
	return input
}

func (s *MySql) disActivateProducts(now string) error {
	//_, err := s.db.ExecContext(s.ctx, `
	//	UPDATE ?product SET status = 0 WHERE date_modified < ?`,
	//	s.prefix, now)
	return nil
}
