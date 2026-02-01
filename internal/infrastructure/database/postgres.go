package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/config"
	_ "github.com/lib/pq"
)

type SQLDriver struct {
	Conn *sql.DB
}

func NewSQLDriver(conf *config.Database) (*SQLDriver, error) {
	db, err := sql.Open("postgres", conf.GetDSN())
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(conf.ConnTimeout)*time.Millisecond,
	)

	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return &SQLDriver{Conn: db}, nil
}
