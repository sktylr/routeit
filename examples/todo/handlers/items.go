package handlers

import (
	"time"

	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/dao"
	"github.com/sktylr/routeit/examples/todo/db"
)

type GetItemResponse itemResponse

type UpdateItemNameRequest itemRequest

type UpdateItemStatusRequest struct {
	Status string `json:"status"`
}

type CreateItemRequest itemRequest

type CreateItemResponse itemResponse

type ListItemsResponse struct {
	Items []itemResponse `json:"items"`
}

type itemRequest struct {
	Name string `json:"name"`
}

type itemResponse struct {
	Id      string    `json:"id"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
	Name    string    `json:"name"`
	Status  string    `json:"status"`
}

func ItemsIndividualHandler(repo *db.TodoItemRepository) routeit.Handler {
	return routeit.MultiMethod(routeit.MultiMethodHandler{
		Get: func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			item, _ := routeit.ContextValueAs[*dao.TodoItem](req, "item")
			res := GetItemResponse{
				Id:      item.Id,
				Created: item.Created,
				Updated: item.Updated,
				Name:    item.Name,
				Status:  item.Status,
			}
			return rw.Json(res)
		},
		Put: func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			var body UpdateItemNameRequest
			if err := req.BodyFromJson(&body); err != nil {
				return err
			}

			id, _ := req.PathParam("item")
			if err := repo.UpdateName(req.Context(), id, body.Name); err != nil {
				return err
			}

			return rw.Json(body)
		},
		Patch: func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			var body UpdateItemStatusRequest
			if err := req.BodyFromJson(&body); err != nil {
				return err
			}

			id, _ := req.PathParam("item")
			switch body.Status {
			case "COMPLETED":
				return repo.MarkAsCompleted(req.Context(), id)
			case "PENDING":
				return repo.MarkAsPending(req.Context(), id)
			default:
				return routeit.ErrBadRequest().WithMessagef("Unrecognised status %#q", body.Status)
			}
		},
		Delete: func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			id, _ := req.PathParam("item")
			return repo.DeleteItem(req.Context(), id)
		},
	})
}

func ItemsMultiHandler(repo *db.TodoItemRepository) routeit.Handler {
	return routeit.MultiMethod(routeit.MultiMethodHandler{
		Post: func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			listId, _ := req.PathParam("list")

			var body CreateItemRequest
			if err := req.BodyFromJson(&body); err != nil {
				return err
			}

			user := userIdFromRequest(req)
			item, err := repo.CreateItem(req.Context(), user, listId, body.Name)
			if err != nil {
				return err
			}

			res := CreateItemResponse{
				Id:      item.Id,
				Created: item.Created,
				Updated: item.Updated,
				Name:    item.Name,
				Status:  item.Status,
			}
			return rw.Json(res)
		},
		Get: func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			userId := userIdFromRequest(req)
			listId, _ := req.PathParam("list")

			pagination, err := extractPagination(req.Queries())
			if err != nil {
				return err
			}

			items, err := repo.GetByListAndUser(req.Context(), userId, listId, pagination.Page, pagination.PageSize)
			if err != nil {
				return err
			}

			res := ListItemsResponse{}
			for _, item := range items {
				resItem := itemResponse{
					Id:      item.Id,
					Created: item.Created,
					Updated: item.Updated,
					Name:    item.Name,
					Status:  item.Status,
				}
				res.Items = append(res.Items, resItem)
			}

			return rw.Json(res)
		},
	})
}
