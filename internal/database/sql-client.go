package database

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"ocapi/entity"
	"ocapi/internal/config"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

type MySql struct {
	db           *sql.DB
	prefix       string
	structure    map[string]map[string]Column
	statements   map[string]*sql.Stmt
	mu           sync.Mutex
	customFields map[string]bool // allowed custom field names for products
}

// NewSQLClient creates a new MySQL client, establishes the connection, configures the pool,
// and ensures that all required custom columns exist in the OpenCart database schema.
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

	// try to ping three times with a 30-second interval; wait for a database to start
	for i := 0; i < 3; i++ {
		if err = db.Ping(); err == nil {
			break
		}
		if i == 2 {
			return nil, fmt.Errorf("ping database: %w", err)
		}
		time.Sleep(30 * time.Second)
	}

	db.SetMaxOpenConns(50)           // max number of open connections
	db.SetMaxIdleConns(10)           // max number of idle connections in the pool
	db.SetConnMaxLifetime(time.Hour) // connection max lifetime

	// Initialize allowed custom fields with defaults
	customFields := map[string]bool{
		"sku":      true,
		"upc":      true,
		"ean":      true,
		"jan":      true,
		"isbn":     true,
		"mpn":      true,
		"location": true,
	}
	// Add configured custom fields
	for _, field := range conf.Product.CustomFields {
		customFields[field] = true
	}

	sdb := &MySql{
		db:           db,
		prefix:       conf.SQL.Prefix,
		structure:    make(map[string]map[string]Column),
		statements:   make(map[string]*sql.Stmt),
		customFields: customFields,
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
	if err = sdb.addColumnIfNotExists("product_image", "file_uid", "VARCHAR(64) NOT NULL"); err != nil {
		return nil, err
	}

	return sdb, nil
}

// Close closes all prepared statements and the database connection.
func (s *MySql) Close() {
	s.closeStmt()
	_ = s.db.Close()
}

// Stats returns a formatted string with current database connection pool statistics.
func (s *MySql) Stats() string {
	stats := s.db.Stats()
	return fmt.Sprintf("open: %d, inuse: %d, idle: %d, stmts: %d, structure: %d",
		stats.OpenConnections,
		stats.InUse,
		stats.Idle,
		len(s.statements),
		len(s.structure))
}

// ProductSearch returns all product table columns for a product identified by its UID.
func (s *MySql) ProductSearch(uid string) (interface{}, error) {
	return s.ReadTable(
		fmt.Sprintf("%sproduct", s.prefix),
		fmt.Sprintf("product_uid='%s'", uid),
		0,
		false)
}

// SaveProducts upserts a batch of products: creates new ones or updates existing by UID.
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

// SaveProductsDescription upserts descriptions for a batch of products identified by UID.
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

// SaveProductSpecial upserts special price records for a batch of products.
func (s *MySql) SaveProductSpecial(products []*entity.ProductSpecial) error {
	for _, special := range products {

		productId, err := s.getProductByUID(special.ProductUid)
		if err != nil {
			return fmt.Errorf("product search: %v", err)
		}

		if productId == 0 {
			return fmt.Errorf("product special: uid %s not found", special.ProductUid)
		}

		err = s.upsertProductSpecial(productId, special)
		if err != nil {
			return fmt.Errorf("product special %s: %v", special.ProductUid, err)
		}
	}
	return nil
}

// SaveCategories upserts a batch of categories, resolving parent UIDs to IDs.
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

// SaveCategoriesDescription upserts descriptions for a batch of categories.
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

// SaveAttributes upserts a batch of attributes and their descriptions.
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

// SaveProductAttributes upserts attribute values for a batch of product-attribute pairs.
func (s *MySql) SaveProductAttributes(productAttributes []*entity.ProductAttribute) error {
	for _, productAttribute := range productAttributes {

		err := s.upsertProductAttribute(productAttribute)

		if err != nil {
			return fmt.Errorf("product attribute %s: %v", productAttribute.ProductUid, err)
		}
	}
	return nil
}

// UpdateProductImage routes the image update to either the main product image or an additional image handler.
func (s *MySql) UpdateProductImage(imageData *entity.ProductImageData) error {
	if imageData.IsMain {
		return s.updateMainProductImage(imageData)
	}
	return s.updateProductImage(imageData)
}

// updateMainProductImage sets the main image URL on the product record.
func (s *MySql) updateMainProductImage(imageData *entity.ProductImageData) error {
	stmt, err := s.stmtUpdateProductImage()
	if err != nil {
		return err
	}
	_, err = stmt.Exec(imageData.ImageUrl, imageData.ProductUid)
	if err != nil {
		return fmt.Errorf("update: %v", err)
	}
	return nil
}

// updateProductImage inserts or updates an additional (non-main) product image in the product_image table.
// If the image with the given file_uid does not exist, it inserts a new row; otherwise updates sort_order.
func (s *MySql) updateProductImage(imageData *entity.ProductImageData) error {
	productId, err := s.getProductByUID(imageData.ProductUid)
	if err != nil {
		return err
	}

	if productId == 0 {
		return fmt.Errorf("no product found: %s", imageData.ProductUid)
	}

	stmt, err := s.stmtGetProductNotMainImage()
	if err != nil {
		return err
	}
	var productImageId int
	err = stmt.QueryRow(productId, imageData.FileUid).Scan(
		&productImageId,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			userData := map[string]interface{}{
				"product_id": productId,
				"file_uid":   imageData.FileUid,
				"image":      imageData.ImageUrl,
				"sort_order": imageData.SortOrder,
			}

			_, err = s.insert("product_image", userData)
		}
		return err
	}

	// update sort order if the image is already in DB
	state, err := s.stmtUpdateProductImageAdd()
	if err != nil {
		return err
	}
	_, err = state.Exec(imageData.SortOrder, productImageId)

	return err
}

// CleanUpProductImages removes additional product images that are not present in the provided list.
// It also deduplicates images with the same file_uid. Returns a set of file UIDs that already
// exist in the database (and were kept), so the caller can determine which UIDs are new.
func (s *MySql) CleanUpProductImages(productUid string, images []string) (map[string]bool, error) {
	// 1) Get the product ID by UID
	productId, err := s.getProductByUID(productUid)
	if err != nil {
		return nil, err
	}
	if productId == 0 {
		return nil, fmt.Errorf("no product found: %s", productUid)
	}

	// 2) Query existing (product_image_id, file_uid) pairs from the database
	stmt, err := s.stmtGetProductImages()
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(productId)
	if err != nil {
		return nil, fmt.Errorf("select: %v", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	// 3) Build a set from the input image list for fast lookup
	validImages := make(map[string]bool, len(images))
	for _, uid := range images {
		validImages[uid] = true
	}

	// 4) Iterate over existing rows and decide which product_image_id to delete:
	//    - if file_uid is not in validImages
	//    - or if the file_uid was already seen (duplicate)
	seen := make(map[string]bool, len(images))
	var idsToDelete []int64

	for rows.Next() {
		var (
			imgID   int64
			fileUid string
		)
		if err = rows.Scan(&imgID, &fileUid); err != nil {
			return nil, fmt.Errorf("scan: %v", err)
		}

		// not in the valid list — mark for deletion
		if !validImages[fileUid] {
			idsToDelete = append(idsToDelete, imgID)
			continue
		}
		// already seen this file_uid — duplicate — mark for deletion
		if seen[fileUid] {
			idsToDelete = append(idsToDelete, imgID)
		} else {
			seen[fileUid] = true
		}
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %v", err)
	}

	// 5) If there are rows to delete, execute a single batch DELETE by product_image_id
	if len(idsToDelete) > 0 {
		placeholders := make([]string, len(idsToDelete))
		args := make([]interface{}, len(idsToDelete))
		for i, id := range idsToDelete {
			placeholders[i] = "?"
			args[i] = id
		}

		delQuery := fmt.Sprintf(
			`DELETE FROM %sproduct_image
             WHERE product_image_id IN (%s)`,
			s.prefix,
			strings.Join(placeholders, ","),
		)

		if _, err = s.db.Exec(delQuery, args...); err != nil {
			return nil, fmt.Errorf("delete by id: %v", err)
		}
	}

	return seen, nil
}

// InsertProductImage adds a new additional image row into the product_image table for the given product.
func (s *MySql) InsertProductImage(productUid string, fileUid string, imageUrl string, sortOrder int) error {
	productId, err := s.getProductByUID(productUid)
	if err != nil {
		return err
	}
	if productId == 0 {
		return fmt.Errorf("no product found: %s", productUid)
	}

	userData := map[string]interface{}{
		"product_id": productId,
		"file_uid":   fileUid,
		"image":      imageUrl,
		"sort_order": sortOrder,
	}

	_, err = s.insert("product_image", userData)
	return err
}

// GetAllImages returns all image paths from the product and product_image tables (used for orphan cleanup).
func (s *MySql) GetAllImages() ([]string, error) {
	query := fmt.Sprintf(`SELECT image FROM %sproduct UNION SELECT image FROM %sproduct_image`, s.prefix, s.prefix)
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var images []string
	for rows.Next() {
		var image sql.NullString
		if err = rows.Scan(
			&image,
		); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		if image.Valid {
			images = append(images, image.String)
		}
	}
	return images, nil
}

// CheckApiKey looks up an API key in the database and returns the associated username.
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

// updateProduct updates an existing product record, its category links, and custom fields.
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
		//productData.Sku,
		//productData.Upc,
		//productData.Ean,
		//productData.Jan,
		//productData.Isbn,
		//productData.Mpn,
		//productData.Location,
		productData.Quantity,
		productData.StockStatusID(),
		productData.Price,
		manufacturerId,
		productData.Status(),
		productData.Weight,
		productData.WeightClassId,
		time.Now(),
		productData.BatchUid,
		productId)
	if err != nil {
		return fmt.Errorf("update: %v", err)
	}

	err = s.setProductCategories(productId, productData.Categories)
	if err != nil {
		return err
	}

	err = s.updateCustomFields(productId, productData)
	if err != nil {
		return err
	}

	return nil
}

// updateCustomFields applies whitelisted custom field updates to a product record.
func (s *MySql) updateCustomFields(productId int64, productData *entity.ProductData) error {
	if len(productData.CustomFields) > 0 {
		for _, field := range productData.CustomFields {
			// Validate field name against whitelist (defaults + configured)
			if !s.customFields[field.FieldName] {
				return fmt.Errorf("custom field not allowed: %s", field.FieldName)
			}
			data := map[string]interface{}{
				field.FieldName: field.FieldValue,
			}
			err := s.update("product", data, "product_id=?", productId)
			if err != nil {
				return fmt.Errorf("update custom field %s: %v", field.FieldName, err)
			}
		}
	}
	return nil
}

// setProductCategories replaces all category associations for a product with the given category UIDs.
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

// addProductToCategory adds a product to a category with an INSERT statement
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

// addProduct inserts a new product record along with its store, layout, category, and custom field associations.
func (s *MySql) addProduct(product *entity.ProductData) error {
	manufacturerId, err := s.getManufacturerId(product.Manufacturer)
	if err != nil {
		return fmt.Errorf("manufacturer search: %v", err)
	}

	userData := map[string]interface{}{
		"product_uid": product.Uid,
		"model":       product.Article,
		//"sku":             product.Sku,
		//"upc":             product.Upc,
		//"ean":             product.Ean,
		//"jan":             product.Jan,
		//"isbn":            product.Isbn,
		//"mpn":             product.Mpn,
		//"location":        product.Location,
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
		"weight":          product.Weight,
		"weight_class_id": product.WeightClassId,
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

	err = s.updateCustomFields(productId, product)
	if err != nil {
		return err
	}

	return nil
}

// addProductToStore links a product to the default store (store_id=0).
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

// addCategoryToStore links a category to the default store (store_id=0).
func (s *MySql) addCategoryToStore(categoryID int64) error {
	query := fmt.Sprintf(
		`INSERT INTO %scategory_to_store (
				category_id,
				store_id)
			VALUES (?, ?)`,
		s.prefix)

	_, err := s.db.Exec(query,
		categoryID, 0)

	if err != nil {
		return fmt.Errorf("category to store: %v", err)
	}

	return nil
}

// addProductToLayout links a product to the default layout (store_id=0, layout_id=0).
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

// findProductDescription looks up a product description by product ID and language ID. Returns nil if not found.
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

// upsertProductDescription creates or updates a product description for the given product and language.
// If the description exists and UpdateDescription is false, only the name is updated.
func (s *MySql) upsertProductDescription(productId int64, productDescription *entity.ProductDescription) error {
	var query string
	var err error

	desc, err := s.findProductDescription(productId, productDescription.LanguageId)
	if err != nil {
		return fmt.Errorf("lookup description: %v", err)
	}

	if desc != nil {
		if productDescription.Description == "" || !productDescription.UpdateDescription {
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

// findProductSpecial checks whether a special price record exists for the given product and customer group.
func (s *MySql) findProductSpecial(productId int64, customerGroupId int64) (bool, error) {
	stmt, err := s.stmtFindProductSpecial()
	if err != nil {
		return false, err
	}

	rows, err := stmt.Query(productId, customerGroupId)
	if err != nil {
		return false, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	return rows.Next(), nil
}

// upsertProductSpecial creates or updates a product special price record in the database.
// It first checks if a record exists for the given product ID and customer group ID.
// If it exists, the function updates the existing record with the new special price data.
// If it doesn't exist, a new record is inserted.
//
// Parameters:
//   - productId: The ID of the product to which the special price applies
//   - special: Pointer to the ProductSpecial entity containing price, date range, priority, and customer group data
//
// Returns:
//   - error: Any error that occurred during the database operation, or nil if successful
func (s *MySql) upsertProductSpecial(productId int64, special *entity.ProductSpecial) error {

	// Check if a record exists
	exists, err := s.findProductSpecial(productId, special.GroupId)
	if err != nil {
		return fmt.Errorf("lookup product special: %v", err)
	}

	if exists {
		// Update existing record
		stmt, err := s.stmtUpdateProductSpecial()
		if err != nil {
			return err
		}
		_, err = stmt.Exec(
			special.Price,
			special.Priority,
			special.DateStart,
			special.DateEnd,
			productId,
			special.GroupId)
		if err != nil {
			return fmt.Errorf("update: %v", err)
		}
	} else {
		// Insert a new record
		specialData := map[string]interface{}{
			"product_id":        productId,
			"customer_group_id": special.GroupId,
			"price":             special.Price,
			"date_start":        special.DateStart,
			"date_end":          special.DateEnd,
			"priority":          special.Priority,
		}
		_, err = s.insert("product_special", specialData)
		if err != nil {
			return fmt.Errorf("insert: %v", err)
		}
	}

	return nil
}

// addAttribute inserts a new attribute record and returns its auto-generated ID.
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

// updateAttribute updates an existing attribute's UID, group ID, and sort order.
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

// findAttributeDescription looks up an attribute description by attribute ID and language ID. Returns nil if not found.
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
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

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

// upsertAttributeDescription creates or updates an attribute description for the given attribute and language.
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

// findProductAttribute looks up a product attribute value by product, attribute, and language IDs. Returns nil if not found.
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
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

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

// upsertProductAttribute creates or updates a product attribute text value, resolving product and attribute UIDs to IDs.
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

// getProductByUID returns the product_id for a given product UID, or 0 if not found.
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

// getCategoryByUID returns the category_id for a given category UID.
// If the category does not exist, it creates a new one and returns its ID.
func (s *MySql) getCategoryByUID(uid string) (int64, error) {
	if uid == "" {
		return 0, nil
	}

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
		} else {
			return 0, err
		}
	}
	if categoryId != 0 {
		return categoryId, nil
	}

	userData := map[string]interface{}{
		"category_uid":  uid,
		"date_added":    time.Now(),
		"date_modified": time.Now(),
	}
	categoryId, err = s.insert("category", userData)
	if err != nil {
		return 0, err
	}

	_ = s.addCategoryToStore(categoryId)

	return categoryId, nil
}

// getAttributeByUID returns the attribute_id for a given attribute UID, or 0 if not found.
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

// updateCategory updates an existing category's parent, sort order, status, and other fields.
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

// findCategoryDescription looks up a category description by category ID and language ID. Returns nil if not found.
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

// upsertCategoryDescription creates or updates a category description for the given category and language.
// If the description is empty, only the name is updated.
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

// getManufacturerId returns the manufacturer_id for a given name.
// If the manufacturer does not exist, it creates a new one with a default store association.
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
		return 0, err
	}

	manufacturerId, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}

	query = fmt.Sprintf(`INSERT INTO %smanufacturer_to_store (manufacturer_id, store_id) VALUES (?, 0)`, s.prefix)
	_, err = s.db.Exec(query, manufacturerId)
	if err != nil {
		return 0, err
	}

	return manufacturerId, nil
}

// allowedReadTables defines tables that can be queried via ReadTable
var allowedReadTables = map[string]bool{
	"product": true, "product_description": true, "product_image": true,
	"product_attribute": true, "product_special": true, "product_to_category": true,
	"category": true, "category_description": true,
	"order": true, "order_product": true, "order_total": true, "order_history": true,
	"attribute": true, "attribute_description": true,
	"manufacturer": true, "currency": true,
}

// dangerousSQLPatterns contains patterns that indicate SQL injection attempts
var dangerousSQLPatterns = []string{
	";", "--", "/*", "*/", "DROP", "DELETE", "UPDATE", "INSERT",
	"TRUNCATE", "ALTER", "CREATE", "UNION", "EXEC", "EXECUTE",
	"xp_", "sp_", "0x", "CHAR(", "CONCAT(",
}

// containsDangerousSQL checks if a filter string contains potential SQL injection patterns
func containsDangerousSQL(filter string) bool {
	upper := strings.ToUpper(filter)
	for _, pattern := range dangerousSQLPatterns {
		if strings.Contains(upper, pattern) {
			return true
		}
	}
	return false
}

// ReadTable performs a generic SELECT on a whitelisted table with optional filter and limit.
// If plain is false, byte columns are decoded from base64 where possible.
func (s *MySql) ReadTable(table, filter string, limit int, plain bool) (interface{}, error) {
	// Strip prefix if provided (e.g., "oc_product" -> "product")
	tableName := table
	if s.prefix != "" && strings.HasPrefix(table, s.prefix) {
		tableName = strings.TrimPrefix(table, s.prefix)
	}

	// Validate table name against whitelist
	if !allowedReadTables[tableName] {
		return nil, fmt.Errorf("table not allowed: %s", table)
	}

	// Block dangerous SQL patterns in filter
	if filter != "" && containsDangerousSQL(filter) {
		return nil, fmt.Errorf("invalid filter: contains forbidden pattern")
	}

	query := fmt.Sprintf("SELECT * FROM %s%s", s.prefix, tableName)
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
		return nil, err
	}

	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		columnPointers := make([]interface{}, len(columns))
		columnValues := make([]interface{}, len(columns))
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}

		if err = rows.Scan(columnPointers...); err != nil {
			return nil, err
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
		return nil, err
	}

	return results, nil
}

// FinalizeProductBatch deactivates all products not in the given batch, clears batch markers,
// and returns the count of active products remaining.
func (s *MySql) FinalizeProductBatch(batchUid string) (int, error) {
	// count products in the batch
	query := fmt.Sprintf("SELECT COUNT(*) FROM %sproduct WHERE batch_uid=?", s.prefix)
	var count int
	err := s.db.QueryRow(query, batchUid).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("batch count: %w", err)
	}
	if count == 0 {
		return 0, fmt.Errorf("empty batch %s", batchUid)
	}

	// deactivate all products not belonging to this batch
	query = fmt.Sprintf("UPDATE %sproduct SET status=0 WHERE batch_uid<>?", s.prefix)
	_, err = s.db.Exec(query, batchUid)
	if err != nil {
		return 0, fmt.Errorf("update status: %w", err)
	}

	// clear batch markers from all products
	query = fmt.Sprintf("UPDATE %sproduct SET batch_uid=''", s.prefix)
	_, err = s.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("update batch_uid: %w", err)
	}

	// count remaining active products
	query = fmt.Sprintf("SELECT COUNT(*) FROM %sproduct WHERE status=1", s.prefix)
	err = s.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return count, nil
}

// OrderSearchId retrieves a full order record by its ID. Returns nil if not found.
func (s *MySql) OrderSearchId(orderId int64) (*entity.Order, error) {
	stmt, err := s.stmtSelectOrder()
	if err != nil {
		return nil, err
	}
	var order entity.Order
	err = stmt.QueryRow(orderId).Scan(
		&order.OrderID,
		&order.InvoiceNo,
		&order.InvoicePrefix,
		&order.StoreID,
		&order.StoreName,
		&order.StoreURL,
		&order.CustomerID,
		&order.CustomerGroupID,
		&order.Firstname,
		&order.Lastname,
		&order.Email,
		&order.Telephone,
		&order.CustomField,
		&order.PaymentFirstname,
		&order.PaymentLastname,
		&order.PaymentCompany,
		&order.PaymentAddress1,
		&order.PaymentAddress2,
		&order.PaymentCity,
		&order.PaymentPostcode,
		&order.PaymentCountry,
		&order.PaymentCountryID,
		&order.PaymentZone,
		&order.PaymentZoneID,
		&order.PaymentAddressFormat,
		&order.PaymentCustomField,
		&order.PaymentMethod,
		&order.PaymentCode,
		&order.ShippingFirstname,
		&order.ShippingLastname,
		&order.ShippingCompany,
		&order.ShippingAddress1,
		&order.ShippingAddress2,
		&order.ShippingCity,
		&order.ShippingPostcode,
		&order.ShippingCountry,
		&order.ShippingCountryID,
		&order.ShippingZone,
		&order.ShippingZoneID,
		&order.ShippingAddressFormat,
		&order.ShippingCustomField,
		&order.ShippingMethod,
		&order.ShippingCode,
		&order.Comment,
		&order.Total,
		&order.OrderStatusID,
		&order.AffiliateID,
		&order.Commission,
		&order.MarketingID,
		&order.Tracking,
		&order.LanguageID,
		&order.CurrencyID,
		&order.CurrencyCode,
		&order.CurrencyValue,
		&order.IP,
		&order.ForwardedIP,
		&order.UserAgent,
		&order.AcceptLanguage,
		&order.DateAdded,
		&order.DateModified,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // no order found
		}
		return nil, err
	}
	return &order, nil
}

// OrderSearchStatus returns a list of order IDs matching the given status and created after the specified time.
func (s *MySql) OrderSearchStatus(statusId int64, from time.Time) ([]int64, error) {
	stmt, err := s.stmtSelectOrderStatus()
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(statusId, from)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var orderIds []int64
	for rows.Next() {
		var orderId int64
		if err = rows.Scan(&orderId); err != nil {
			return nil, err
		}
		orderIds = append(orderIds, orderId)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orderIds, nil
}

// OrderProducts returns all product line items for the given order ID.
func (s *MySql) OrderProducts(orderId int64) ([]*entity.ProductOrder, error) {
	stmt, err := s.stmtSelectOrderProducts()
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(orderId)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var products []*entity.ProductOrder
	for rows.Next() {
		var product entity.ProductOrder
		if err = rows.Scan(
			&product.DiscountAmount,
			&product.DiscountType,
			&product.Ean,
			&product.Isbn,
			&product.Jan,
			&product.Location,
			&product.Model,
			&product.Mpn,
			&product.Name,
			&product.OrderId,
			&product.Price,
			&product.ProductId,
			&product.Quantity,
			&product.Reward,
			&product.Sku,
			&product.Tax,
			&product.Total,
			&product.Upc,
			&product.Weight,
			&product.ProductUid,
		); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// OrderTotals returns all total line items (subtotal, tax, shipping, etc.) for the given order ID.
func (s *MySql) OrderTotals(orderId int64) ([]*entity.OrderTotal, error) {
	stmt, err := s.stmtSelectOrderTotals()
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(orderId)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var totals []*entity.OrderTotal
	for rows.Next() {
		var total entity.OrderTotal
		if err = rows.Scan(
			&total.Code,
			&total.Title,
			&total.Value,
		); err != nil {
			return nil, err
		}
		totals = append(totals, &total)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return totals, nil
}

// UpdateOrderStatus updates the current order status only if 'statusId' is less than the current status_id value
func (s *MySql) UpdateOrderStatus(orderId int64, statusId int, comment string) error {
	stmt, err := s.stmtUpdateOrderStatus()
	if err != nil {
		return err
	}

	res, err := stmt.Exec(
		statusId,
		time.Now(),
		orderId,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if rows > 0 {
		// add order history record
		rec := map[string]interface{}{
			"order_id":        orderId,
			"order_status_id": statusId,
			"notify":          0,
			"comment":         comment,
			"date_added":      time.Now(),
		}
		_, err = s.insert("order_history", rec)
		if err != nil {
			return fmt.Errorf("insert order history: %w", err)
		}
	}

	return nil
}

// UpdateCurrencyValue sets the exchange rate value for the given currency code.
func (s *MySql) UpdateCurrencyValue(currencyCode string, value float64) error {
	stmt, err := s.stmtUpdateCurrencyValue()
	if err != nil {
		return err
	}

	_, err = stmt.Exec(value, currencyCode)
	if err != nil {
		return err
	}

	return nil
}
