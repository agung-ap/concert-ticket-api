package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Database holds the database configuration
type Database struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
}

// Config holds all configuration for the application
type Config struct {
	LogLevel   string   `mapstructure:"log_level"`
	RESTPort   int      `mapstructure:"rest_port"`
	GRPCPort   int      `mapstructure:"grpc_port"`
	Database   Database `mapstructure:"database"`
	MaxRetries int      `mapstructure:"max_retries"`
}

// DSN returns the PostgreSQL connection string
func (d *Database) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.Username, d.Password, d.Name, d.SSLMode,
	)
}

// MigrationDSN returns the PostgreSQL connection string for migrations
func (d *Database) MigrationDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.Username, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

// Load reads the configuration from a file and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("log_level", "info")
	v.SetDefault("rest_port", 8080)
	v.SetDefault("grpc_port", 50051)
	v.SetDefault("max_retries", 3)
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.username", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.name", "concert_tickets")
	v.SetDefault("database.sslmode", "disable")

	// Set config file properties
	configName := filepath.Base(configPath)
	configNameWithoutExt := strings.TrimSuffix(configName, filepath.Ext(configName))
	configDir := filepath.Dir(configPath)

	v.SetConfigName(configNameWithoutExt)
	v.SetConfigType("yaml")
	v.AddConfigPath(configDir)

	// Read from environment variables
	v.AutomaticEnv()
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read from config file
	err := v.ReadInConfig()
	if err != nil {
		// It's ok if config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal config
	var config Config
	err = v.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
