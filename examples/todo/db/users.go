package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

func (r *UsersRepository) CreateUser(ctx context.Context, name, email, password string) (*dao.User, error) {
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
		ctx,
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
	return r.getUserByX(ctx, "email", email)
}

func (r *UsersRepository) GetUserById(ctx context.Context, id string) (*dao.User, bool, error) {
	return r.getUserByX(ctx, "id", id)
}

func (r *UsersRepository) getUserByX(ctx context.Context, column, value string) (*dao.User, bool, error) {
	query := fmt.Sprintf("SELECT id, name, email, password, created, updated FROM users WHERE %s = ?", column)
	row := r.db.QueryRowContext(ctx, query, value)

	var user dao.User
	err := row.Scan(
		&user.Id,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Created,
		&user.Updated,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	return &user, true, nil
}
