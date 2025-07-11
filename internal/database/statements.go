package database

import (
	"database/sql"
	"fmt"
)

func (s *MySql) prepareStmt(name, query string) (*sql.Stmt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// если уже есть — возвращаем
	if stmt, ok := s.statements[name]; ok {
		return stmt, nil
	}

	// подготавливаем новый
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("prepare statement [%s]: %w", name, err)
	}

	s.statements[name] = stmt
	return stmt, nil
}

func (s *MySql) closeStmt() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for name, stmt := range s.statements {
		_ = stmt.Close()
		delete(s.statements, name)
	}
}

func (s *MySql) stmtSelectProductId() (*sql.Stmt, error) {
	query := fmt.Sprintf(`SELECT product_id FROM %sproduct WHERE product_uid=? LIMIT 1`, s.prefix)
	return s.prepareStmt("selectProductId", query)
}

func (s *MySql) stmtCategoryId() (*sql.Stmt, error) {
	query := fmt.Sprintf(`SELECT category_id FROM %scategory WHERE category_uid=? LIMIT 1`, s.prefix)
	return s.prepareStmt("selectCategoryId", query)
}

func (s *MySql) stmtSelectAttributeId() (*sql.Stmt, error) {
	query := fmt.Sprintf(`SELECT attribute_id FROM %sattribute WHERE attribute_uid=? LIMIT 1`, s.prefix)
	return s.prepareStmt("selectAttributeId", query)
}

func (s *MySql) stmtCategoryDescription() (*sql.Stmt, error) {
	query := fmt.Sprintf(
		`SELECT
					name,
					description
				FROM %scategory_description 
				WHERE category_id=? AND language_id=?
				LIMIT 1`,
		s.prefix,
	)
	return s.prepareStmt("selectCategoryDescription", query)
}

func (s *MySql) stmtManufacturerId() (*sql.Stmt, error) {
	query := fmt.Sprintf(`SELECT manufacturer_id FROM %smanufacturer WHERE name=? LIMIT 1`, s.prefix)
	return s.prepareStmt("selectManufacturerId", query)
}

func (s *MySql) stmtSelectOrder() (*sql.Stmt, error) {
	query := fmt.Sprintf(
		`SELECT
            order_id,
            invoice_no,
            invoice_prefix,
            store_id,
            store_name,
            store_url,
            customer_id,
            customer_group_id,
            firstname,
            lastname,
            email,
            telephone,
            custom_field,
            payment_firstname,
            payment_lastname,
            payment_company,
            payment_address_1,
            payment_address_2,
            payment_city,
            payment_postcode,
            payment_country,
            payment_country_id,
            payment_zone,
            payment_zone_id,
            payment_address_format,
            payment_custom_field,
            payment_method,
            payment_code,
            shipping_firstname,
            shipping_lastname,
            shipping_company,
            shipping_address_1,
            shipping_address_2,
            shipping_city,
            shipping_postcode,
            shipping_country,
            shipping_country_id,
            shipping_zone,
            shipping_zone_id,
            shipping_address_format,
            shipping_custom_field,
            shipping_method,
            shipping_code,
            comment,
            total,
            order_status_id,
            affiliate_id,
            commission,
            marketing_id,
            tracking,
            language_id,
            currency_id,
            currency_code,
            currency_value,
            ip,
            forwarded_ip,
            user_agent,
            accept_language,
            date_added,
            date_modified
         FROM %sorder
         WHERE order_id = ?
         LIMIT 1`,
		s.prefix,
	)
	return s.prepareStmt("selectOrder", query)
}

func (s *MySql) stmtSelectOrderStatus() (*sql.Stmt, error) {
	query := fmt.Sprintf(
		`SELECT
			order_id
		 FROM %sorder
		 WHERE order_status_id = ?
		 LIMIT 100`,
		s.prefix,
	)
	return s.prepareStmt("selectOrderStatus", query)
}

func (s *MySql) stmtSelectOrderProducts() (*sql.Stmt, error) {
	query := fmt.Sprintf(
		`SELECT
			op.discount_amount,
			op.discount_type,
			op.ean,
			op.isbn,
			op.jan,
			op.location,
			op.model,
			op.mpn,
			op.name,
			op.order_id,
			op.price,
			op.product_id,
			op.quantity,
			op.reward,
			op.sku,
			op.tax,
			op.total,
			op.upc,
			op.weight,
			p.product_uid
		 FROM %sorder_product op
		 JOIN %sproduct p ON op.product_id = p.product_id
		 WHERE op.order_id = ?`,
		s.prefix, s.prefix,
	)
	return s.prepareStmt("selectOrderProducts", query)
}

func (s *MySql) stmtSelectOrderTotals() (*sql.Stmt, error) {
	query := fmt.Sprintf(
		`SELECT
			op.code,
			op.title,
			op.value
		 FROM %sorder_total op
		 WHERE op.order_id = ?`,
		s.prefix,
	)
	return s.prepareStmt("selectOrderTotals", query)
}

func (s *MySql) stmtUpdateCategoryDescription() (*sql.Stmt, error) {
	query := fmt.Sprintf(
		`UPDATE %scategory_description
				SET
					name=?,
					description=?
				WHERE category_id=? AND language_id=?`,
		s.prefix,
	)
	return s.prepareStmt("updateCategoryDescription", query)
}

func (s *MySql) stmtUpdateCategoryName() (*sql.Stmt, error) {
	query := fmt.Sprintf(
		`UPDATE %scategory_description
				SET
					name=?
				WHERE category_id=? AND language_id=?`,
		s.prefix,
	)
	return s.prepareStmt("updateCategoryName", query)
}

func (s *MySql) stmtUpdateCategory() (*sql.Stmt, error) {
	query := fmt.Sprintf(
		`UPDATE %scategory
				SET
					parent_id=?,
					parent_uid=?,
					top=?,
					sort_order=?,
					status=?,
					date_modified=?
				WHERE category_id=?`,
		s.prefix,
	)
	return s.prepareStmt("updateCategory", query)
}

func (s *MySql) stmtUpdateProduct() (*sql.Stmt, error) {
	query := fmt.Sprintf(
		`UPDATE %sproduct SET
				model = ?, 
				quantity = ?, 
                stock_status_id = ?,
				price = ?, 
				manufacturer_id = ?, 
				status = ?, 
				weight = ?,
                weight_class_id = ?,
				date_modified = ?,
                batch_uid = ?
			    WHERE product_id = ?`,
		s.prefix,
	)
	return s.prepareStmt("updateProduct", query)
}

func (s *MySql) stmtUpdateProductImage() (*sql.Stmt, error) {
	query := fmt.Sprintf(`UPDATE %sproduct SET image = ? WHERE product_uid = ?`, s.prefix)
	return s.prepareStmt("updateProductImage", query)
}

func (s *MySql) stmtUpdateAttribute() (*sql.Stmt, error) {
	query := fmt.Sprintf(
		`UPDATE %sattribute SET
				attribute_uid = ?, 
				attribute_group_id = ?, 
                sort_order = ?
			    WHERE attribute_id = ?`,
		s.prefix,
	)
	return s.prepareStmt("updateAttribute", query)
}

func (s *MySql) stmtGetProductNotMainImage() (*sql.Stmt, error) {
	query := fmt.Sprintf(
		"SELECT product_image_id FROM %sproduct_image WHERE product_id=? AND file_uid=? LIMIT 1",
		s.prefix,
	)
	return s.prepareStmt("getProductNotMainImage", query)
}

func (s *MySql) stmtGetProductImages() (*sql.Stmt, error) {
	query := fmt.Sprintf(
		`SELECT product_image_id, file_uid
         FROM %sproduct_image
         WHERE product_id = ?`,
		s.prefix,
	)
	return s.prepareStmt("getProductImages", query)
}

func (s *MySql) stmtGetApiUsername() (*sql.Stmt, error) {
	query := fmt.Sprintf("SELECT username FROM %sapi WHERE `key`=? AND status=1 LIMIT 1",
		s.prefix,
	)
	return s.prepareStmt("getApiUsername", query)
}

func (s *MySql) stmtFindProductSpecial() (*sql.Stmt, error) {
	query := fmt.Sprintf(
		`SELECT product_special_id FROM %sproduct_special WHERE product_id=? AND customer_group_id=? LIMIT 1`,
		s.prefix,
	)
	return s.prepareStmt("findProductSpecial", query)
}

func (s *MySql) stmtUpdateProductSpecial() (*sql.Stmt, error) {
	query := fmt.Sprintf(
		`UPDATE %sproduct_special SET
				price = ?, 
				priority = ?, 
				date_start = ?,
				date_end = ?
			    WHERE product_id = ? AND customer_group_id = ?`,
		s.prefix,
	)
	return s.prepareStmt("updateProductSpecial", query)
}

func (s *MySql) stmtUpdateOrderStatus() (*sql.Stmt, error) {
	query := fmt.Sprintf(
		`UPDATE %sorder SET
				order_status_id = ?,
			    date_modified = ?
			    WHERE order_id = ? AND order_status_id < ?`,
		s.prefix,
	)
	return s.prepareStmt("updateOrderStatus", query)
}
