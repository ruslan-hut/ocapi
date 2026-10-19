package database

import (
	"database/sql"
	"fmt"
	"ocapi/entity"
	"strings"
)

// syncChunkSize limits the number of values in a single IN list or multi-row INSERT.
const syncChunkSize = 500

// zeroDateTime is the "no bound" value of product_special dates expected by OpenCart queries.
const zeroDateTime = "0000-00-00 00:00:00"

// migrateProductSpecialDates converts product_special.date_start/date_end from DATE to DATETIME,
// so specials can start and end at an exact time. Existing values become midnight.
func (s *MySql) migrateProductSpecialDates() error {
	var modify []string
	for _, column := range []string{"date_start", "date_end"} {
		var dataType string
		err := s.db.QueryRow(
			`SELECT DATA_TYPE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?`,
			s.prefix+"product_special", column).Scan(&dataType)
		if err != nil {
			return fmt.Errorf("checking product_special.%s type: %w", column, err)
		}
		if strings.EqualFold(dataType, "date") {
			modify = append(modify, fmt.Sprintf("MODIFY `%s` DATETIME NOT NULL DEFAULT '%s'", column, zeroDateTime))
		}
	}
	if len(modify) == 0 {
		return nil
	}
	query := fmt.Sprintf("ALTER TABLE `%sproduct_special` %s", s.prefix, strings.Join(modify, ", "))
	if _, err := s.db.Exec(query); err != nil {
		return fmt.Errorf("migrate product_special dates: %w", err)
	}
	return nil
}

// createCustomerSpecialTable creates the personal customer prices table read by the site.
func (s *MySql) createCustomerSpecialTable() error {
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%scustomer_product_special` ("+
		"`customer_id` INT NOT NULL, "+
		"`product_id` INT NOT NULL, "+
		"`price` DECIMAL(15,4) NOT NULL, "+
		"`date_start` DATETIME NULL, "+
		"`date_end` DATETIME NULL, "+
		"KEY `customer_product` (`customer_id`, `product_id`), "+
		"KEY `product_id` (`product_id`)"+
		") ENGINE=InnoDB", s.prefix)
	if _, err := s.db.Exec(query); err != nil {
		return fmt.Errorf("create customer_product_special: %w", err)
	}
	return nil
}

// SyncProductSpecials replaces special prices of the given customer groups for every product in the batch.
// Specials of other groups are kept. Unknown product UIDs are skipped and reported.
func (s *MySql) SyncProductSpecials(groups []int64, specials []*entity.SpecialSync) (*entity.SyncResult, error) {
	result := entity.NewSyncResult()

	productIds, err := s.productIdsByUid(specialProductUids(specials))
	if err != nil {
		return nil, fmt.Errorf("product search: %w", err)
	}

	var products []interface{}
	var rows [][]interface{}
	seen := make(map[string]bool)
	for _, special := range specials {
		productId, ok := productIds[special.ProductUid]
		if !seen[special.ProductUid] {
			seen[special.ProductUid] = true
			if ok {
				products = append(products, productId)
			} else {
				result.SkippedProducts = append(result.SkippedProducts, special.ProductUid)
			}
		}
		if !ok {
			continue
		}
		rows = append(rows, []interface{}{
			productId,
			special.GroupId,
			special.Priority,
			special.Price,
			dateOrZero(special.DateStart),
			dateOrZero(special.DateEnd),
		})
	}
	if len(products) == 0 {
		return result, nil
	}

	groupArgs := make([]interface{}, len(groups))
	for i, group := range groups {
		groupArgs[i] = group
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	for _, chunk := range chunks(products) {
		query := fmt.Sprintf("DELETE FROM `%sproduct_special` WHERE customer_group_id IN (%s) AND product_id IN (%s)",
			s.prefix, placeholders(len(groupArgs)), placeholders(len(chunk)))
		res, err := tx.Exec(query, append(append([]interface{}{}, groupArgs...), chunk...)...)
		if err != nil {
			return nil, fmt.Errorf("delete product specials: %w", err)
		}
		result.Deleted += rowsAffected(res)
	}

	written, err := s.insertRows(tx, "product_special",
		[]string{"product_id", "customer_group_id", "priority", "price", "date_start", "date_end"}, rows)
	if err != nil {
		return nil, fmt.Errorf("insert product specials: %w", err)
	}
	result.Written = written

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return result, nil
}

// SyncCustomerSpecials replaces personal prices of every customer in the batch.
// Customers missing on the site and unknown product UIDs are skipped and reported.
func (s *MySql) SyncCustomerSpecials(specials []*entity.CustomerSpecial) (*entity.SyncResult, error) {
	result := entity.NewSyncResult()

	var customers []interface{}
	seenCustomers := make(map[int64]bool)
	for _, special := range specials {
		if !seenCustomers[special.CustomerId] {
			seenCustomers[special.CustomerId] = true
			customers = append(customers, special.CustomerId)
		}
	}
	if len(customers) == 0 {
		return result, nil
	}

	existing, err := s.existingCustomers(customers)
	if err != nil {
		return nil, fmt.Errorf("customer search: %w", err)
	}
	for _, customerId := range customers {
		if !existing[customerId.(int64)] {
			result.SkippedCustomers = append(result.SkippedCustomers, customerId.(int64))
		}
	}

	var uids []string
	seenUids := make(map[string]bool)
	for _, special := range specials {
		if existing[special.CustomerId] && !seenUids[special.ProductUid] {
			seenUids[special.ProductUid] = true
			uids = append(uids, special.ProductUid)
		}
	}
	productIds, err := s.productIdsByUid(uids)
	if err != nil {
		return nil, fmt.Errorf("product search: %w", err)
	}
	for _, uid := range uids {
		if _, ok := productIds[uid]; !ok {
			result.SkippedProducts = append(result.SkippedProducts, uid)
		}
	}

	var rows [][]interface{}
	for _, special := range specials {
		productId, ok := productIds[special.ProductUid]
		if !ok || !existing[special.CustomerId] {
			continue
		}
		rows = append(rows, []interface{}{
			special.CustomerId,
			productId,
			special.Price,
			dateOrNull(special.DateStart),
			dateOrNull(special.DateEnd),
		})
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// rows of every customer in the batch are removed, including customers missing on the site
	if result.Deleted, err = s.deleteCustomerSpecials(tx, customers); err != nil {
		return nil, err
	}

	written, err := s.insertRows(tx, "customer_product_special",
		[]string{"customer_id", "product_id", "price", "date_start", "date_end"}, rows)
	if err != nil {
		return nil, fmt.Errorf("insert customer specials: %w", err)
	}
	result.Written = written

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return result, nil
}

// DeleteCustomerSpecials removes all personal prices of the given customers.
func (s *MySql) DeleteCustomerSpecials(customerIds []int64) (int64, error) {
	customers := make([]interface{}, len(customerIds))
	for i, customerId := range customerIds {
		customers[i] = customerId
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	deleted, err := s.deleteCustomerSpecials(tx, customers)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return deleted, nil
}

func (s *MySql) deleteCustomerSpecials(tx *sql.Tx, customers []interface{}) (int64, error) {
	var deleted int64
	for _, chunk := range chunks(customers) {
		query := fmt.Sprintf("DELETE FROM `%scustomer_product_special` WHERE customer_id IN (%s)",
			s.prefix, placeholders(len(chunk)))
		res, err := tx.Exec(query, chunk...)
		if err != nil {
			return 0, fmt.Errorf("delete customer specials: %w", err)
		}
		deleted += rowsAffected(res)
	}
	return deleted, nil
}

// insertRows writes rows with multi-row INSERT statements and returns the number of inserted rows.
func (s *MySql) insertRows(tx *sql.Tx, table string, columns []string, rows [][]interface{}) (int64, error) {
	rowPlaceholder := "(" + placeholders(len(columns)) + ")"
	var written int64
	for start := 0; start < len(rows); start += syncChunkSize {
		end := min(start+syncChunkSize, len(rows))
		values := make([]string, 0, end-start)
		args := make([]interface{}, 0, (end-start)*len(columns))
		for _, row := range rows[start:end] {
			values = append(values, rowPlaceholder)
			args = append(args, row...)
		}
		query := fmt.Sprintf("INSERT INTO `%s%s` (`%s`) VALUES %s",
			s.prefix, table, strings.Join(columns, "`, `"), strings.Join(values, ", "))
		res, err := tx.Exec(query, args...)
		if err != nil {
			return 0, err
		}
		written += rowsAffected(res)
	}
	return written, nil
}

// productIdsByUid resolves product UIDs to IDs in chunks; unknown UIDs are absent from the result.
func (s *MySql) productIdsByUid(uids []string) (map[string]int64, error) {
	ids := make(map[string]int64, len(uids))
	args := make([]interface{}, len(uids))
	for i, uid := range uids {
		args[i] = uid
	}
	for _, chunk := range chunks(args) {
		query := fmt.Sprintf("SELECT product_id, product_uid FROM `%sproduct` WHERE product_uid IN (%s)",
			s.prefix, placeholders(len(chunk)))
		rows, err := s.db.Query(query, chunk...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var productId int64
			var uid string
			if err = rows.Scan(&productId, &uid); err != nil {
				_ = rows.Close()
				return nil, err
			}
			if _, ok := ids[uid]; !ok {
				ids[uid] = productId
			}
		}
		err = rows.Err()
		_ = rows.Close()
		if err != nil {
			return nil, err
		}
	}
	return ids, nil
}

// existingCustomers returns the subset of customer IDs present in the customer table.
func (s *MySql) existingCustomers(customers []interface{}) (map[int64]bool, error) {
	existing := make(map[int64]bool, len(customers))
	for _, chunk := range chunks(customers) {
		query := fmt.Sprintf("SELECT customer_id FROM `%scustomer` WHERE customer_id IN (%s)",
			s.prefix, placeholders(len(chunk)))
		rows, err := s.db.Query(query, chunk...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var customerId int64
			if err = rows.Scan(&customerId); err != nil {
				_ = rows.Close()
				return nil, err
			}
			existing[customerId] = true
		}
		err = rows.Err()
		_ = rows.Close()
		if err != nil {
			return nil, err
		}
	}
	return existing, nil
}

func specialProductUids(specials []*entity.SpecialSync) []string {
	var uids []string
	seen := make(map[string]bool)
	for _, special := range specials {
		if !seen[special.ProductUid] {
			seen[special.ProductUid] = true
			uids = append(uids, special.ProductUid)
		}
	}
	return uids
}

func chunks(values []interface{}) [][]interface{} {
	var result [][]interface{}
	for start := 0; start < len(values); start += syncChunkSize {
		result = append(result, values[start:min(start+syncChunkSize, len(values))])
	}
	return result
}

func placeholders(n int) string {
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

func rowsAffected(res sql.Result) int64 {
	n, err := res.RowsAffected()
	if err != nil {
		return 0
	}
	return n
}

func dateOrZero(date string) string {
	if date == "" {
		return zeroDateTime
	}
	return date
}

func dateOrNull(date string) interface{} {
	if date == "" {
		return nil
	}
	return date
}
