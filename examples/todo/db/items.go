package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sktylr/routeit/examples/todo/dao"
)

type TodoItemRepository struct {
	db *sql.DB
}

type ErrItemNotFound struct {
	itemId string
}

func NewTodoItemRepository(db *sql.DB) *TodoItemRepository {
	return &TodoItemRepository{db: db}
}

func (r *TodoItemRepository) CreateItem(ctx context.Context, userId, listId, name string) (*dao.TodoItem, error) {
	uuid, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	id := uuid.String()

	now := time.Now()
	query := `
		INSERT INTO items (id, created, updated, user_id, list_id, name, status)
		VALUES (?, ?, ?, ?, ?, ?, 'PENDING')
	`
	_, err = r.db.ExecContext(ctx, query, id, now, now, userId, listId, name)
	if err != nil {
		return nil, err
	}

	item := dao.TodoItem{
		Meta: dao.Meta{
			Id:      id,
			Created: now,
			Updated: now,
		},
		UserId:     userId,
		TodoListId: listId,
		Name:       name,
		Status:     "PENDING",
	}
	return &item, nil
}

func (r *TodoItemRepository) GetById(ctx context.Context, id string) (*dao.TodoItem, error) {
	query := `
		SELECT id, created, updated, user_id, list_id, name, status
		FROM items
		WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var item dao.TodoItem
	err := row.Scan(
		&item.Id,
		&item.Created,
		&item.Updated,
		&item.UserId,
		&item.TodoListId,
		&item.Name,
		&item.Status,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrItemNotFound{itemId: id}
		}
		return nil, err
	}

	return &item, nil
}

func (r *TodoItemRepository) UpdateName(ctx context.Context, id, newName string) error {
	query := `
		UPDATE items
		SET name = ?, updated = ?
		WHERE id = ?
	`
	res, err := r.db.ExecContext(ctx, query, newName, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrItemNotFound{itemId: id}
	}

	return nil
}

func (r *TodoItemRepository) MarkAsCompleted(ctx context.Context, id string) error {
	return r.markAsX(ctx, id, "COMPLETED")
}

func (r *TodoItemRepository) MarkAsPending(ctx context.Context, id string) error {
	return r.markAsX(ctx, id, "PENDING")
}

func (r *TodoItemRepository) DeleteItem(ctx context.Context, id string) error {
	query := `
		DELETE FROM items
		WHERE id = ?
	`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrItemNotFound{itemId: id}
	}

	return nil
}

func (r *TodoItemRepository) markAsX(ctx context.Context, id, status string) error {
	query := fmt.Sprintf(`
		UPDATE items
		SET status = '%s',
		    updated = ?
		WHERE id = ?
	`, status)
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (e ErrItemNotFound) Error() string {
	return fmt.Sprintf("todo item with ID %q not found", e.itemId)
}
