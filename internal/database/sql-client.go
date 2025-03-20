package database

import (
	"database/sql"
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
		return nil, fmt.Errorf("sql connect error: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &MySql{
		db:     db,
		prefix: conf.SQL.Prefix,
	}, nil
}

func (s *MySql) Close() {
	_ = s.db.Close()
}

func (s *MySql) ProductSearch(model string) ([]*entity.Product, error) {
	query := fmt.Sprintf(
		`SELECT 
					product_id,
					model,
					status,
					stock_status_id,
					quantity,
					price,
					date_modified 
				FROM %sproduct 
				WHERE model=?`,
		s.prefix,
	)
	rows, err := s.db.Query(query, model)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
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
			&product.Status,
			&product.StockStatusId,
			&product.Quantity,
			&product.Price,
			&product.DateModified,
		); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &product)
	}
	return products, nil
}

func (s *MySql) SaveProducts(productsData []*entity.ProductData) error {
	var err error
	var products []*entity.Product
	for _, productData := range productsData {
		products, err = s.ProductSearch(productData.UID)
		if err != nil {
			return fmt.Errorf("product search failed: %v", err)
		}

		if products == nil {
			err = s.addProduct(productData)
		} else {
			err = s.updateProduct(productData)
		}

		if err != nil {
			return fmt.Errorf("product %s failed: %v", productData.UID, err)
		}
	}
	return nil
}

func (s *MySql) SaveProductsDescription(productsDescData []*entity.ProductDescription) error {
	for _, productDescData := range productsDescData {
		productId, err := s.getProductByUID(productDescData.ProductUid)
		if err != nil {
			return fmt.Errorf("product search failed: %v", err)
		}

		err = s.upsertProductDescription(productId, productDescData)

		if err != nil {
			return fmt.Errorf("productDesc %s failed: %v", productDescData.ProductUid, err)
		}
	}
	return nil
}

func (s *MySql) SaveCategories(categoriesData []*entity.CategoryData) error {
	for _, categoryData := range categoriesData {
		categoryId, err := s.getCategoryByUID(categoryData.CategoryUID)
		if err != nil {
			return fmt.Errorf("category search failed: %v", err)
		}
		category := entity.CategoryFromCategoryData(categoryData)
		category.CategoryId = categoryId

		err = s.updateCategory(category)

		if err != nil {
			return fmt.Errorf("category %s failed: %v", categoryData.CategoryUID, err)
		}
	}
	return nil
}

func (s *MySql) SaveCategoriesDescription(categoriesDescData []*entity.CategoryDescriptionData) error {
	for _, categoryDescData := range categoriesDescData {
		categoryId, err := s.getCategoryByUID(categoryDescData.CategoryUid)
		if err != nil {
			return fmt.Errorf("category search failed: %v", err)
		}
		category := entity.CategoryDescriptionFromCategoryDescriptionData(categoryDescData)
		category.CategoryId = categoryId
		err = s.upsertCategoryDescription(category)

		if err != nil {
			return fmt.Errorf("save category description %s: %v", categoryDescData.CategoryUid, err)
		}
	}
	return nil
}

func (s *MySql) updateProduct(product *entity.ProductData) error {
	query := fmt.Sprintf(
		`UPDATE %sproduct SET
				sku = ?, 
				quantity = ?, 
				price = ?, 
				manufacturer_id = ?, 
				status = ?, 
				date_modified = ? 
			    WHERE product_id = ?`,
		s.prefix,
	)

	var status = 0
	if product.Active {
		status = 1
	}

	res, err := s.db.Exec(query,
		product.Article,
		product.Quantity,
		product.Price,
		product.ManufacturerUID,
		status,
		time.Now(),
		product.UID)
	if err != nil {
		return fmt.Errorf("failed to update product: %v", err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("product %s not found", product.UID)
	}
	return nil
}

func (s *MySql) addProduct(productData *entity.ProductData) error {

	product := entity.ProductFromProductData(productData)

	query := fmt.Sprintf(
		`INSERT INTO %sproduct (
				model,
				sku,
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
				date_added,
				date_modified)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.prefix)
	// Insert product into the product table
	res, err := s.db.Exec(query,
		product.Model,
		product.Sku,
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
		product.DateAdded,
		product.DateModified)

	if err != nil {
		return fmt.Errorf("failed to insert product: %v", err)
	}

	// Get the last inserted product_id
	productId, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %v", err)
	}

	if err = s.addProductToStore(productId); err != nil {
		return err
	}

	if err = s.addProductToCategory(productData); err != nil {
		return err
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
		return fmt.Errorf("failed addProductToStore: %v", err)
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
		return fmt.Errorf("failed addProductToLayout: %v", err)
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
		return nil, fmt.Errorf("query failed: %w", err)
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
			return nil, fmt.Errorf("failed to scan ProductDescription: %w", err)
		}
		productsDesc = append(productsDesc, &prodDesc)
	}

	return productsDesc, nil
}

func (s *MySql) upsertProductDescription(productId int64, productDescription *entity.ProductDescription) error {
	query := fmt.Sprintf(
		`UPDATE %sproduct_description SET
					name = ?,
					description = ?
			    WHERE product_id = ? AND language_id = ?`,
		s.prefix,
	)
	res, err := s.db.Exec(query,
		productDescription.Name,
		productDescription.Description,
		productId,
		productDescription.LanguageId)
	if err != nil {
		return fmt.Errorf("update product description: %v", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		query = fmt.Sprintf(
			`INSERT INTO %sproduct_description (
				product_id,
				language_id,
                name,
                description,
                meta_title,
                meta_description,
                meta_keyword,
                tag)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			s.prefix)

		_, err = s.db.Exec(query,
			productId,
			productDescription.LanguageId,
			productDescription.Name,
			productDescription.Description,
			productDescription.Name,
			"",
			"",
			"")
		if err != nil {
			return fmt.Errorf("insert product description: %v", err)
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
		return 0, fmt.Errorf("query failed: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var productId int64
		if err = rows.Scan(
			productId,
		); err != nil {
			return 0, err
		}
		return productId, nil
	}

	return 0, fmt.Errorf("no product found with uid: %s", uid)
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
		return 0, fmt.Errorf("query failed: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var categoryId int64
		if err = rows.Scan(
			categoryId,
		); err != nil {
			return 0, err
		}
		return categoryId, nil
	}

	query = fmt.Sprintf(`INSERT INTO %scategory (category_uid, date_added) VALUES (?,?)`, s.prefix)
	res, err := s.db.Exec(query, uid, time.Now())
	if err != nil {
		return 0, fmt.Errorf("insert category: %w", err)
	}

	categoryId, _ := res.LastInsertId()

	return categoryId, nil
}

func (s *MySql) updateCategory(category *entity.Category) error {
	query := fmt.Sprintf(
		`UPDATE %scategory SET
                        parent_id,
                        parent_uid,
                        top,
                        column,
                        sort_order,
                        status,
                        date_modified
			    WHERE category_id = ?`,
		s.prefix,
	)
	_, err := s.db.Query(query,
		category.ParentId,
		category.ParentUID,
		category.Top,
		category.Column,
		category.SortOrder,
		category.Status,
		category.DateModified,
		category.CategoryId)
	if err != nil {
		return fmt.Errorf("update category: %v", err)
	}

	return nil
}

func (s *MySql) upsertCategoryDescription(categoryDesc *entity.CategoryDescription) error {
	query := fmt.Sprintf(
		`UPDATE %scategory_description SET
                        name,
                        description
			    WHERE category_id = ? AND language_id = ?`,
		s.prefix,
	)
	res, err := s.db.Exec(query,
		categoryDesc.Name,
		categoryDesc.Description,
		categoryDesc.CategoryId,
		categoryDesc.LanguageId)
	if err != nil {
		return fmt.Errorf("update category description: %v", err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		query = fmt.Sprintf(
			`INSERT INTO %scategory_description (
                       category_id,
                       language_id,
                       name,
                       description)
			VALUES (?, ?, ?, ?)`,
			s.prefix)

		_, err = s.db.Exec(query,
			categoryDesc.CategoryId,
			categoryDesc.LanguageId,
			categoryDesc.Name,
			categoryDesc.Description)

		if err != nil {
			return fmt.Errorf("insert category description: %v", err)
		}
	}

	return nil
}

// Helper function to add product to category
func (s *MySql) addProductToCategory(product *entity.ProductData) error {
	query := fmt.Sprintf(
		`INSERT INTO %sproduct_to_category (
                        product_id,
                        category_id)
			VALUES (?, ?)`,
		s.prefix)

	_, err := s.db.Exec(query,
		product.UID,
		product.CategoryUID)

	if err != nil {
		return fmt.Errorf("product to category insert: %v", err)
	}
	return nil
}

func (s *MySql) addParentUIDToCategory() error {
	// Check if the column exists
	query := fmt.Sprintf(`SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = '%scategory' AND COLUMN_NAME = 'parent_uid'`,
		s.prefix)
	err := s.db.QueryRow(query).Scan()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Column does not exist, so add it
			alterQuery := fmt.Sprintf(`ALTER TABLE %scategory ADD COLUMN parent_uid VARCHAR(64) NULL`, s.prefix)
			_, err = s.db.Exec(alterQuery)
			if err != nil {
				return fmt.Errorf("failed to add parent_uid column: %w", err)
			}
			fmt.Println("parent_uid column added successfully.")
		} else {
			return fmt.Errorf("error checking column existence: %w", err)
		}
	} else {
		fmt.Println("parent_uid column already exists.")
	}

	return nil
}

func (s *MySql) addCategoryUIDToCategory() error {
	// Check if the column exists
	query := fmt.Sprintf(`SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = '%scategory' AND COLUMN_NAME = 'category_uid'`,
		s.prefix)
	err := s.db.QueryRow(query).Scan()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Column does not exist, so add it
			alterQuery := fmt.Sprintf(`ALTER TABLE %scategory ADD COLUMN category_uid VARCHAR(64) NULL`, s.prefix)
			_, err = s.db.Exec(alterQuery)
			if err != nil {
				return fmt.Errorf("failed to add category_uid column: %w", err)
			}
			fmt.Println("category_uid column added successfully.")
		} else {
			return fmt.Errorf("error checking column existence: %w", err)
		}
	} else {
		fmt.Println("category_uid column already exists.")
	}

	return nil
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
