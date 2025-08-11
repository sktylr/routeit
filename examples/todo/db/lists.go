package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sktylr/routeit/examples/todo/dao"
)

type TodoListRepository struct {
	db *sql.DB
}

type ErrListNotFound struct {
	listId string
}

func NewTodoListRepository(db *sql.DB) *TodoListRepository {
	return &TodoListRepository{db: db}
}

func (r *TodoListRepository) CreateList(ctx context.Context, userId, name, description string) (*dao.TodoList, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate list ID: %w", err)
	}
	idS := id.String()

	now := time.Now()
	query := `
		INSERT INTO lists (id, created, updated, user_id, name, description)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err = r.db.ExecContext(ctx, query, idS, now, now, userId, name, description)
	if err != nil {
		return nil, fmt.Errorf("failed to create list: %w", err)
	}

	list := dao.TodoList{
		Meta: dao.Meta{
			Id:      idS,
			Created: now,
			Updated: now,
		},
		UserId:      userId,
		Name:        name,
		Description: description,
	}
	return &list, nil
}

func (r *TodoListRepository) UpdateList(ctx context.Context, listId, name, description string) (*dao.TodoList, error) {
	now := time.Now()
	query := `
		UPDATE lists
		SET name = ?, description = ?, updated = ?
		WHERE id = ?
	`
	res, err := r.db.ExecContext(ctx, query, name, description, now, listId)
	if err != nil {
		return nil, fmt.Errorf("failed to update list: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to determine rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return nil, &ErrListNotFound{listId: listId}
	}

	list := &dao.TodoList{
		Meta: dao.Meta{
			Id:      listId,
			Updated: now,
		},
		Name:        name,
		Description: description,
	}

	return list, nil
}

func (e *ErrListNotFound) Error() string {
	return fmt.Sprintf("todo list with ID %s not found", e.listId)
}
