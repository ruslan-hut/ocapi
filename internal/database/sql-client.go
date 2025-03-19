package database

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"ocapi/entity"
	"ocapi/internal/config"
	"strings"
	"time"
)

type SQLDB struct {
	ctx    context.Context
	db     *sql.DB
	prefix string
}

func NewSQLClient(conf *config.Config) (*SQLDB, error) {
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

	return &SQLDB{
		db:     db,
		ctx:    context.Background(),
		prefix: conf.SQL.Prefix,
	}, nil
}

func (s *SQLDB) Close() {
	_ = s.db.Close()
}

func (s *SQLDB) ProductSearch(model string) ([]*entity.Product, error) {
	query := fmt.Sprintf("SELECT * FROM %sproduct WHERE model=?", s.prefix)
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
		if err = rows.Scan(&product); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &product)
	}
	return products, nil
}

func (s *SQLDB) AddProducts(products []entity.Product) error {
	for _, product := range products {
		err := s.addProduct(product)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLDB) addProduct(product entity.Product) error {
	// Check if the product already exists in the database by its code (uuid)
	var productId int64
	err := s.db.QueryRowContext(s.ctx, "SELECT product_id FROM ?product WHERE LCASE(code) = ?", s.prefix, product.UUID).Scan(&productId)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to query product: %v", err)
	}

	var nowDate = time.Now().Format("2006-01-02 15:04:05")

	manufacturerId, err := s.getManufacturerId(product.Manufacturer)
	if err != nil {
		manufacturerId = 0
	}

	var stockStatusId = 5
	if product.Stock > 0 {
		stockStatusId = 7
	}

	var status = 0
	if product.Active {
		status = 1
	}
	// If the product exists, update it
	if err == nil {
		// Update the product details in the database
		_, err := s.db.ExecContext(s.ctx, `
			UPDATE ?product SET
				model = ?, 
				sku = ?, 
				quantity = ?, 
				price = ?, 
				weight = ?, 
				stock_status_id = ?, 
				manufacturer_id = ?, 
				length = ?, 
				width = ?, 
				height = ?, 
				status = ?, 
				date_modified = ? 
			    WHERE product_id = ?`,
			s.prefix,
			product.UUID,
			product.Sku,
			product.Quantity,
			product.Price,
			0,
			stockStatusId,
			manufacturerId,
			product.Length,
			product.Width,
			product.Height,
			status,
			nowDate,
			productId)
		if err != nil {
			return fmt.Errorf("failed to update product: %v", err)
		}

		// Update the product descriptions in different languages
		if err := s.updateProductDescription(productId, 1, product.UUID, product.Description); err != nil {
			return err
		}
		if err := s.updateProductDescription(productId, 2, product.UUID, product.Description); err != nil {
			return err
		}
		if err := s.updateProductDescription(productId, 3, product.UUID, product.Description); err != nil {
			return err
		}
	} else {
		// Insert the product if it does not exist
		dateAvailable := time.Now().Add(-3 * 24 * time.Hour).Format("2006-01-02")

		// Insert product into the product table
		res, err := s.db.ExecContext(s.ctx, `
			INSERT INTO product (
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
			VALUES (?, ?, ?, ?, '1', '1', ?, ?, ?, '1', ?, '0', 0, '1', ?, ?, ?, '1', ?, '9', '0', ?, ?)`,
			product.Code, product.UUID, product.Sku, product.Quantity, product.StockStatusId, dateAvailable,
			manufacturerId, product.Price, product.Length, product.Width, product.Height, product.Status,
			nowDate, nowDate)
		if err != nil {
			return fmt.Errorf("failed to insert product: %v", err)
		}

		// Get the last inserted product_id
		productId, err = res.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get last insert id: %v", err)
		}

		// Insert product description in different languages
		if err := s.insertProductDescription(productId, 1, product.UUID, product.Description); err != nil {
			return err
		}
		if err := s.insertProductDescription(productId, 2, product.UUID, product.Description); err != nil {
			return err
		}
		if err := s.insertProductDescription(productId, 3, product.UUID, product.Description); err != nil {
			return err
		}
	}

	// Insert product to category (if applicable)
	if err := s.addProductToCategory(productId, product); err != nil {
		return err
	}

	// Insert SEO URL
	seoURL := s.TransLit(product.UUID)
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

	err = s.disActivateProducts(nowDate)
	if err != nil {
		return fmt.Errorf("failed to disactivate products: %v", err)
	}
	return nil
}

func (s *SQLDB) insertProductDescription(productId int64, languageId int, name, description string) error {
	_, err := s.db.ExecContext(s.ctx, `
		INSERT INTO ?product_description (
		    product_id, 
		    language_id, 
		    name, 
		    description, 
		    tag,
		    meta_title,
		    meta_description,
		    meta_keyword)
		VALUES (?, ?, ?, ?, '', ?, '', '')`,
		s.prefix,
		productId, languageId, name, description, name)
	return err
}

// Helper function to update product descriptions
func (s *SQLDB) updateProductDescription(productId int64, languageId int, name, description string) error {
	_, err := s.db.ExecContext(s.ctx, `
		UPDATE ?product_description 
		SET name = ?, description = ?, meta_title = ? 
		WHERE product_id = ? AND language_id = ?`,
		s.prefix,
		name, description, name, productId, languageId)
	return err
}

// Helper function to add product to category
func (s *SQLDB) addProductToCategory(productId int64, product entity.Product) error {
	if product.CategoryUUID != "" {
		var categoryId int
		err := s.db.QueryRowContext(s.ctx, "SELECT category_id FROM ?category WHERE category_uid = ?", s.prefix, product.CategoryUUID).Scan(&categoryId)
		if err != nil {
			return fmt.Errorf("failed to find category: %v", err)
		}
		if categoryId != 0 {
			_, err := s.db.ExecContext(s.ctx, "INSERT INTO ?product_to_category (product_id, category_id) VALUES (?, ?)", s.prefix, productId, categoryId)
			return err
		}
	}
	return nil
}

func (s *SQLDB) getManufacturerId(manufacturer string) (int64, error) {
	// Convert manufacturer name to lowercase and prepare it for the query
	manufacturerLower := strings.ToLower(manufacturer)

	// Query to find the manufacturer
	var manufacturerId int64
	query := `SELECT manufacturer_id FROM ?manufacturer WHERE LOWER(name) = ?`
	err := s.db.QueryRowContext(s.ctx, query, s.prefix, manufacturerLower).Scan(&manufacturerId)

	// If manufacturer exists, return the manufacturer ID
	if err == nil {
		return manufacturerId, nil
	}

	// If manufacturer does not exist (err is sql.ErrNoRows), insert the manufacturer
	if err == sql.ErrNoRows {
		// Insert new manufacturer
		insertQuery := `INSERT INTO ?manufacturer (name, sort_order) VALUES (?, 0)`
		result, err := s.db.ExecContext(s.ctx, insertQuery, s.prefix, manufacturer)
		if err != nil {
			return 0, fmt.Errorf("failed to insert manufacturer: %w", err)
		}

		// Get the last inserted ID
		manufacturerId, err = result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failed to get last insert ID: %w", err)
		}

		// Insert the manufacturer-to-store relationship
		insertStoreQuery := `INSERT INTO ?manufacturer_to_store (manufacturer_id, store_id) VALUES (?, 0)`
		_, err = s.db.ExecContext(s.ctx, insertStoreQuery, s.prefix, manufacturerId)
		if err != nil {
			return 0, fmt.Errorf("failed to insert manufacturer_to_store: %w", err)
		}

		// Generate SEO URL and insert into seo_url table
		seoUrl := s.TransLit(manufacturer)
		seoUrl = s.MetaURL(seoUrl)
		seoUrl = strings.ToLower(seoUrl)

		insertSeoQuery := `INSERT INTO ?seo_url (store_id, language_id, query, keyword) VALUES (0, 1, ?, ?)`
		_, err = s.db.ExecContext(s.ctx, insertSeoQuery, s.prefix, fmt.Sprintf("manufacturer_id=%d", manufacturerId), seoUrl)
		if err != nil {
			return 0, fmt.Errorf("failed to insert seo_url: %w", err)
		}

		// Optionally clear cache if applicable
		// s.cache.Delete("manufacturer") // Assuming cache handling is implemented

		// Return the new manufacturer ID
		return manufacturerId, nil
	}

	// If it's another type of error, return it
	return 0, fmt.Errorf("failed to retrieve or insert manufacturer: %w", err)
}

// Placeholder for the TransLit function
func (s *SQLDB) TransLit(input string) string {
	// Transliteration logic here
	return input
}

// Placeholder for the MetaURL function
func (s *SQLDB) MetaURL(input string) string {
	// Meta URL logic here
	return input
}

func (s *SQLDB) disActivateProducts(now string) error {
	_, err := s.db.ExecContext(s.ctx, `
		UPDATE ?product SET status = 0 WHERE date_modified < ?`,
		s.prefix, now)
	return err
}
