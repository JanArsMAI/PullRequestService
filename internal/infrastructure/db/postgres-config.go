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
		user:     os.Getenv("DB_USER"),
		password: os.Getenv("DB_PASSWORD"),
		dbname:   os.Getenv("DB_NAME"),
		host:     os.Getenv("DB_HOST"),
		port:     os.Getenv("DB_PORT"),
		ssl:      os.Getenv("DB_SSL"),
	}
}
