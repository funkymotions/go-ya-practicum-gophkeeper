package repository

import (
	"database/sql"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/apperror"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
)

type userRepository struct {
	*SQLRepository[model.User]
}

var _ Repository[model.User] = (*userRepository)(nil)

func NewUserRepository(r *SQLRepository[model.User]) *userRepository {
	return &userRepository{
		SQLRepository: r,
	}
}

func (u *userRepository) ReadByID(userID int) (*model.User, error) {
	sqlText := `SELECT id, username, password_hash, created_at FROM users WHERE id = $1;`
	var user model.User
	err := u.db.Conn.QueryRow(sqlText, userID).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, apperror.DBErrorNoRows
	}

	return &user, nil
}

func (u *userRepository) Create(user *model.User) (*model.User, error) {
	sqlText := `
		INSERT INTO
			users (
				username,
				password_hash,
				created_at
			)
		VALUES (
			$1,
			$2,
			NOW()
		)
		RETURNING
			id,
			username,
			password_hash,
			created_at;`

	row := u.db.Conn.QueryRow(
		sqlText,
		user.Username,
		user.PasswordHash,
	)

	result, err := u.scanRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.DBErrorNoRows
		}
		return nil, err
	}

	return result, nil
}

func (u *userRepository) ReadAll() ([]*model.User, error) {
	sqlText := `SELECT id, username, password_hash, created_at FROM users;`
	rows, err := u.db.Conn.Query(sqlText)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.DBErrorNoRows
		}

		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for user := range ModelSeq(rows, u.scanRows) {
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (u *userRepository) ReadByUsername(username string) (*model.User, error) {
	sqlText := `SELECT id, username, password_hash, created_at FROM users WHERE username = $1;`
	row := u.db.Conn.QueryRow(sqlText, username)
	result, err := u.scanRow(row)
	if err == sql.ErrNoRows {
		return nil, apperror.DBErrorNoRows
	}
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (u *userRepository) scanRows(rows *sql.Rows) (*model.User, error) {
	var user model.User
	err := rows.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u *userRepository) scanRow(row *sql.Row) (*model.User, error) {
	var user model.User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
