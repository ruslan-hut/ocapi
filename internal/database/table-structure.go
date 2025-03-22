package database

import (
	"database/sql"
	"errors"
	"fmt"
)

type Tables struct {
	Product map[string]Column
}

type Column struct {
	Name         string  // Имя столбца
	DefaultValue *string // Значение по умолчанию (nil, если в БД оно NULL)
	IsNullable   bool    // Разрешает ли столбец NULL
	DataType     string  // Тип данных (например, 'int', 'varchar' и т.д.)
}

func (s *MySql) LoadTablesStructure() (Tables, error) {
	tables := Tables{
		Product: make(map[string]Column),
	}

	// Load product table structure
	productColumns, err := s.LoadProductTableStructure("product")
	if err != nil {
		return tables, fmt.Errorf("product table: %w", err)
	}
	tables.Product = productColumns

	return tables, nil
}

// LoadProductTableStructure считывает структуру столбцов из information_schema
// и возвращает её в виде map[имя_колонки]ColumnInfo.
func (s *MySql) LoadProductTableStructure(tableName string) (map[string]Column, error) {
	query := fmt.Sprintf(`
        SELECT COLUMN_NAME, COLUMN_DEFAULT, IS_NULLABLE, DATA_TYPE
          FROM information_schema.columns
         WHERE table_name = '%s%s'
         ORDER BY ORDINAL_POSITION`, s.prefix, tableName)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	columns := make(map[string]Column)

	for rows.Next() {
		var colName, colDefault, isNullable, dataType string

		// Считываем строку
		if err = rows.Scan(&colName, &colDefault, &isNullable, &dataType); err != nil {
			return nil, fmt.Errorf("failed to scan column info: %w", err)
		}

		// Преобразуем флаг "YES"/"NO" в логический
		nullable := isNullable == "YES"

		// Чтобы отличить "NULL в базе" от "пустой строки", храним DefaultValue в *string
		// Если colDefault == "" (пустая строка) и колонка в БД действительно имеет DEFAULT=NULL,
		// то это будет корректно считаться как nil.
		// В некоторых СУБД пустая строка и NULL могут отличаться.
		var defValPtr *string
		if colDefault != "" {
			defValPtr = &colDefault
		}

		columns[colName] = Column{
			Name:         colName,
			DefaultValue: defValPtr,
			IsNullable:   nullable,
			DataType:     dataType,
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("after scanning rows: %w", err)
	}

	return columns, nil
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
