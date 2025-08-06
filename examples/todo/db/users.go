package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sktylr/routeit/examples/todo/auth"
	"github.com/sktylr/routeit/examples/todo/dao"
)

type UsersRepository struct {
	db *sql.DB
}

func NewUsersRepository(db *sql.DB) *UsersRepository {
	return &UsersRepository{db: db}
}

func (r *UsersRepository) CreateUser(context context.Context, name, email, password string) (*dao.User, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	idS := id.String()

	hashedPw, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	_, err = r.db.ExecContext(
		context,
		"INSERT INTO users (id, name, email, password, created, updated) VALUES (?, ?, ?, ?, ?, ?)",
		idS,
		name,
		email,
		hashedPw,
		now,
		now,
	)
	if err != nil {
		return nil, err
	}

	user := dao.User{
		Id:       idS,
		Created:  now,
		Updated:  now,
		Name:     name,
		Email:    email,
		Password: hashedPw,
	}
	return &user, nil
}

func (r *UsersRepository) GetUserByEmail(ctx context.Context, email string) (*dao.User, bool, error) {
	query := `
		SELECT id, name, email, password, created, updated
		FROM users
		WHERE email = ?
	`

	row := r.db.QueryRowContext(ctx, query, email)

	var user dao.User
	err := row.Scan(
		&user.Id,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Created,
		&user.Updated,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &user, true, nil
}
