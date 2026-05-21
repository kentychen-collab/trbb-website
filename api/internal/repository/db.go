package repository

import (
	"database/sql"
	"fmt"
	"sports-platform/internal/config"

	_ "github.com/go-sql-driver/mysql"
)

func NewDB(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("db ping failed: %w", err)
	}
	return db, nil
}
