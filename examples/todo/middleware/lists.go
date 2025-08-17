package middleware

import (
	"errors"
	"strings"

	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/dao"
	"github.com/sktylr/routeit/examples/todo/db"
)

// This middleware loads the corresponding list so it is available in the
// handler whenever the route involves looking up an individual list - e.g.
// /lists/my-list
func LoadListMiddleware(repo *db.TodoListRepository) routeit.Middleware {
	return func(c routeit.Chain, rw *routeit.ResponseWriter, req *routeit.Request) error {
		if !strings.HasPrefix(req.Path(), "/lists/") {
			return c.Proceed(rw, req)
		}

		user, ok := routeit.ContextValueAs[*dao.User](req, "user")
		id, _ := req.PathParam("list")
		list, err := repo.GetListById(req.Context(), id)
		if err != nil {
			var nf db.ErrListNotFound
			if errors.As(err, &nf) {
				return routeit.ErrNotFound().WithCause(err)
			}
			return routeit.ErrServiceUnavailable().WithCause(err)
		}

		if !ok || list.UserId != user.Id {
			return routeit.ErrForbidden()
		}

		req.NewContextValue("list", list)
		return c.Proceed(rw, req)
	}
}
