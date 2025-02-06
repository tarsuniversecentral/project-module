package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds the database credentials and other configuration parameters.
type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
}

// LoadConfig loads the environment variables from the .env file and returns a Config instance.
func LoadConfig() (*Config, error) {
	// Load environment variables from the .env file.
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	cfg := &Config{
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBName:     os.Getenv("DB_NAME"),
	}

	return cfg, nil
}
