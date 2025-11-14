package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewPostgresConnection(cfg PostgresConfig) (*sqlx.DB, error) {
	ds := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s",
		cfg.user,
		cfg.password,
		cfg.dbname,
		cfg.host,
		cfg.port,
		cfg.ssl,
	)
	db, err := sqlx.Connect("postgres", ds)
	if err != nil {
		return nil, fmt.Errorf("failed connection to db: %w", err)
	}
	//на всякий случай пингуем БД, хоть и внутри Коннекта это уже вшито
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}
	return db, nil
}
