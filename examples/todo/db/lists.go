package db

import "database/sql"

type TodoListRepository struct {
	db *sql.DB
}

func NewTodoListRepository(db *sql.DB) *TodoListRepository {
	return &TodoListRepository{db: db}
}
