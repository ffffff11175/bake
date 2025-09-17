package generator

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func LoadTemplate(templatePath string) (*template.Template, error) {
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return nil, err
	}

	tmpl, err := template.New(filepath.Base(templatePath)).ParseFiles(templatePath)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func MapSQLTypeToGoType(sqlType string, nullable bool) string {
	var goType string
	switch strings.ToLower(sqlType) {
	case "tinyint", "smallint", "mediumint", "int", "integer":
		goType = "int"
	case "bigint":
		goType = "int64"
	case "float", "double", "decimal":
		goType = "float64"
	case "char", "varchar", "text", "tinytext", "mediumtext", "longtext", "json":
		goType = "string"
	case "date", "datetime", "timestamp", "time":
		goType = "time.Time"
	case "binary", "varbinary", "blob", "tinyblob", "mediumblob", "longblob":
		goType = "[]byte"
	case "bit", "bool", "boolean":
		goType = "bool"
	default:
		goType = "interface{}"
	}

	if nullable && goType != "string" && goType != "[]byte" && goType != "interface{}" {
		goType = "*" + goType
	}

	return goType
}
