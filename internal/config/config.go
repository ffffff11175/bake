package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

type Config struct {
	Database    DatabaseConfig `yaml:"database"`
	OutputDir   string         `yaml:"output_dir"`
	PackageName string         `yaml:"package_name"`
	Tables      []string       `yaml:"tables"`
}

func LoadConfig(path string) (*Config, error) {
	config := &Config{}

	if path == "" {
		path = "config.yaml"
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file %s does not exist", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %v", err)
	}

	return config, nil
}

func MergeConfig(fileConfig *Config, c *cli.Context) *Config {
	finalConfig := &Config{}

	if fileConfig != nil {
		finalConfig = fileConfig
	}

	finalConfig.Database = mergeDatabaseConfig(finalConfig.Database, c)

	finalConfig.OutputDir = mergeString(finalConfig.OutputDir, c.String("output"), "./models")
	finalConfig.PackageName = mergeString(finalConfig.PackageName, c.String("package"), "models")
	finalConfig.Tables = mergeTables(finalConfig.Tables, c.StringSlice("tables"))

	return finalConfig
}

func mergeDatabaseConfig(fileDB DatabaseConfig, c *cli.Context) DatabaseConfig {
	dbConfig := DatabaseConfig{
		Host:     mergeString(fileDB.Host, c.String("host"), "localhost"),
		Port:     mergeInt(fileDB.Port, c.Int("port"), 3306),
		User:     mergeString(fileDB.User, c.String("user"), "root"),
		Password: mergeString(fileDB.Password, c.String("password"), ""),
		Name:     mergeString(fileDB.Name, c.String("database"), ""),
	}
	return dbConfig
}

func mergeString(fileValue, cliValue, defaultValue string) string {
	if cliValue != "" {
		return cliValue
	}
	if fileValue != "" {
		return fileValue
	}
	return defaultValue
}

func mergeInt(fileValue, cliValue, defaultValue int) int {
	if cliValue != 0 {
		return cliValue
	}
	if fileValue != 0 {
		return fileValue
	}
	return defaultValue
}

func mergeTables(fileTables, cliTables []string) []string {
	if len(cliTables) > 0 {
		return cliTables
	}
	return fileTables
}

func ValidateConfig(config *Config) error {
	if config.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if config.Database.Port <= 0 || config.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", config.Database.Port)
	}

	if config.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	if config.OutputDir == "" {
		return fmt.Errorf("output directory is required")
	}

	if config.PackageName == "" {
		return fmt.Errorf("package name is required")
	}

	return nil
}

func CreateDefaultConfig(path string) error {
	if path == "" {
		path = "config.yaml"
	}

	defaultConfig := &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     3306,
			User:     "root",
			Password: "your_password",
			Name:     "your_database",
		},
		OutputDir:   "./models",
		PackageName: "models",
		Tables:      []string{},
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	if err := encoder.Encode(defaultConfig); err != nil {
		return fmt.Errorf("failed to encode config: %v", err)
	}

	return nil
}

// FindConfigFile 查找配置文件
func FindConfigFile(preferredPath string) string {
	if preferredPath != "" {
		if _, err := os.Stat(preferredPath); err == nil {
			return preferredPath
		}
	}

	if _, err := os.Stat("config.yaml"); err == nil {
		return "config.yaml"
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		configPath := filepath.Join(exeDir, "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}
	return ""
}
