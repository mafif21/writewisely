package config

import (
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setDBEnv(t *testing.T) {
	t.Helper()
	t.Setenv(envDBHost, "localhost")
	t.Setenv(envDBUser, "u")
	t.Setenv(envDBPassword, "p")
	t.Setenv(envDBName, "db")
	t.Setenv(envDBPort, "")
	t.Setenv(envDBSSLMode, "")
}

func TestFromEnv_Defaults(t *testing.T) {
	setDBEnv(t)
	t.Setenv(envPort, "")
	t.Setenv(envAppEnv, "")

	cfg, err := fromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != defaultPort {
		t.Errorf("Port = %q, want %q", cfg.Port, defaultPort)
	}
	if cfg.Database.Port != defaultDBPort {
		t.Errorf("Database.Port = %q, want %q", cfg.Database.Port, defaultDBPort)
	}
	if cfg.Database.SSLMode != defaultDBSSLMode {
		t.Errorf("Database.SSLMode = %q, want %q", cfg.Database.SSLMode, defaultDBSSLMode)
	}
}

func TestFromEnv_Overrides(t *testing.T) {
	setDBEnv(t)
	t.Setenv(envPort, "9090")
	t.Setenv(envAppEnv, "development")
	t.Setenv(envDBPort, "5433")
	t.Setenv(envDBSSLMode, "require")

	cfg, err := fromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := DatabaseConfig{Host: "localhost", Port: "5433", User: "u", Password: "p", Name: "db", SSLMode: "require"}
	if cfg.Port != "9090" || cfg.AppEnv != "development" || cfg.Database != want {
		t.Errorf("got %+v", cfg)
	}
}

func TestFromEnv_MissingRequired(t *testing.T) {
	setDBEnv(t)
	t.Setenv(envDBHost, "")
	t.Setenv(envDBPassword, "")

	_, err := fromEnv()
	if !errors.Is(err, ErrMissingDatabaseConfig) {
		t.Fatalf("err = %v, want ErrMissingDatabaseConfig", err)
	}
	if !strings.Contains(err.Error(), "DB_HOST, DB_PASSWORD") {
		t.Errorf("err = %q, want it to name DB_HOST, DB_PASSWORD", err)
	}
}

func TestDSN_ParsesSpecialCharacters(t *testing.T) {
	d := DatabaseConfig{
		Host:     "db.example.com",
		Port:     "5433",
		User:     "app user",
		Password: `p@ss w:o/r?d#'x\y=`,
		Name:     "write wisely",
		SSLMode:  "disable",
	}

	pc, err := pgxpool.ParseConfig(d.DSN())
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}
	cc := pc.ConnConfig
	if cc.Host != d.Host || cc.Port != 5433 || cc.User != d.User || cc.Password != d.Password || cc.Database != d.Name {
		t.Errorf("parsed host=%q port=%d user=%q password=%q db=%q", cc.Host, cc.Port, cc.User, cc.Password, cc.Database)
	}
}
