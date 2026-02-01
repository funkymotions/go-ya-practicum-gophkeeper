package repository

import (
	"database/sql"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/apperror"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/infrastructure/database"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/ports"
)

type userRepository struct {
	db *database.SQLDriver
}

var _ ports.UserRepositoryReader = (*userRepository)(nil)
var _ ports.UserRepositoryWriter = (*userRepository)(nil)

func NewUserRepository(db *database.SQLDriver) *userRepository {
	return &userRepository{
		db: db,
	}
}

func (u *userRepository) ReadUserByID(userID int32) (*model.User, error) {
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

func (u userRepository) ReadUserByUsername(username string) (*model.User, error) {
	sqlText := `SELECT id, username, password_hash, created_at FROM users WHERE username = $1;`
	var user model.User
	err := u.db.Conn.QueryRow(sqlText, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, apperror.DBErrorNoRows
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u *userRepository) CreateUser(user *model.User) (*model.User, error) {
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
		RETURNING id;`

	err := u.db.Conn.QueryRow(
		sqlText,
		user.Username,
		user.PasswordHash,
	).Scan(&user.ID)
	if err == sql.ErrNoRows {
		return nil, apperror.DBErrorNoRows
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}
