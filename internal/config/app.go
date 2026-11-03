package config

import "os"

type Config struct {
	Address, Database string
}

// Load application settings from environment variables with local defaults
func Load() Config {
	c := Config{
		Address:  os.Getenv("HTTP_ADDR"),
		Database: os.Getenv("DATABASE_PATH"),
	}

	if c.Address == "" {
		c.Address = ":8080"
	}
	if c.Database == "" {
		c.Database = "invest-pool.db"
	}

	return c
}
