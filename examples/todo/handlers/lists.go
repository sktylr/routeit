package handlers

import (
	"time"

	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/db"
)

type CreateListRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateListResponse struct {
	Id          string    `json:"id"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

func ListsMultiHandler(repo *db.TodoListRepository) routeit.Handler {
	return routeit.MultiMethod(routeit.MultiMethodHandler{
		Get: func(rw *routeit.ResponseWriter, req *routeit.Request) error { return nil },
		Post: func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			userId, hasUser := userIdFromRequest(req)
			if !hasUser {
				// This shouldn't happen, but we add it as a fail safe
				return routeit.ErrUnauthorized()
			}

			var body CreateListRequest
			if err := req.BodyFromJson(&body); err != nil {
				return routeit.ErrBadRequest().WithCause(err)
			}

			if body.Name == "" {
				return routeit.ErrBadRequest().WithMessage("Name must not be empty.")
			}

			list, err := repo.CreateList(req.Context(), userId, body.Name, body.Description)
			if err != nil {
				return routeit.ErrServiceUnavailable().WithCause(err)
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
