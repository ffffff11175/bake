package generator

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func LoadTemplate(templatePath string) (*template.Template, error) {

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		// 如果不存在，使用内置默认模板
		return template.New("model").Parse(defaultTemplate)
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

// 默认模板内容
const defaultTemplate = `package {{.PackageName}}

import (
	{{if .HasTimeFields}}"time"{{end}}
)

// {{.StructName}} represents a row from the '{{.TableName}}' table.
type {{.StructName}} struct {
	{{range .Columns}}
	{{.GoName}} {{.GoType}} ` + "`db:\"{{.DBName}}\" json:\"{{.DBName}}\"`" + ` // {{.Comment}}
	{{end}}
}

// TableName returns the name of the table.
func ({{.StructName}}) TableName() string {
	return "{{.TableName}}"
}`
