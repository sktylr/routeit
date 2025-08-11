package db

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"
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

func (r *TodoListRepository) GetListsByUser(ctx context.Context, userId string, page, pageSize int) ([]*dao.AggregateTodoList, error) {
	offset := (page - 1) * pageSize
	listQuery := `
		SELECT l.id, l.created, l.updated, l.user_id, l.name, l.description,
		       COUNT(i.id) AS total_items,
		       SUM(CASE WHEN i.status = 'COMPLETED' THEN 1 ELSE 0 END) AS completed_items
		FROM lists l
		LEFT JOIN items i ON l.id = i.list_id
		WHERE l.user_id = ?
		GROUP BY l.id
		ORDER BY l.created DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, listQuery, userId, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lists := []*dao.AggregateTodoList{}
	listIds := []any{}
	for rows.Next() {
		var l dao.AggregateTodoList
		if err := rows.Scan(&l.Id, &l.Created, &l.Updated, &l.UserId, &l.Name, &l.Description, &l.TotalItems, &l.CompletedItems); err != nil {
			return nil, err
		}
		l.Items = []dao.TodoItem{}
		listIds = append(listIds, l.Id)
		lists = append(lists, &l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(lists) == 0 {
		return lists, nil
	}

	itemsByList, err := r.getListItems(ctx, listIds)
	if err != nil {
		return nil, err
	}

	for i := range lists {
		id := lists[i].Id
		lists[i].Items = itemsByList[id]
	}
	return lists, nil
}

func (r *TodoListRepository) getListItems(ctx context.Context, ids []any) (map[string][]dao.TodoItem, error) {
	placeholder := strings.Join(slices.Repeat([]string{"?"}, len(ids)), ",")
	query := fmt.Sprintf(`
		SELECT id, created, updated, user_id, list_id, name, status
		FROM items
		WHERE list_id IN (%s)
		ORDER BY list_id, created ASC
	`, placeholder)
	rows, err := r.db.QueryContext(ctx, query, ids...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	itemsByList := make(map[string][]dao.TodoItem)
	for rows.Next() {
		var item dao.TodoItem
		if err := rows.Scan(
			&item.Id,
			&item.Created,
			&item.Updated,
			&item.UserId,
			&item.TodoListId,
			&item.Name,
			&item.Status,
		); err != nil {
			return nil, err
		}
		if len(itemsByList[item.TodoListId]) < 10 {
			itemsByList[item.TodoListId] = append(itemsByList[item.TodoListId], item)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return itemsByList, nil
}

func (e *ErrListNotFound) Error() string {
	return fmt.Sprintf("todo list with ID %s not found", e.listId)
}
