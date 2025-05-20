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
	db         *sql.DB
	prefix     string
	structure  map[string]map[string]Column
	statements map[string]*sql.Stmt
	mu         sync.Mutex
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
		db:         db,
		prefix:     conf.SQL.Prefix,
		structure:  make(map[string]map[string]Column),
		statements: make(map[string]*sql.Stmt),
	}

	if err = sdb.addColumnIfNotExists("product", "batch_uid", "VARCHAR(64) NOT NULL"); err != nil {
		return nil, err
	}
	if err = sdb.addColumnIfNotExists("product", "product_uid", "VARCHAR(64) NOT NULL"); err != nil {
		return nil, err
	}
	if err = sdb.addColumnIfNotExists("category", "parent_uid", "VARCHAR(64) NOT NULL"); err != nil {
		return nil, err
	}
	if err = sdb.addColumnIfNotExists("category", "category_uid", "VARCHAR(64) NOT NULL"); err != nil {
		return nil, err
	}
	if err = sdb.addColumnIfNotExists("attribute", "attribute_uid", "VARCHAR(64) NOT NULL"); err != nil {
		return nil, err
	}

	return sdb, nil
}

func (s *MySql) Close() {
	s.closeStmt()
	_ = s.db.Close()
}

func (s *MySql) Stats() string {
	stats := s.db.Stats()
	return fmt.Sprintf("open: %d, inuse: %d, idle: %d, stmts: %d, structure: %d",
		stats.OpenConnections,
		stats.InUse,
		stats.Idle,
		len(s.statements),
		len(s.structure))
}

func (s *MySql) ProductSearch(uid string) (interface{}, error) {
	return s.ReadTable(
		fmt.Sprintf("%sproduct", s.prefix),
		fmt.Sprintf("product_uid='%s'", uid),
		0,
		false)
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
			return fmt.Errorf("category search: %s %v", categoryData.CategoryUID, err)
		}
		parentId, err := s.getCategoryByUID(categoryData.ParentUID)
		if err != nil {
			return fmt.Errorf("parent search: %s %v", categoryData.ParentUID, err)
		}

		category := entity.CategoryFromCategoryData(categoryData)
		category.CategoryId = categoryId
		category.ParentId = parentId

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

func (s *MySql) SaveAttributes(attributes []*entity.Attribute) error {
	for _, attribute := range attributes {
		attributeId, err := s.getAttributeByUID(attribute.Uid)
		if err != nil {
			return fmt.Errorf("attribute search: %v", err)
		}

		if attributeId == 0 {
			attributeId, err = s.addAttribute(attribute)
		} else {
			err = s.updateAttribute(attributeId, attribute)
		}

		for _, attributeDesc := range attribute.Descriptions {
			if err = s.upsertAttributeDescription(attributeId, attributeDesc); err != nil {
				return fmt.Errorf("description: %v", err)
			}
		}

		if err != nil {
			return fmt.Errorf("attribute %s: %v", attribute.Uid, err)
		}
	}
	return nil
}

func (s *MySql) SaveProductAttributes(productAttributes []*entity.ProductAttribute) error {
	for _, productAttribute := range productAttributes {

		err := s.upsertProductAttribute(productAttribute)

		if err != nil {
			return fmt.Errorf("product attribute %s: %v", productAttribute.ProductUid, err)
		}
	}
	return nil
}

func (s *MySql) UpdateProductImage(productUid, image string, isMain bool) error {
	if isMain {
		return s.updateMainProductImage(productUid, image)
	} else {
		return s.updateProductImage(productUid, image)
	}
}

func (s *MySql) updateMainProductImage(productUid, image string) error {
	stmt, err := s.stmtUpdateProductImage()
	if err != nil {
		return err
	}
	_, err = stmt.Exec(image, productUid)
	if err != nil {
		return fmt.Errorf("update: %v", err)
	}
	return nil
}

func (s *MySql) updateProductImage(productUid, image string) error {
	productId, err := s.getProductByUID(productUid)
	if err != nil {
		return err
	}

	if productId == 0 {
		return fmt.Errorf("no product found: %s", productUid)
	}

	stmt, err := s.stmtGetProductNotMainImage()
	if err != nil {
		return err
	}
	var productImageId int
	err = stmt.QueryRow(productId, image).Scan(
		&productImageId,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			userData := map[string]interface{}{
				"product_id": productId,
				"image":      image,
			}

			_, err = s.insert("product_image", userData)
			if err != nil {
				return err
			}
		}
		return err
	}

	return nil
}

func (s *MySql) CheckApiKey(key string) (string, error) {

	stmt, err := s.stmtGetApiUsername()
	if err != nil {
		return "", err
	}
	var userName string
	err = stmt.QueryRow(key).Scan(
		&userName,
	)
	if err != nil {
		return "", err
	}

	return userName, nil
}

func (s *MySql) updateProduct(productId int64, productData *entity.ProductData) error {
	manufacturerId, err := s.getManufacturerId(productData.Manufacturer)
	if err != nil {
		return fmt.Errorf("manufacturer search: %v", err)
	}

	stmt, err := s.stmtUpdateProduct()
	if err != nil {
		return err
	}
	_, err = stmt.Exec(
		productData.Article,
		productData.Quantity,
		productData.StockStatusID(),
		productData.Price,
		manufacturerId,
		productData.Status(),
		time.Now(),
		productData.BatchUid,
		productId)
	if err != nil {
		return fmt.Errorf("update: %v", err)
	}

	err = s.setProductCategories(productId, productData.Categories)
	if err != nil {
		return fmt.Errorf("set categories: %v", err)
	}

	return nil
}

func (s *MySql) setProductCategories(productId int64, categories []string) error {
	if len(categories) == 0 {
		return nil
	}
	query := fmt.Sprintf(`DELETE FROM %sproduct_to_category WHERE product_id=?`, s.prefix)
	_, err := s.db.Exec(query, productId)
	if err != nil {
		return fmt.Errorf("delete: %v", err)
	}

	for _, categoryUID := range categories {
		categoryId, err := s.getCategoryByUID(categoryUID)
		if err != nil {
			return fmt.Errorf("category search: %v", err)
		}

		if err = s.addProductToCategory(productId, categoryId); err != nil {
			return fmt.Errorf("add to category: %v", err)
		}
	}

	return nil
}

// addProductToCategory adds a product to a category with INSERT statement
func (s *MySql) addProductToCategory(productId, categoryId int64) error {
	query := fmt.Sprintf(`INSERT INTO %sproduct_to_category (
                        product_id,
                        category_id)
			VALUES (?, ?)`, s.prefix)
	_, err := s.db.Exec(query, productId, categoryId)
	if err != nil {
		return fmt.Errorf("insert: %v", err)
	}
	return nil
}

func (s *MySql) addProduct(product *entity.ProductData) error {
	manufacturerId, err := s.getManufacturerId(product.Manufacturer)
	if err != nil {
		return fmt.Errorf("manufacturer search: %v", err)
	}

	// known columns
	userData := map[string]interface{}{
		"product_uid":     product.Uid,
		"model":           product.Article,
		"price":           product.Price,
		"manufacturer_id": manufacturerId,
		"quantity":        product.Quantity,
		"status":          product.Status(),
		"stock_status_id": product.StockStatusID(),
		"minimum":         1,
		"subtract":        1,
		"date_available":  time.Now().AddDate(0, 0, -3),
		"shipping":        1,
		"tax_class_id":    9,
		"weight_class_id": 0,
		"length_class_id": 0,
		"date_added":      time.Now(),
		"date_modified":   time.Now(),
		"batch_uid":       product.BatchUid,
	}

	productId, err := s.insert("product", userData)
	if err != nil {
		return err
	}

	if err = s.addProductToStore(productId); err != nil {
		return err
	}

	err = s.setProductCategories(productId, product.Categories)
	if err != nil {
		return fmt.Errorf("set categories: %v", err)
	}

	if err = s.addProductToLayout(productId); err != nil {
		return err
	}

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
	defer rows.Close()

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

func (s *MySql) addAttribute(attribute *entity.Attribute) (int64, error) {
	userData := map[string]interface{}{
		"attribute_uid":      attribute.Uid,
		"attribute_group_id": attribute.GroupId,
		"sort_order":         attribute.SortOrder,
	}

	attributeId, err := s.insert("attribute", userData)
	if err != nil {
		return 0, err
	}

	return attributeId, nil
}

func (s *MySql) updateAttribute(attributeId int64, attribute *entity.Attribute) error {

	stmt, err := s.stmtUpdateAttribute()
	if err != nil {
		return err
	}
	_, err = stmt.Exec(
		attribute.Uid,
		attribute.GroupId,
		attribute.SortOrder,
		attributeId)
	if err != nil {
		return fmt.Errorf("update: %v", err)
	}

	return nil
}

func (s *MySql) findAttributeDescription(attributeId, languageId int64) (*entity.AttributeDescription, error) {
	query := fmt.Sprintf(
		`SELECT
					name
				FROM %sattribute_description 
				WHERE attribute_id=? AND language_id=? LIMIT 1`,
		s.prefix,
	)
	rows, err := s.db.Query(query, attributeId, languageId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var attributeDesc entity.AttributeDescription
		if err = rows.Scan(
			&attributeDesc.Name,
		); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		return &attributeDesc, nil
	}

	return nil, nil
}

func (s *MySql) upsertAttributeDescription(attributeId int64, attributeDescription *entity.AttributeDescription) error {
	var query string
	var err error

	desc, err := s.findAttributeDescription(attributeId, attributeDescription.LanguageId)
	if err != nil {
		return fmt.Errorf("lookup attribute description: %v", err)
	}

	if desc != nil {
		query = fmt.Sprintf(
			`UPDATE %sattribute_description SET
					name = ?
			    WHERE attribute_id = ? AND language_id = ?`,
			s.prefix,
		)
		_, err = s.db.Exec(query,
			attributeDescription.Name,
			attributeId,
			attributeDescription.LanguageId)
		if err != nil {
			return fmt.Errorf("update: %v", err)
		}
	} else {

		userData := map[string]interface{}{
			"attribute_id": attributeId,
			"language_id":  attributeDescription.LanguageId,
			"name":         attributeDescription.Name,
		}

		_, err = s.insert("attribute_description", userData)
		if err != nil {
			return err
		}
	}

	return nil
}
func (s *MySql) findProductAttribute(productId, attributeId, languageId int64) (*entity.ProductAttribute, error) {
	query := fmt.Sprintf(
		`SELECT
					text
				FROM %sproduct_attribute
				WHERE product_id=? AND attribute_id=? AND language_id=? LIMIT 1`,
		s.prefix,
	)
	rows, err := s.db.Query(query,
		productId,
		attributeId,
		languageId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var productAttribute entity.ProductAttribute
		if err = rows.Scan(
			&productAttribute.Text,
		); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		return &productAttribute, nil
	}

	return nil, nil
}

func (s *MySql) upsertProductAttribute(productAttribute *entity.ProductAttribute) error {
	var query string
	var err error

	productId, err := s.getProductByUID(productAttribute.ProductUid)
	if err != nil {
		return fmt.Errorf("product search: %v", err)
	}

	attributeId, err := s.getAttributeByUID(productAttribute.AttributeUid)
	if err != nil {
		return fmt.Errorf("attribute search: %v", err)
	}

	desc, err := s.findProductAttribute(productId, attributeId, productAttribute.LanguageId)
	if err != nil {
		return fmt.Errorf("lookup product attribute: %v", err)
	}

	if desc != nil {
		query = fmt.Sprintf(
			`UPDATE %sproduct_attribute SET
					text = ?
			    WHERE product_id=? AND attribute_id=? AND language_id=?`,
			s.prefix,
		)
		_, err = s.db.Exec(query,
			productAttribute.Text,
			productId,
			attributeId,
			productAttribute.LanguageId)
		if err != nil {
			return fmt.Errorf("update: %v", err)
		}
	} else {

		userData := map[string]interface{}{
			"product_id":   productId,
			"attribute_id": attributeId,
			"language_id":  productAttribute.LanguageId,
			"text":         productAttribute.Text,
		}

		_, err = s.insert("product_attribute", userData)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MySql) getProductByUID(uid string) (int64, error) {
	stmt, err := s.stmtSelectProductId()
	if err != nil {
		return 0, err
	}

	var productId int64
	err = stmt.QueryRow(uid).Scan(&productId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return productId, nil
}

func (s *MySql) getCategoryByUID(uid string) (int64, error) {
	stmt, err := s.stmtCategoryId()
	if err != nil {
		return 0, err
	}
	var categoryId int64
	err = stmt.QueryRow(uid).Scan(&categoryId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			categoryId = 0
			err = nil
		}
		return 0, err
	}
	if categoryId != 0 {
		return categoryId, nil
	}

	userData := map[string]interface{}{
		"category_uid": uid,
		"date_added":   time.Now(),
	}
	categoryId, err = s.insert("category", userData)
	if err != nil {
		return 0, err
	}

	return categoryId, nil
}

func (s *MySql) getAttributeByUID(uid string) (int64, error) {
	stmt, err := s.stmtSelectAttributeId()
	if err != nil {
		return 0, err
	}

	var attributeId int64
	err = stmt.QueryRow(uid).Scan(&attributeId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return attributeId, nil
}

func (s *MySql) updateCategory(category *entity.Category) error {
	stmt, err := s.stmtUpdateCategory()
	if err != nil {
		return err
	}
	_, err = stmt.Exec(
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
	stmt, err := s.stmtCategoryDescription()
	if err != nil {
		return nil, err
	}
	var categoryDescription entity.CategoryDescription
	err = stmt.QueryRow(categoryId, languageId).Scan(
		&categoryDescription.Name,
		&categoryDescription.Description,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &categoryDescription, nil
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
		if categoryDesc.Description != "" {
			stmt, e := s.stmtUpdateCategoryDescription()
			if e != nil {
				return e
			}
			_, err = stmt.Exec(
				categoryDesc.Name,
				categoryDesc.Description,
				categoryDesc.CategoryId,
				categoryDesc.LanguageId)
		} else {
			stmt, e := s.stmtUpdateCategoryName()
			if e != nil {
				return e
			}
			_, err = stmt.Exec(
				categoryDesc.Name,
				categoryDesc.CategoryId,
				categoryDesc.LanguageId)
		}
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

func (s *MySql) getManufacturerId(name string) (int64, error) {
	if name == "" {
		return 0, nil
	}
	stmt, err := s.stmtManufacturerId()
	if err != nil {
		return 0, err
	}
	var manufacturerId int64
	err = stmt.QueryRow(name).Scan(&manufacturerId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			manufacturerId = 0
		} else {
			return 0, err
		}
	}
	if manufacturerId != 0 {
		return manufacturerId, nil
	}

	query := fmt.Sprintf(`INSERT INTO %smanufacturer (name) VALUES (?)`, s.prefix)
	res, err := s.db.Exec(query, name)
	if err != nil {
		return 0, fmt.Errorf("insert: %w", err)
	}

	manufacturerId, err = res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}

	query = fmt.Sprintf(`INSERT INTO %smanufacturer_to_store (manufacturer_id, store_id) VALUES (?, 0)`, s.prefix)
	_, err = s.db.Exec(query, manufacturerId)
	if err != nil {
		return 0, fmt.Errorf("insert: %w", err)
	}

	return manufacturerId, nil
}

func (s *MySql) ReadTable(table, filter string, limit int, plain bool) (interface{}, error) {
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
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("to get columns: %w", err)
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
			if plain {
				rowMap[colName] = columnValues[i]
				continue
			}
			if str, ok := columnValues[i].([]byte); ok {
				decodedValue, e := base64.StdEncoding.DecodeString(string(str))
				if e != nil {
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

func (s *MySql) DeleteRecords(table, filter string) (int64, error) {
	query := fmt.Sprintf("DELETE FROM %s", table)
	if filter != "" {
		query = fmt.Sprintf("%s WHERE %s", query, filter)
	}
	result, err := s.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("exec: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected: %w", err)
	}
	return rowsAffected, nil
}

func (s *MySql) FinalizeProductBatch(batchUid string) (int, error) {
	// calculate the number of products in the batch
	query := fmt.Sprintf("SELECT COUNT(*) FROM %sproduct WHERE batch_uid=?", s.prefix)
	var count int
	err := s.db.QueryRow(query, batchUid).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("batch count: %w", err)
	}
	if count == 0 {
		return 0, fmt.Errorf("empty batch %s", batchUid)
	}
	// deactivate products not in the batch
	query = fmt.Sprintf("UPDATE %sproduct SET status=0 WHERE batch_uid<>?", s.prefix)
	_, err = s.db.Exec(query, batchUid)
	if err != nil {
		return 0, fmt.Errorf("update status: %w", err)
	}
	// remove batch_uid from products
	query = fmt.Sprintf("UPDATE %sproduct SET batch_uid=''", s.prefix)
	_, err = s.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("update batch_uid: %w", err)
	}
	// calculate the number of products that have status=1
	query = fmt.Sprintf("SELECT COUNT(*) FROM %sproduct WHERE status=1", s.prefix)
	err = s.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return count, nil
}
