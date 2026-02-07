package repository

import (
	"database/sql"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/apperror"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
)

var _ Repository[model.Block] = (*blockRepository)(nil)

type blockRepository struct {
	*SQLRepository[model.Block]
}

func NewBlockRepository(r *SQLRepository[model.Block]) *blockRepository {
	return &blockRepository{
		SQLRepository: r,
	}
}

func (r *blockRepository) Create(data *model.Block) (*model.Block, error) {
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
		RETURNING
			id,
			user_id,
			type_id,
			title,
			data,
			salt,
			nonce,
			profile
		;`

	row := r.db.Conn.QueryRow(
		sqlText,
		data.UserID,
		data.TypeID,
		data.Title,
		data.Data,
		data.Salt,
		data.Nonce,
		data.Profile,
	)

	result, err := r.scanRowNoType(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.DBErrorNoRows
		}

		return nil, err
	}

	return result, nil
}

func (r *blockRepository) ReadByID(id int) (*model.Block, error) {
	sqlText := `
		SELECT
			b.id, b.user_id, b.type_id, b.title, b.data, b.salt, b.nonce, b.profile, 
			t.id, t.type_name, t.description
		FROM blocks b
		INNER JOIN block_types t ON b.type_id = t.id
		WHERE
			b.id = $1;`

	var block *model.Block
	row := r.db.Conn.QueryRow(sqlText, id)
	block, err := r.scanRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.DBErrorNoRows
		}

		return nil, err
	}

	return block, nil
}

func (r *blockRepository) ReadAll() ([]*model.Block, error) {
	sqlText := `
		SELECT
			b.id, b.user_id, b.type_id, b.title, b.data, b.salt, b.nonce, b.profile,
			t.id, t.type_name, t.description
		FROM blocks b
		INNER JOIN block_types t ON b.type_id = t.id;`

	rows, err := r.db.Conn.Query(sqlText)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.DBErrorNoRows
		}

		return nil, err
	}
	defer rows.Close()

	var blocks []*model.Block
	for block := range ModelSeq(rows, r.scanRows) {
		blocks = append(blocks, block)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return blocks, nil
}

func (r *blockRepository) ReadByUserID(userID int) ([]*model.Block, error) {
	sqlText := `
		SELECT
			b.id, b.user_id, b.type_id, b.title, b.data, b.salt, b.nonce, b.profile,
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
	for block := range ModelSeq(rows, r.scanRows) {
		blocks = append(blocks, block)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return blocks, nil
}

func (b *blockRepository) scanRows(rows *sql.Rows) (*model.Block, error) {
	var block model.Block
	var t model.Type
	err := rows.Scan(
		&block.ID,
		&block.UserID,
		&block.TypeID,
		&block.Title,
		&block.Data,
		&block.Salt,
		&block.Nonce,
		&block.Profile,
		&t.ID,
		&t.TypeName,
		&t.Description,
	)

	block.Type = &t
	if err != nil {
		return nil, err
	}

	return &block, nil
}

func (b *blockRepository) scanRow(row *sql.Row) (*model.Block, error) {
	var block model.Block
	var t model.Type
	err := row.Scan(
		&block.ID,
		&block.UserID,
		&block.TypeID,
		&block.Title,
		&block.Data,
		&block.Salt,
		&block.Nonce,
		&block.Profile,
		&t.ID,
		&t.TypeName,
		&t.Description,
	)

	block.Type = &t
	if err != nil {
		return nil, err
	}

	return &block, nil
}

func (b *blockRepository) scanRowNoType(row *sql.Row) (*model.Block, error) {
	var block model.Block
	err := row.Scan(
		&block.ID,
		&block.UserID,
		&block.TypeID,
		&block.Title,
		&block.Data,
		&block.Salt,
		&block.Nonce,
		&block.Profile,
	)
	if err != nil {
		return nil, err
	}

	return &block, nil
}
