package db

import (
	"database/sql"
	"fmt"

	"avito_coin/internal/config"
	"avito_coin/internal/resource"
	// Импортируем драйвер для работы с PostgreSQL через database/sql.
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func NewPostgresDB(cfg config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err = MigrateDB(db); err != nil {
		return nil, err
	}

	return db, err
}

func MigrateDB(db *sql.DB) error {
	goose.SetBaseFS(resource.EmbedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	return nil
}
