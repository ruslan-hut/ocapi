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
	query := fmt.Sprintf(`SELECT product_id FROM %sproduct WHERE model=? LIMIT 1`, s.prefix)
	return s.prepareStmt("selectProductId", query)
}

func (s *MySql) stmtCategoryId() (*sql.Stmt, error) {
	query := fmt.Sprintf(`SELECT category_id FROM %scategory WHERE category_uid=? LIMIT 1`, s.prefix)
	return s.prepareStmt("selectCategoryId", query)
}

func (s *MySql) stmtCategoryDescription() (*sql.Stmt, error) {
	query := fmt.Sprintf(
		`SELECT
					name,
					description,
					meta_title,
					meta_description,
					meta_keyword
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
	return s.prepareStmt("updateProduct", query)
}

func (s *MySql) stmtUpdateProductImage() (*sql.Stmt, error) {
	query := fmt.Sprintf(`UPDATE %sproduct SET image = ? WHERE model = ?`, s.prefix)
	return s.prepareStmt("updateProductImage", query)
}
