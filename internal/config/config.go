package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBName     string
	DBPort     string
	// DBPath     string
}

func NewConfig() Config {
	return Config{
		DBUser:     os.Getenv("DATABASE_USER"),
		DBPassword: os.Getenv("DATABASE_PASSWORD"),
		DBHost:     os.Getenv("DATABASE_HOST"),
		DBName:     os.Getenv("DATABASE_NAME"),
		DBPort:     os.Getenv("DATABASE_PORT"),
	}
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func (c *Config) GetIntegDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, 5433, c.DBName)
}
