package main

import (
	"bake/internal/config"
	"bake/internal/database"
	"bake/internal/generator"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "bake",
		Version: "1.0.0",
		Usage:   "Generate Go models from MySQL database - CLI and YAML config support",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "config.yaml",
				Usage:   "Load configuration from `FILE`",
			},
			&cli.StringFlag{
				Name:    "host",
				Aliases: []string{"H"},
				Value:   "",
				Usage:   "MySQL host",
				EnvVars: []string{"MYSQL_HOST"},
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"P"},
				Value:   0,
				Usage:   "MySQL port",
				EnvVars: []string{"MYSQL_PORT"},
			},
			&cli.StringFlag{
				Name:    "user",
				Aliases: []string{"u"},
				Value:   "",
				Usage:   "MySQL user",
				EnvVars: []string{"MYSQL_USER"},
			},
			&cli.StringFlag{
				Name:    "password",
				Aliases: []string{"p"},
				Value:   "",
				Usage:   "MySQL password",
				EnvVars: []string{"MYSQL_PASSWORD"},
			},
			&cli.StringFlag{
				Name:    "database",
				Aliases: []string{"d"},
				Value:   "",
				Usage:   "MySQL database name",
				EnvVars: []string{"MYSQL_DATABASE"},
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   "",
				Usage:   "Output directory for generated models",
			},
			&cli.StringFlag{
				Name:    "package",
				Aliases: []string{"pkg"},
				Value:   "",
				Usage:   "Package name for generated models",
			},
			&cli.StringSliceFlag{
				Name:    "tables",
				Aliases: []string{"t"},
				Usage:   "Specific tables to generate (default: all tables)",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Value:   false,
				Usage:   "Verbose output",
			},
		},
		Action: func(c *cli.Context) error {
			return generateModels(c)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func generateModels(c *cli.Context) error {
	configFile := c.String("config")

	cfg, err := config.LoadConfig(configFile)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Could not load config file: %v", err)
	}

	finalConfig := config.MergeConfig(cfg, c)

	if finalConfig.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	if c.Bool("verbose") {
		log.Printf("Using configuration: %+v", finalConfig)
	}

	db, err := database.Connect(
		finalConfig.Database.Host,
		finalConfig.Database.Port,
		finalConfig.Database.User,
		finalConfig.Database.Password,
		finalConfig.Database.Name,
	)
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}
	defer db.Close()

	tables, err := generator.GetTables(db, finalConfig.Tables)
	if err != nil {
		return fmt.Errorf("error getting tables: %v", err)
	}

	if len(tables) == 0 {
		return fmt.Errorf("no tables found in database %s", finalConfig.Database.Name)
	}

	templatePath := filepath.Join("templates", "model.tmpl")
	tmpl, err := generator.LoadTemplate(templatePath)
	if err != nil {
		return fmt.Errorf("error loading template: %v", err)
	}

	if err := os.MkdirAll(finalConfig.OutputDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %v", err)
	}

	if err := generator.GeneratePackageFile(finalConfig.OutputDir, finalConfig.PackageName); err != nil {
		return fmt.Errorf("error generating package file: %v", err)
	}

	for _, table := range tables {
		err := generator.GenerateModelFile(tmpl, finalConfig.OutputDir, finalConfig.PackageName, table)
		if err != nil {
			log.Printf("Error generating model for table %s: %v", table.Name, err)
		} else if c.Bool("verbose") {
			log.Printf("Generated model for table: %s", table.Name)
		}
	}

	log.Printf("Successfully generated %d model files in %s", len(tables), finalConfig.OutputDir)
	return nil
}
