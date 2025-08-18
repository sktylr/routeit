package db

import (
	"context"
	"database/sql"
	"errors"
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

func NewTodoListRepository(db *sql.DB) *TodoListRepository {
	return &TodoListRepository{db: db}
}

func (r *TodoListRepository) CreateList(ctx context.Context, userId, name, description string) (*dao.TodoList, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("%w [failed to generate list id]: %v", ErrDatabaseIssue, err)
	}
	idS := id.String()

	now := time.Now()
	query := `
		INSERT INTO lists (id, created, updated, user_id, name, description)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err = r.db.ExecContext(ctx, query, idS, now, now, userId, name, description)
	if err != nil {
		return nil, fmt.Errorf("%w [failed to create list]: %v", ErrDatabaseIssue, err)
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
		return nil, fmt.Errorf("%w [failed to update list]: %v", ErrDatabaseIssue, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("%w [failed to determine rows affected]: %v", ErrDatabaseIssue, err)
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("%w [%s]", ErrListNotFound, listId)
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

func (r *TodoListRepository) DeleteList(ctx context.Context, id string) error {
	query := `DELETE FROM lists WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseIssue, err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseIssue, err)
	}
	if rows == 0 {
		return fmt.Errorf("%w [%s]", ErrListNotFound, id)
	}
	return nil
}

func (r *TodoListRepository) GetListById(ctx context.Context, id string) (*dao.TodoList, error) {
	query := `
		SELECT id, created, updated, user_id, name, description
		FROM lists
		WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var list dao.TodoList
	err := row.Scan(
		&list.Id,
		&list.Created,
		&list.Updated,
		&list.UserId,
		&list.Name,
		&list.Description,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w [%s]: %v", ErrListNotFound, id, err)
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseIssue, err)
	}

	return &list, nil
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
		return nil, fmt.Errorf("%w: %v", ErrDatabaseIssue, err)
	}
	defer rows.Close()

	lists := []*dao.AggregateTodoList{}
	listIds := []any{}
	for rows.Next() {
		var l dao.AggregateTodoList
		if err := rows.Scan(&l.Id, &l.Created, &l.Updated, &l.UserId, &l.Name, &l.Description, &l.TotalItems, &l.CompletedItems); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrDatabaseIssue, err)
		}
		l.Items = []dao.TodoItem{}
		listIds = append(listIds, l.Id)
		lists = append(lists, &l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseIssue, err)
	}
	if len(lists) == 0 {
		return lists, nil
	}

	itemsByList, err := r.getListItems(ctx, listIds)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseIssue, err)
	}

	for i := range lists {
		id := lists[i].Id
		lists[i].Items = itemsByList[id]
	}
	return lists, nil
}

func (r *TodoListRepository) CountByUser(ctx context.Context, userId string) (uint, error) {
	query := `
		SELECT COUNT(id)
		FROM lists
		WHERE user_id = ?
	`
	row := r.db.QueryRowContext(ctx, query, userId)

	var count uint
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("%w: %v", ErrDatabaseIssue, err)
	}

	return count, nil
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
		return nil, fmt.Errorf("%w: %v", ErrDatabaseIssue, err)
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
			return nil, fmt.Errorf("%w: %v", ErrDatabaseIssue, err)
		}
		if len(itemsByList[item.TodoListId]) < 10 {
			itemsByList[item.TodoListId] = append(itemsByList[item.TodoListId], item)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseIssue, err)
	}
	return itemsByList, nil
}
