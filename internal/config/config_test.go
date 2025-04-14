package config_test

import (
	"os"
	"testing"

	"github.com/mi4r/avito-pvz/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	os.Setenv("DATABASE_USER", "testuser")
	os.Setenv("DATABASE_PASSWORD", "testpassword")
	os.Setenv("DATABASE_HOST", "localhost")
	os.Setenv("DATABASE_NAME", "testdb")
	os.Setenv("DATABASE_PORT", "5432")

	cfg := config.NewConfig()

	assert.Equal(t, "testuser", cfg.DBUser)
	assert.Equal(t, "testpassword", cfg.DBPassword)
	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, "testdb", cfg.DBName)
	assert.Equal(t, "5432", cfg.DBPort)
}

func TestGetDSN(t *testing.T) {
	cfg := config.Config{
		DBUser:     "testuser",
		DBPassword: "testpassword",
		DBHost:     "localhost",
		DBName:     "testdb",
		DBPort:     "5432",
	}

	expectedDSN := "postgres://testuser:testpassword@localhost:5432/testdb?sslmode=disable"
	assert.Equal(t, expectedDSN, cfg.GetDSN())
}

func TestGetIntegDSN(t *testing.T) {
	cfg := config.Config{
		DBUser:     "testuser",
		DBPassword: "testpassword",
		DBHost:     "localhost",
		DBName:     "testdb",
	}

	expectedDSN := "postgres://testuser:testpassword@localhost:5433/testdb?sslmode=disable"
	assert.Equal(t, expectedDSN, cfg.GetIntegDSN())
}
