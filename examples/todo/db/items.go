package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type TodoItemRepository struct {
	db *sql.DB
}

func NewTodoItemRepository(db *sql.DB) *TodoItemRepository {
	return &TodoItemRepository{db: db}
}

func (r *TodoItemRepository) MarkAsCompleted(ctx context.Context, id string) error {
	return r.markAsX(ctx, id, "COMPLETED")
}

func (r *TodoItemRepository) MarkAsPending(ctx context.Context, id string) error {
	return r.markAsX(ctx, id, "PENDING")
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
