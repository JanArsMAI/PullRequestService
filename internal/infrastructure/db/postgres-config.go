package db

import "os"

type PostgresConfig struct {
	User     string
	Password string
	DbName   string
	Host     string
	Port     string
	SSLMode  string
}

func ReadConfig() PostgresConfig {
	return PostgresConfig{
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DbName:   os.Getenv("POSTGRES_DB"),
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		SSLMode:  os.Getenv("DB_SSL"),
	}
}
