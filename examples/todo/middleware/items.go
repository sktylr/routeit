package middleware

import (
	"strings"

	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/dao"
	"github.com/sktylr/routeit/examples/todo/db"
)

func LoadItemMiddleware(repo *db.TodoItemRepository) routeit.Middleware {
	return func(c routeit.Chain, rw *routeit.ResponseWriter, req *routeit.Request) error {
		path := req.Path()
		if !strings.HasPrefix(path, "/lists/") || !strings.Contains(path, "/items/") {
			return c.Proceed(rw, req)
		}

		user, userOk := routeit.ContextValueAs[*dao.User](req, "user")
		list, listOk := routeit.ContextValueAs[*dao.TodoList](req, "list")
		id := req.PathParam("item")

		item, err := repo.GetById(req.Context(), id)
		if err != nil {
			return err
		}

		if !listOk || item.TodoListId != list.Id {
			return routeit.ErrNotFound()
		}

		if !userOk || item.UserId != user.Id {
			return routeit.ErrForbidden()
		}

		req.NewContextValue("item", item)
		return c.Proceed(rw, req)
	}
}
