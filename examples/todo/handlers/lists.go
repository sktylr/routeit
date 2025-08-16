package handlers

import (
	"errors"
	"strconv"
	"time"

	"github.com/sktylr/routeit"
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

type CreateListRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// TODO: should consolidate these!
type CreateListResponse struct {
	Id          string    `json:"id"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type GetListResponse struct {
	Id          string    `json:"id"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

func ListsMultiHandler(repo *db.TodoListRepository) routeit.Handler {
	queryParamOrDefault := func(param string, qs *routeit.QueryParams, def int) (int, error) {
		val, hasVal, err := qs.Only(param)
		if err != nil {
			return 0, err
		}

		valNum := def
		if hasVal {
			valNum, err = strconv.Atoi(val)
			if err != nil || valNum <= 0 {
				return 0, routeit.ErrBadRequest().
					WithMessagef("%#q is not a valid %s number", param, val).
					WithCause(err)
			}
		}
		return valNum, nil
	}

	return routeit.MultiMethod(routeit.MultiMethodHandler{
		Get: func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			userId, hasUser := userIdFromRequest(req)
			if !hasUser {
				// This shouldn't happen, but we add it as a fail safe
				return routeit.ErrUnauthorized()
			}

			page, err := queryParamOrDefault("page", req.Queries(), 1)
			if err != nil {
				return err
			}
			pageSize, err := queryParamOrDefault("page_size", req.Queries(), 10)
			if err != nil {
				return err
			}

			lists, err := repo.GetListsByUser(req.Context(), userId, page, pageSize)
			if err != nil {
				return routeit.ErrServiceUnavailable().WithCause(err)
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
			userId, hasUser := userIdFromRequest(req)
			if !hasUser {
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

func ListsIndividualHandler(repo *db.TodoListRepository) routeit.Handler {
	return routeit.MultiMethod(routeit.MultiMethodHandler{
		Get: func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			userId, hasUser := userIdFromRequest(req)
			if !hasUser {
				return routeit.ErrUnauthorized()
			}

			id, _ := req.PathParam("list")
			list, err := repo.GetListById(req.Context(), id)
			if err != nil {
				var nf db.ErrListNotFound
				if errors.As(err, &nf) {
					return routeit.ErrNotFound().WithCause(err)
				}
				return routeit.ErrServiceUnavailable().WithCause(err)
			}

			if list.UserId != userId {
				return routeit.ErrForbidden()
			}

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
			userId, hasUser := userIdFromRequest(req)
			if !hasUser {
				return routeit.ErrUnauthorized()
			}

			id, _ := req.PathParam("list")
			list, err := repo.GetListById(req.Context(), id)
			if err != nil {
				var nf db.ErrListNotFound
				if errors.As(err, &nf) {
					return routeit.ErrNotFound().WithCause(err)
				}
				return routeit.ErrServiceUnavailable().WithCause(err)
			}

			if list.UserId != userId {
				return routeit.ErrForbidden()
			}

			// TODO: in future, this error mapping should be handled at the top level. Unfortunately the tests would all need to be reworked since they depend on the function returning a HttpError.
			err = repo.DeleteList(req.Context(), list.Id)
			if err != nil {
				var nf db.ErrListNotFound
				if errors.As(err, &nf) {
					return routeit.ErrNotFound().WithCause(err)
				}
				return routeit.ErrServiceUnavailable().WithCause(err)
			}
			return nil
		},
	})
}
