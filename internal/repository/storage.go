package repository

import (
	"database/sql"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/apperror"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/infrastructure/database"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/ports"
)

var _ ports.StorageRepository = (*storageRepository)(nil)

type storageRepository struct {
	db *database.SQLDriver
}

func NewStorageRepository(db *database.SQLDriver) *storageRepository {
	return &storageRepository{
		db: db,
	}
}

func (r *storageRepository) CreateBlock(data *model.Block) (*model.Block, error) {
	sqlText := `
		INSERT INTO
			blocks (
				user_id,
				type_id,
				title,
				data,
				salt,
				nonce,
				profile
			)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id;
	`
	err := r.db.Conn.QueryRow(
		sqlText,
		data.UserID,
		data.TypeID,
		data.Title,
		data.Data,
		data.Salt,
		data.Nonce,
		data.Profile,
	).Scan(&data.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.DBErrorNoRows
		}

		return nil, err
	}

	return data, nil
}

func (r *storageRepository) ReadUserBlocks(userID int) ([]*model.Block, error) {
	sqlText := `
		SELECT
			b.id, b.user_id, b.type_id, b.title, b.data, b.profile, b.salt, b.nonce,
			t.id, t.type_name, t.description
		FROM blocks b
		INNER JOIN block_types t ON b.type_id = t.id
		WHERE
			user_id = $1;`

	rows, err := r.db.Conn.Query(sqlText, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.DBErrorNoRows
		}

		return nil, err
	}
	defer rows.Close()

	var blocks []*model.Block
	for rows.Next() {
		var block model.Block
		var t model.Type
		err := rows.Scan(
			&block.ID,
			&block.UserID,
			&block.TypeID,
			&block.Title,
			&block.Data,
			&block.Profile,
			&block.Salt,
			&block.Nonce,
			&t.ID,
			&t.TypeName,
			&t.Description,
		)

		block.Type = &t
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, &block)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return blocks, nil
}

func (r *storageRepository) ReadBlockTypes() ([]*model.Type, error) {
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
	for rows.Next() {
		var t model.Type
		err := rows.Scan(
			&t.ID,
			&t.TypeName,
			&t.Description,
		)
		if err != nil {
			return nil, err
		}

		types = append(types, &t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return types, nil
}
