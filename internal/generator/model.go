package generator

import (
	"database/sql"
)

type Column struct {
	Name     string
	Type     string
	Nullable bool
	Comment  string
}

type Table struct {
	Name    string
	Columns []Column
}

func GetTables(db *sql.DB, specificTables []string) ([]Table, error) {
	var tables []Table
	if len(specificTables) > 0 {
		for _, tableName := range specificTables {
			table, err := GetTableColumns(db, tableName)
			if err != nil {
				return nil, err
			}
			tables = append(tables, table)
		}
	} else {
		rows, err := db.Query(`
			SELECT TABLE_NAME 
			FROM INFORMATION_SCHEMA.TABLES 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_TYPE = 'BASE TABLE'
		`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				return nil, err
			}

			table, err := GetTableColumns(db, tableName)
			if err != nil {
				return nil, err
			}

			tables = append(tables, table)
		}
	}

	return tables, nil
}

func GetTableColumns(db *sql.DB, tableName string) (Table, error) {
	rows, err := db.Query(`
		SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_COMMENT
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_NAME = ? AND TABLE_SCHEMA = DATABASE()
		ORDER BY ORDINAL_POSITION
	`, tableName)
	if err != nil {
		return Table{}, err
	}
	defer rows.Close()

	table := Table{Name: tableName}
	for rows.Next() {
		var colName, dataType, isNullable, comment string
		if err := rows.Scan(&colName, &dataType, &isNullable, &comment); err != nil {
			return Table{}, err
		}

		column := Column{
			Name:     colName,
			Type:     dataType,
			Nullable: isNullable == "YES",
			Comment:  comment,
		}
		table.Columns = append(table.Columns, column)
	}

	return table, nil
}
