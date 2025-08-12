package db

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/sktylr/routeit/examples/todo/dao"
)

func TestCreateList(t *testing.T) {
	tests := []struct {
		name      string
		userId    string
		listName  string
		desc      string
		mockSetup func(sqlmock.Sqlmock)
		expectErr bool
	}{
		{
			name:     "success",
			userId:   "user-123",
			listName: "Groceries",
			desc:     "Things to buy this week",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`INSERT INTO lists`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						"user-123", "Groceries", "Things to buy this week").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name:     "db error",
			userId:   "user-456",
			listName: "Chores",
			desc:     "Weekend cleaning tasks",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`INSERT INTO lists`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						"user-456", "Chores", "Weekend cleaning tasks").
					WillReturnError(errors.New("insert failed"))
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			WithUnitTestConnection(t, func(dbConn *sql.DB, mock sqlmock.Sqlmock) {
				tc.mockSetup(mock)
				repo := NewTodoListRepository(dbConn)

				list, err := repo.CreateList(t.Context(), tc.userId, tc.listName, tc.desc)

				if tc.expectErr && err == nil {
					t.Errorf("expected error, got nil")
				}
				if !tc.expectErr && err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !tc.expectErr {
					if list.UserId != tc.userId || list.Name != tc.listName || list.Description != tc.desc {
						t.Errorf("unexpected list: %+v", list)
					}
					if list.Created.IsZero() || list.Updated.IsZero() || list.Id == "" {
						t.Errorf("timestamps or id not set properly: %+v", list)
					}
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unmet SQL expectations: %v", err)
				}
			})
		})
	}
}

func TestUpdateList(t *testing.T) {
	tests := []struct {
		name      string
		listId    string
		newName   string
		newDesc   string
		mockSetup func(sqlmock.Sqlmock)
		expectErr bool
		errCheck  func(error) bool
	}{
		{
			name:    "success",
			listId:  "list-123",
			newName: "Updated Groceries",
			newDesc: "Updated description",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE lists`).
					WithArgs("Updated Groceries", "Updated description", sqlmock.AnyArg(), "list-123").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:    "list not found",
			listId:  "missing-456",
			newName: "Doesn't matter",
			newDesc: "Nope",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE lists`).
					WithArgs("Doesn't matter", "Nope", sqlmock.AnyArg(), "missing-456").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectErr: true,
			errCheck: func(err error) bool {
				var notFound ErrListNotFound
				return errors.As(err, &notFound) && notFound.listId == "missing-456"
			},
		},
		{
			name:    "db error",
			listId:  "list-999",
			newName: "Broken",
			newDesc: "DB error",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE lists`).
					WithArgs("Broken", "DB error", sqlmock.AnyArg(), "list-999").
					WillReturnError(errors.New("update failed"))
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			WithUnitTestConnection(t, func(dbConn *sql.DB, mock sqlmock.Sqlmock) {
				tc.mockSetup(mock)
				repo := NewTodoListRepository(dbConn)

				list, err := repo.UpdateList(t.Context(), tc.listId, tc.newName, tc.newDesc)

				if tc.expectErr {
					if err == nil {
						t.Errorf("expected error, got nil")
					} else if tc.errCheck != nil && !tc.errCheck(err) {
						t.Errorf("error check failed: %v", err)
					}
				} else {
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}
					if list.Name != tc.newName || list.Description != tc.newDesc {
						t.Errorf("unexpected list values: %+v", list)
					}
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unmet SQL expectations: %v", err)
				}
			})
		})
	}
}

func TestGetListsByUser(t *testing.T) {
	type listRow struct {
		id             string
		created        time.Time
		updated        time.Time
		userId         string
		name           string
		description    string
		totalItems     int64
		completedItems int64
	}
	type itemRow struct {
		id      string
		created time.Time
		updated time.Time
		userId  string
		listId  string
		name    string
		status  string
	}
	now := time.Now()
	tests := []struct {
		name          string
		page          int
		pageSize      int
		userId        string
		mockListRows  []listRow
		mockItemRows  []itemRow
		expectedLists []dao.AggregateTodoList
	}{
		{
			name:     "no lists",
			page:     1,
			pageSize: 10,
			userId:   "user1",
		},
		{
			name:     "one list no items",
			page:     1,
			pageSize: 10,
			userId:   "user1",
			mockListRows: []listRow{
				{
					id: "list1", created: now, updated: now,
					userId: "user1", name: "List One", description: "Desc",
					totalItems: 0, completedItems: 0,
				},
			},
			expectedLists: []dao.AggregateTodoList{
				{
					TodoList: dao.TodoList{
						Meta: dao.Meta{
							Id:      "list1",
							Created: now,
							Updated: now,
						},
						UserId:      "user1",
						Name:        "List One",
						Description: "Desc",
					},
					TotalItems:     0,
					CompletedItems: 0,
				},
			},
		},
		{
			name:     "one list with <=10 items",
			page:     1,
			pageSize: 10,
			userId:   "user1",
			mockListRows: []listRow{
				{
					id: "list1", created: now, updated: now,
					userId: "user1", name: "List One", description: "Desc",
					totalItems: 2, completedItems: 1,
				},
			},
			mockItemRows: []itemRow{
				{
					id: "item1", created: now, updated: now,
					userId: "user1", listId: "list1", name: "Task 1", status: "PENDING",
				},
				{
					id: "item2", created: now.Add(time.Second), updated: now.Add(time.Second),
					userId: "user1", listId: "list1", name: "Task 2", status: "COMPLETED",
				},
			},
			expectedLists: []dao.AggregateTodoList{
				{
					TodoList: dao.TodoList{
						Meta: dao.Meta{
							Id:      "list1",
							Created: now,
							Updated: now,
						},
						UserId:      "user1",
						Name:        "List One",
						Description: "Desc",
					},
					Items: []dao.TodoItem{
						{
							Meta:   dao.Meta{Id: "item1", Created: now, Updated: now},
							UserId: "user1", TodoListId: "list1", Name: "Task 1", Status: "PENDING",
						},
						{
							Meta:   dao.Meta{Id: "item2", Created: now.Add(time.Second), Updated: now.Add(time.Second)},
							UserId: "user1", TodoListId: "list1", Name: "Task 2", Status: "COMPLETED",
						},
					},
					TotalItems:     2,
					CompletedItems: 1,
				},
			},
		},
		{
			name:     "one list with >10 items truncates",
			page:     1,
			pageSize: 10,
			userId:   "user1",
			mockListRows: []listRow{
				{
					id: "list1", created: now, updated: now,
					userId: "user1", name: "Big List", description: "Lots of tasks",
					totalItems: 12, completedItems: 5,
				},
			},
			mockItemRows: func() []itemRow {
				items := []itemRow{}
				for i := 0; i < 12; i++ {
					items = append(items, itemRow{
						id:      uuid.NewString(),
						created: now.Add(time.Duration(i) * time.Second),
						updated: now.Add(time.Duration(i) * time.Second),
						userId:  "user1", listId: "list1",
						name: "Task", status: "PENDING",
					})
				}
				return items
			}(),
			expectedLists: func() []dao.AggregateTodoList {
				expected := []dao.TodoItem{}
				for i := 0; i < 10; i++ {
					expected = append(expected, dao.TodoItem{
						Meta: dao.Meta{
							Created: now.Add(time.Duration(i) * time.Second),
							Updated: now.Add(time.Duration(i) * time.Second),
						},
						UserId:     "user1",
						TodoListId: "list1",
						Name:       "Task",
						Status:     "PENDING",
					})
				}
				return []dao.AggregateTodoList{
					{
						TodoList: dao.TodoList{
							Meta: dao.Meta{
								Id:      "list1",
								Created: now,
								Updated: now,
							},
							UserId:      "user1",
							Name:        "Big List",
							Description: "Lots of tasks",
						},
						Items:          expected,
						TotalItems:     12,
						CompletedItems: 5,
					},
				}
			}(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			WithUnitTestConnection(t, func(db *sql.DB, mock sqlmock.Sqlmock) {
				r := NewTodoListRepository(db)

				listQuery := `
					SELECT l.id, l.created, l.updated, l.user_id, l.name, l.description,
					       COUNT\(i.id\) AS total_items,
					       SUM\(CASE WHEN i.status = 'COMPLETED' THEN 1 ELSE 0 END\) AS completed_items
					FROM lists l
					LEFT JOIN items i ON l.id = i.list_id
					WHERE l.user_id = \?
					GROUP BY l.id
					ORDER BY l.created DESC
					LIMIT \? OFFSET \?
				`
				listRows := mock.NewRows([]string{
					"id", "created", "updated", "user_id", "name", "description", "total_items", "completed_items",
				})
				for _, lr := range tc.mockListRows {
					listRows.AddRow(lr.id, lr.created, lr.updated, lr.userId, lr.name, lr.description, lr.totalItems, lr.completedItems)
				}
				mock.ExpectQuery(listQuery).WithArgs(tc.userId, tc.pageSize, (tc.page-1)*tc.pageSize).WillReturnRows(listRows)

				if len(tc.mockListRows) > 0 {
					itemQuery := `SELECT id, created, updated, user_id, list_id, name, status
						FROM items
						WHERE list_id IN \(.+\)
						ORDER BY list_id, created ASC`
					itemRows := mock.NewRows([]string{
						"id", "created", "updated", "user_id", "list_id", "name", "status",
					})
					for _, ir := range tc.mockItemRows {
						itemRows.AddRow(ir.id, ir.created, ir.updated, ir.userId, ir.listId, ir.name, ir.status)
					}
					mock.ExpectQuery(itemQuery).WillReturnRows(itemRows)
				}

				got, err := r.GetListsByUser(t.Context(), tc.userId, tc.page, tc.pageSize)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if len(got) != len(tc.expectedLists) {
					t.Errorf("expected %d lists, got %d", len(tc.expectedLists), len(got))
				}

				if tc.name == "one list with >10 items truncates" {
					if len(got[0].Items) != 10 {
						t.Errorf("expected 10 items, got %d", len(got[0].Items))
					}
					if got[0].TotalItems <= 10 {
						t.Errorf("expected totalItems > 10, got %d", got[0].TotalItems)
					}
				} else {
					for i := range tc.expectedLists {
						if got[i].Id != tc.expectedLists[i].Id ||
							got[i].Name != tc.expectedLists[i].Name ||
							got[i].Description != tc.expectedLists[i].Description ||
							got[i].TotalItems != tc.expectedLists[i].TotalItems ||
							got[i].CompletedItems != tc.expectedLists[i].CompletedItems {
							t.Errorf("mismatch in list %d: got %+v, expected %+v", i, got[i], tc.expectedLists[i])
						}
					}
				}

				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("there were unfulfilled expectations: %v", err)
				}
			})
		})
	}
}

func TestGetListById(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name         string
		listId       string
		setupMock    func(sqlmock.Sqlmock)
		wantErr      bool
		expectedList *dao.TodoList
	}{
		{
			name:   "list found",
			listId: "list1",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`
					SELECT id, created, updated, user_id, name, description
					FROM lists
					WHERE id = \?
				`).WithArgs("list1").
					WillReturnRows(
						sqlmock.NewRows([]string{
							"id", "created", "updated", "user_id", "name", "description",
						}).AddRow(
							"list1", now, now, "user1", "Groceries", "Things to buy",
						),
					)
			},
			expectedList: &dao.TodoList{
				Meta: dao.Meta{
					Id:      "list1",
					Created: now,
					Updated: now,
				},
				UserId:      "user1",
				Name:        "Groceries",
				Description: "Things to buy",
			},
		},
		{
			name:   "list not found",
			listId: "missing-list",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`
					SELECT id, created, updated, user_id, name, description
					FROM lists
					WHERE id = \?
				`).WithArgs("missing-list").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name:   "unexpected db error",
			listId: "bad-query",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`
					SELECT id, created, updated, user_id, name, description
					FROM lists
					WHERE id = \?
				`).WithArgs("bad-query").
					WillReturnError(errors.New("db connection failed"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			WithUnitTestConnection(t, func(db *sql.DB, mock sqlmock.Sqlmock) {
				repo := NewTodoListRepository(db)
				tc.setupMock(mock)

				got, err := repo.GetListById(t.Context(), tc.listId)

				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unmet mock expectations: %v", err)
				}
				if tc.wantErr {
					if err == nil {
						t.Fatal("expected error got nil")
					}
					if got != nil {
						t.Errorf("expected no list, got %+v", got)
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got == nil {
					t.Fatalf("expected a list, got nil")
				}
				if got.Id != tc.expectedList.Id ||
					!got.Created.Equal(tc.expectedList.Created) ||
					!got.Updated.Equal(tc.expectedList.Updated) ||
					got.UserId != tc.expectedList.UserId ||
					got.Name != tc.expectedList.Name ||
					got.Description != tc.expectedList.Description {
					t.Errorf("got %+v, expected %+v", got, tc.expectedList)
				}
			})
		})
	}
}

func TestDeleteList(t *testing.T) {
	tests := []struct {
		name      string
		listId    string
		setupMock func(sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:   "list deleted successfully",
			listId: "list1",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM lists WHERE id = \?`).
					WithArgs("list1").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:   "list not found",
			listId: "missing-list",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM lists WHERE id = \?`).
					WithArgs("missing-list").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name:   "db error",
			listId: "list1",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM lists WHERE id = \?`).
					WithArgs("list1").
					WillReturnError(errors.New("connection lost"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			WithUnitTestConnection(t, func(db *sql.DB, mock sqlmock.Sqlmock) {
				repo := NewTodoListRepository(db)
				tc.setupMock(mock)

				err := repo.DeleteList(t.Context(), tc.listId)

				if tc.wantErr {
					if err == nil {
						t.Fatal("expected error got nil")
					}
				} else {
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unmet mock expectations: %v", err)
				}
			})
		})
	}
}
