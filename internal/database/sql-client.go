package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"ocapi/entity"
	"ocapi/internal/config"
	"time"
)

type MySql struct {
	ctx    context.Context
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
		ctx:    context.Background(),
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
			err = s.addProduct(entity.ProductFromProductData(productData))
			if err != nil {
				return fmt.Errorf("product %s failed: %v", productData.UID, err)
			}
			err = s.addProductToCategory(productData)
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
	var err error
	var products []*entity.Product
	var productsDesc []*entity.ProductDescription
	for _, productDescData := range productsDescData {
		products, err = s.ProductSearch(productDescData.ProductId)
		if err != nil {
			return fmt.Errorf("product search failed: %v", err)
		}

		productsDesc, err = s.getProductDescription(productDescData)
		if err != nil {
			return fmt.Errorf("productDesc search failed: %v", err)
		}

		if products != nil {
			if productsDesc == nil {
				err = s.addProductDescription(productDescData)
			} else {
				err = s.updateProductDescription(productDescData)
			}
		} else {
			return fmt.Errorf("productDesc %s no such product", productDescData.ProductId)
		}

		if err != nil {
			return fmt.Errorf("productDesc %s failed: %v", productDescData.ProductId, err)
		}
	}
	return nil
}

func (s *MySql) SaveCategories(categoriesData []*entity.CategoryData) error {
	var err error
	var categories []*entity.Category
	for _, categoryData := range categoriesData {
		categories, err = s.getCategoryByUID(categoryData.CategoryUID)
		if err != nil {
			return fmt.Errorf("product search failed: %v", err)
		}

		if categories == nil {
			err = s.addCategory(entity.CategoryFromCategoryData(categoryData))
		} else {
			err = s.updateCategory(entity.CategoryFromCategoryData(categoryData))
		}

		if err != nil {
			return fmt.Errorf("category %s failed: %v", categoryData.CategoryUID, err)
		}
	}
	return nil
}

func (s *MySql) SaveCategoriesDescription(categoriesDescData []*entity.CategoryDescriptionData) error {
	var err error
	for _, categoryDescData := range categoriesDescData {
		err = s.upsertCategoryDescription(entity.CategoryDescriptionFromCategoryDescriptionData(categoryDescData))

		if err != nil {
			return fmt.Errorf("categoryDescription %s failed: %v", categoryDescData.CategoryId, err)
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

	_, err := s.db.Query(query,
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

	// Update the product descriptions in different languages
	//if err = s.updateProductDescription(product.Id, 1, product.Model, product.Description); err != nil {
	//	return err
	//}
	//if err = s.updateProductDescription(product.Id, 2, product.Model, product.Description); err != nil {
	//	return err
	//}
	//if err = s.updateProductDescription(product.Id, 3, product.Model, product.Description); err != nil {
	//	return err
	//}
	return nil
}

func (s *MySql) addProduct(product *entity.Product) error {

	//manufacturerId, err := s.getManufacturerId(product.Manufacturer)
	//if err != nil {
	//	manufacturerId = 0
	//}

	//var stockStatusId = 5
	//if product.Stock > 0 {
	//	stockStatusId = 7
	//}

	//var status = 0
	//if product.Active {
	//	status = 1
	//}

	query := fmt.Sprintf(
		`INSERT INTO %sproduct (
                code,
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
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.prefix)
	// Insert product into the product table
	res, err := s.db.Exec(query,
		product.Model,
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

	if err = s.addProductToLayout(productId); err != nil {
		return err
	}

	// Insert SEO URL
	seoURL := s.TransLit(product.Model)
	seoURL = s.MetaURL(seoURL)
	seoURL = fmt.Sprintf("%s", seoURL)
	_, err = s.db.ExecContext(s.ctx, `
		INSERT INTO seo_url (store_id, language_id, query, keyword) VALUES 
		('0', 1, ?, ?), ('0', 2, ?, ?), ('0', 3, ?, ?)`,
		fmt.Sprintf("product_id=%d", productId), seoURL,
		fmt.Sprintf("product_id=%d", productId), seoURL,
		fmt.Sprintf("product_id=%d", productId), seoURL)
	if err != nil {
		return fmt.Errorf("failed to insert SEO URL: %v", err)
	}

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
				WHERE product_id=? AND language_id = ?`,
		s.prefix,
	)
	rows, err := s.db.Query(query,
		productDesc.Name,
		productDesc.Description,
		productDesc.ProductId,
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

func (s *MySql) addProductDescription(productDescription *entity.ProductDescription) error {
	query := fmt.Sprintf(
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

	_, err := s.db.Exec(query,
		productDescription.ProductId,
		productDescription.LanguageId,
		productDescription.Name,
		productDescription.Description,
		productDescription.Name,
		"",
		"",
		"")

	if err != nil {
		return fmt.Errorf("failed AddProductDescription: %v", err)
	}

	return nil
}

func (s *MySql) updateProductDescription(productDescription *entity.ProductDescription) error {
	query := fmt.Sprintf(
		`UPDATE %sproduct_description SET
					name = ?,
					description = ?
			    WHERE product_id = ? AND language_id = ?`,
		s.prefix,
	)
	_, err := s.db.Query(query,
		productDescription.Name,
		productDescription.Description,
		productDescription.ProductId,
		productDescription.LanguageId)
	if err != nil {
		return fmt.Errorf("failed to update product: %v", err)
	}

	return nil
}

func (s *MySql) getCategoryByUID(uid int64) ([]*entity.Category, error) {
	query := fmt.Sprintf(
		`SELECT
					category_id,
					category_uid,
					parent_id,
					top,
					SortOrder,
					Status,
					date_modified
				FROM %scategory 
				WHERE category_uid=?`,
		s.prefix,
	)
	rows, err := s.db.Query(query, uid)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var categories []*entity.Category
	for rows.Next() {
		var category entity.Category
		if err = rows.Scan(
			&category.CategoryId,
			&category.CategoryUID,
			&category.ParentId,
			&category.Top,
			&category.SortOrder,
			&category.Status,
			&category.DateModified,
		); err != nil {
			categoryData := entity.DefaultCategoryData(uid)
			err = s.addCategory(entity.CategoryFromCategoryData(categoryData))
			return s.getCategoryByUID(uid)
		}
		categories = append(categories, &category)
	}

	return categories, nil
}

func (s *MySql) addCategory(category *entity.Category) error {
	query := fmt.Sprintf(
		`INSERT INTO %scategory (
                        category_uid,
                        parent_id,
                        parent_uid,
                        top,
                        column,
                        sort_order,
                        status,
                        date_added,
                        date_modified)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.prefix)

	_, err := s.db.Exec(query,
		category.CategoryUID,
		category.ParentId,
		category.ParentUID,
		category.Top,
		category.Column,
		category.SortOrder,
		category.Status,
		category.DateAdded,
		category.DateModified)

	if err != nil {
		return fmt.Errorf("failed AddCategory: %v", err)
	}

	return nil
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
		return fmt.Errorf("failed to UpdateCategory: %v", err)
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
		return fmt.Errorf("failed to UpdateCategory: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected <= 0 {
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
			return fmt.Errorf("failed AddCategory: %v", err)
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
		return fmt.Errorf("failed addProductToCategory: %v", err)
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
			alterQuery := fmt.Sprintf(`ALTER TABLE %scategory ADD COLUMN parent_uid VARCHAR(36) NULL`, s.prefix)
			_, err := s.db.Exec(alterQuery)
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
			alterQuery := fmt.Sprintf(`ALTER TABLE %scategory ADD COLUMN category_uid VARCHAR(36) NULL`, s.prefix)
			_, err := s.db.Exec(alterQuery)
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
	_, err := s.db.ExecContext(s.ctx, `
		UPDATE ?product SET status = 0 WHERE date_modified < ?`,
		s.prefix, now)
	return err
}
