package repository

import (
	"database/sql"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/apperror"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
)

type typeRepository struct {
	*SQLRepository[model.Type]
}

var _ Repository[model.Type] = (*typeRepository)(nil)

func NewTypeRepository(r *SQLRepository[model.Type]) *typeRepository {
	return &typeRepository{
		SQLRepository: r,
	}
}

func (r *typeRepository) Create(data *model.Type) (*model.Type, error) {
	sqlText := `
		INSERT INTO
			block_types (
				type_name,
				description
			)
		VALUES ($1, $2)
		RETURNING
			id,
			type_name,
			description;
	`
	row := r.db.Conn.QueryRow(
		sqlText,
		data.TypeName,
		data.Description,
	)

	result, err := r.scanRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.DBErrorNoRows
		}

		return nil, err
	}

	return result, nil
}

func (r *typeRepository) ReadByID(id int) (*model.Type, error) {
	sqlText := `
		SELECT
			id,
			type_name,
			description
		FROM block_types
		WHERE id = $1;`

	var t *model.Type
	row := r.db.Conn.QueryRow(sqlText, id)
	t, err := r.scanRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.DBErrorNoRows
		}

		return nil, err
	}

	return t, nil
}

func (r *typeRepository) ReadAll() ([]*model.Type, error) {
	sqlText := `
		SELECT
			id, type_name, description
		FROM block_types;`

	rows, err := r.db.Conn.Query(sqlText)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.DBErrorNoRows
		}

		return nil, err
	}

	defer rows.Close()

	var types []*model.Type
	for t := range ModelSeq(rows, r.scanRows) {
		types = append(types, t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return types, nil
}

func (r *typeRepository) scanRows(rows *sql.Rows) (*model.Type, error) {
	var t model.Type
	err := rows.Scan(
		&t.ID,
		&t.TypeName,
		&t.Description,
	)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *typeRepository) scanRow(row *sql.Row) (*model.Type, error) {
	var t model.Type
	err := row.Scan(
		&t.ID,
		&t.TypeName,
		&t.Description,
	)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
