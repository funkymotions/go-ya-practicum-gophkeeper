package repository

import (
	"database/sql"
	"iter"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/infrastructure/database"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
)

type modelScoped interface {
	model.Block | model.Type | model.User
}

type Repository[T modelScoped] interface {
	Create(item *T) (*T, error)
	ReadByID(id int) (*T, error)
	ReadAll() ([]*T, error)
	scanRows(rows *sql.Rows) (*T, error)
	scanRow(*sql.Row) (*T, error)
}

type SQLRepository[T modelScoped] struct {
	db *database.SQLDriver
}

func NewSQLRepository[T modelScoped](db *database.SQLDriver) *SQLRepository[T] {
	return &SQLRepository[T]{
		db: db,
	}
}

func ModelSeq[T modelScoped](rows *sql.Rows, scanFn func(*sql.Rows) (*T, error)) iter.Seq[*T] {
	return func(yield func(*T) bool) {
		for rows.Next() {
			item, err := scanFn(rows)
			if err != nil {
				break
			}
			if !yield(item) {
				return
			}
		}
	}
}
