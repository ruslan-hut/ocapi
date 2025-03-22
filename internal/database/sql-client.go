package database

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"ocapi/entity"
	"ocapi/internal/config"
	"time"
)

type MySql struct {
	db     *sql.DB
	prefix string
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

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	sdb := &MySql{
		db:     db,
		prefix: conf.SQL.Prefix,
	}

	if err = sdb.addColumnIfNotExists("category", "parent_uid", "VARCHAR(64) NOT NULL"); err != nil {
		return nil, err
	}
	if err = sdb.addColumnIfNotExists("category", "category_uid", "VARCHAR(64) NOT NULL"); err != nil {
		return nil, err
	}
	if err = sdb.addColumnIfNotExists("category_description", "seo_keyword", "VARCHAR(64) NOT NULL"); err != nil {
		return nil, err
	}
	if err = sdb.addColumnIfNotExists("product_description", "seo_keyword", "VARCHAR(64) NOT NULL"); err != nil {
		return nil, err
	}
	if err = sdb.addColumnIfNotExists("product", "code", "VARCHAR(64) NOT NULL"); err != nil {
		return nil, err
	}

	return sdb, nil
}

func (s *MySql) Close() {
	_ = s.db.Close()
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
				date_modified = ? 
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

	query := fmt.Sprintf(
		`INSERT INTO %sproduct (
				model,
			    code,
				sku,
			    upc,
			    ean,
			    jan,
			    isbn,
			    mpn,
			    location,
				quantity,
				minimum,
				subtract,
				stock_status_id,
				date_available,
				manufacturer_id,
                shipping,
				price,
                points, 
                weight, 
                weight_class_id,
				length,
				width,
				height,
                length_class_id,
				status,
                tax_class_id,
                sort_order,
			    meta_robots,
			    seo_canonical,
				availableCarriers,
				date_added,
				date_modified)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.prefix)
	// Insert product into the product table
	res, err := s.db.Exec(query,
		product.Model,
		product.Code,
		product.Sku,
		product.Upc,
		product.Ean,
		product.Jan,
		product.Isbn,
		product.Mpn,
		product.Location,
		product.Quantity,
		product.Minimum,
		product.Subtract,
		product.StockStatusId,
		product.DateAvailable,
		product.ManufacturerId,
		product.Shipping,
		product.Price,
		product.Points,
		product.Weight,
		product.WeightClassId,
		product.Length,
		product.Width,
		product.Height,
		product.LengthClassId,
		product.Status,
		product.TaxClassId,
		product.SortOrder,
		"",
		"",
		"",
		product.DateAdded,
		product.DateModified)

	if err != nil {
		return fmt.Errorf("insert: %v", err)
	}

	// Get the last inserted product_id
	productId, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %v", err)
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

func (s *MySql) getProductDescription(productDesc *entity.ProductDescription) ([]*entity.ProductDescription, error) {
	query := fmt.Sprintf(
		`SELECT
					name,
					description
				FROM %sproduct_description 
				WHERE product_id=? AND language_id=?`,
		s.prefix,
	)
	rows, err := s.db.Query(query,
		productDesc.Name,
		productDesc.Description,
		productDesc.ProductUid,
		productDesc.LanguageId)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var productsDesc []*entity.ProductDescription
	for rows.Next() {
		var prodDesc entity.ProductDescription
		if err = rows.Scan(
			&prodDesc.Name,
			&prodDesc.Description,
		); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		productsDesc = append(productsDesc, &prodDesc)
	}

	return productsDesc, nil
}

func (s *MySql) findProductDescription(productId, languageId int64) (*entity.ProductDescription, error) {
	query := fmt.Sprintf(
		`SELECT
					name,
					description,
					meta_title,
					meta_description,
					meta_keyword,
					seo_keyword
				FROM %sproduct_description 
				WHERE product_id=? AND language_id=?`,
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
			&productDesc.MetaTitle,
			&productDesc.MetaDescription,
			&productDesc.MetaKeyword,
			&productDesc.SeoKeyword,
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
		query = fmt.Sprintf(
			`INSERT INTO %sproduct_description (
				product_id,
				language_id,
				name,
				description,
				meta_title,
				meta_description,
				meta_keyword,
                seo_keyword,
				tag)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			s.prefix)

		_, err = s.db.Exec(query,
			productId,
			productDescription.LanguageId,
			productDescription.Name,
			productDescription.Description,
			productDescription.Name,
			"",
			"",
			"",
			"")
		if err != nil {
			return fmt.Errorf("insert: %v", err)
		}
	}

	return nil
}

func (s *MySql) getProductByUID(uid string) (int64, error) {
	query := fmt.Sprintf(
		`SELECT
					product_id
				FROM %sproduct 
				WHERE model=?
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
		var productId int64
		if err = rows.Scan(
			&productId,
		); err != nil {
			return 0, err
		}
		return productId, nil
	}

	return 0, nil
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

	query = fmt.Sprintf(`INSERT INTO %scategory (category_uid, date_added) VALUES (?,?)`, s.prefix)
	res, err := s.db.Exec(query, uid, time.Now())
	if err != nil {
		return 0, fmt.Errorf("insert: %w", err)
	}

	categoryId, _ := res.LastInsertId()

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
					meta_keyword,
					seo_keyword
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
			&categoryDesc.SeoKeyword,
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
		query := fmt.Sprintf(
			`INSERT INTO %scategory_description (
			   category_id,
			   language_id,
			   name,
			   description,
				meta_title,
				meta_description,
				meta_keyword,
				seo_keyword)
			VALUES (?,?,?,?,?,?,?)`,
			s.prefix)

		_, err = s.db.Exec(query,
			categoryDesc.CategoryId,
			categoryDesc.LanguageId,
			categoryDesc.Name,
			categoryDesc.Description,
			categoryDesc.MetaTitle,
			categoryDesc.MetaDescription,
			categoryDesc.MetaKeyword,
			categoryDesc.SeoKeyword,
		)
		if err != nil {
			return fmt.Errorf("insert: %v", err)
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

func (s *MySql) addColumnIfNotExists(tableName, columnName, columnType string) error {
	// Check if the column exists
	query := fmt.Sprintf(`SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = '%s%s' AND COLUMN_NAME = '%s'`,
		s.prefix, tableName, columnName)
	var column string
	err := s.db.QueryRow(query).Scan(&column)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Column does not exist, so add it
			alterQuery := fmt.Sprintf(`ALTER TABLE %s%s ADD COLUMN %s %s`, s.prefix, tableName, columnName, columnType)
			_, err = s.db.Exec(alterQuery)
			if err != nil {
				return fmt.Errorf("add column %s to table %s: %w", columnName, tableName, err)
			}
		} else {
			return fmt.Errorf("checking column %s existence in %s: %w", columnName, tableName, err)
		}
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
