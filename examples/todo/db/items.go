package db

import "database/sql"

type TodoItemRepository struct {
	db *sql.DB
}

func NewTodoItemRepository(db *sql.DB) *TodoItemRepository {
	return &TodoItemRepository{db: db}
}
