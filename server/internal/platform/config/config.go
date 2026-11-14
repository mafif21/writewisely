package config

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/joho/godotenv"
)

// Environment variable names.
const (
	envPort       = "PORT"
	envAppEnv     = "APP_ENV"
	envDBHost     = "DB_HOST"
	envDBPort     = "DB_PORT"
	envDBUser     = "DB_USER"
	envDBPassword = "DB_PASSWORD"
	envDBName     = "DB_NAME"
	envDBSSLMode  = "DB_SSLMODE"
)

// Defaults for optional variables.
const (
	defaultPort      = "8080"
	defaultDBPort    = "5432"
	defaultDBSSLMode = "disable"
)

var ErrMissingDatabaseConfig = errors.New("config: required database variables not set")

type Config struct {
	Port     string
	AppEnv   string
	Database DatabaseConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		quote(d.Host), quote(d.Port), quote(d.User), quote(d.Password), quote(d.Name), quote(d.SSLMode))
}

func quote(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, `'`, `\'`)
	return "'" + v + "'"
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("config: load .env: %w", err)
	}
	return fromEnv()
}

func fromEnv() (Config, error) {
	cfg := Config{
		Port:   getenvDefault(envPort, defaultPort),
		AppEnv: os.Getenv(envAppEnv),
		Database: DatabaseConfig{
			Host:     os.Getenv(envDBHost),
			Port:     getenvDefault(envDBPort, defaultDBPort),
			User:     os.Getenv(envDBUser),
			Password: os.Getenv(envDBPassword),
			Name:     os.Getenv(envDBName),
			SSLMode:  getenvDefault(envDBSSLMode, defaultDBSSLMode),
		},
	}

	var missing []string
	for name, val := range map[string]string{
		envDBHost:     cfg.Database.Host,
		envDBUser:     cfg.Database.User,
		envDBPassword: cfg.Database.Password,
		envDBName:     cfg.Database.Name,
	} {
		if val == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		slices.Sort(missing)
		return Config{}, fmt.Errorf("%w: %s", ErrMissingDatabaseConfig, strings.Join(missing, ", "))
	}
	return cfg, nil
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
