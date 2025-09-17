package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"bake/internal/utils"
)

type TemplateData struct {
	PackageName string
	StructName  string
	TableName   string
	Columns     []TemplateColumn
	Imports     []string
}

type TemplateColumn struct {
	DBName  string
	GoName  string
	GoType  string
	Comment string
}

func GenerateModelFile(tmpl *template.Template, outputDir, packageName string, table Table) error {

	var columns []TemplateColumn
	importSet := make(map[string]struct{})
	for _, col := range table.Columns {
		goType := MapSQLTypeToGoType(col.Type, col.Nullable)
		switch goType {
		case "time.Time", "*time.Time":
			importSet["time"] = struct{}{}
		case "sql.NullString", "sql.NullInt64", "sql.NullFloat64", "sql.NullBool":
			importSet["database/sql"] = struct{}{}
		}

		columns = append(columns, TemplateColumn{
			DBName:  col.Name,
			GoName:  utils.ToCamelCase(col.Name),
			GoType:  goType,
			Comment: col.Comment,
		})

	}
	var imports []string
	for imp := range importSet {
		imports = append(imports, imp)
	}

	structName := utils.ToCamelCase(table.Name)

	data := TemplateData{
		PackageName: packageName,
		StructName:  structName,
		TableName:   table.Name,
		Columns:     columns,
		Imports:     imports,
	}

	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.gen.go", strings.ToLower(table.Name)))
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}
