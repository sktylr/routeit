package db

import (
	"context"
	"database/sql"
	"time"
)

type TodoItemRepository struct {
	db *sql.DB
}

func NewTodoItemRepository(db *sql.DB) *TodoItemRepository {
	return &TodoItemRepository{db: db}
}

func (r *TodoItemRepository) MarkAsCompleted(ctx context.Context, id string) error {
	query := `
		UPDATE items
		SET status = 'COMPLETED',
			updated = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *TodoItemRepository) MarkAsPending(ctx context.Context, id string) error {
	query := `
		UPDATE items
		SET status = 'PENDING',
		    updated = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}
