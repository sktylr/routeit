package handlers

import (
	"time"

	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/dao"
	"github.com/sktylr/routeit/examples/todo/db"
)

// TODO: could extend this to include metadata about the total number of lists etc.
type ListListsResponse struct {
	Lists []NestedListResponse `json:"lists"`
}

type NestedListResponse struct {
	Id             string                   `json:"id"`
	Created        time.Time                `json:"created"`
	Updated        time.Time                `json:"updated"`
	Name           string                   `json:"name"`
	Description    string                   `json:"description"`
	Items          []NestedListItemResponse `json:"items"`
	TotalItems     int                      `json:"total_items"`
	CompletedItems int                      `json:"completed_items"`
}

type NestedListItemResponse struct {
	Id      string    `json:"id"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
	Name    string    `json:"name"`
	Status  string    `json:"status"`
}

type CreateListRequest listRequest

type CreateListResponse listResponse

type GetListResponse listResponse

type UpdateListRequest listRequest

type UpdateListResponse listResponse

type listRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type listResponse struct {
	Id          string    `json:"id"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

func ListsMultiHandler(repo *db.TodoListRepository) routeit.Handler {
	return routeit.MultiMethod(routeit.MultiMethodHandler{
		Get: func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			userId := userIdFromRequest(req)

			pagination, err := extractPagination(req.Queries())
			if err != nil {
				return err
			}

			lists, err := repo.GetListsByUser(req.Context(), userId, pagination.Page, pagination.PageSize)
			if err != nil {
				return err
			}

			var res []NestedListResponse
			for _, l := range lists {
				var items []NestedListItemResponse
				for _, i := range l.Items {
					items = append(items, NestedListItemResponse{
						Id:      i.Id,
						Created: i.Created,
						Updated: i.Updated,
						Name:    i.Name,
						Status:  i.Status,
					})
				}
				res = append(res, NestedListResponse{
					Id:             l.Id,
					Created:        l.Created,
					Updated:        l.Updated,
					Name:           l.Name,
					Description:    l.Description,
					Items:          items,
					TotalItems:     l.TotalItems,
					CompletedItems: l.CompletedItems,
				})
			}

			return rw.Json(ListListsResponse{Lists: res})
		},
		Post: func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			userId := userIdFromRequest(req)

			var body CreateListRequest
			if err := req.BodyFromJson(&body); err != nil {
				return err
			}

			if body.Name == "" {
				return routeit.ErrBadRequest().WithMessage("Name must not be empty.")
			}

			list, err := repo.CreateList(req.Context(), userId, body.Name, body.Description)
			if err != nil {
				return err
			}

			res := CreateListResponse{
				Id:          list.Id,
				Created:     list.Created,
				Updated:     list.Updated,
				Name:        list.Name,
				Description: list.Description,
			}
			return rw.Json(res)
		},
	})
}

func ListsIndividualHandler(repo *db.TodoListRepository) routeit.Handler {
	return routeit.MultiMethod(routeit.MultiMethodHandler{
		Get: func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			list, _ := routeit.ContextValueAs[*dao.TodoList](req, "list")
			res := GetListResponse{
				Id:          list.Id,
				Created:     list.Created,
				Updated:     list.Updated,
				Name:        list.Name,
				Description: list.Description,
			}
			return rw.Json(res)
		},
		Delete: func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			id, _ := req.PathParam("list")
			return repo.DeleteList(req.Context(), id)
		},
		Put: func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			list, _ := routeit.ContextValueAs[*dao.TodoList](req, "list")

			var body UpdateListRequest
			if err := req.BodyFromJson(&body); err != nil {
				return err
			}

			updated, err := repo.UpdateList(req.Context(), list.Id, body.Name, body.Description)
			if err != nil {
				return err
			}

			res := UpdateListResponse{
				Id:          list.Id,
				Created:     list.Created,
				Updated:     updated.Updated,
				Name:        updated.Name,
				Description: updated.Description,
			}
			return rw.Json(res)
		},
	})
}
