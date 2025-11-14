package db

import "os"

type PostgresConfig struct {
	user     string
	password string
	dbname   string
	host     string
	port     string
	ssl      string
}

func ReadConfig() PostgresConfig {
	return PostgresConfig{
		user:     os.Getenv("POSTGRES_USER"),
		password: os.Getenv("POSTGRES_PASSWORD"),
		dbname:   os.Getenv("POSTGRES_DB"),
		host:     os.Getenv("POSTGRES_HOST"),
		port:     os.Getenv("POSTGRES_PORT"),
		ssl:      os.Getenv("DB_SSL"),
	}
}
