package middleware

import (
	"strings"

	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/auth"
	"github.com/sktylr/routeit/examples/todo/db"
)

func JwtMiddleware(repo *db.UsersRepository) routeit.Middleware {
	return func(c routeit.Chain, rw *routeit.ResponseWriter, req *routeit.Request) error {
		if strings.HasPrefix(req.Path(), "/auth/") {
			return c.Proceed(rw, req)
		}

		// We can safely extract the first Authorization header, since routeit
		// will ensure that at most 1 Authorization header appears in the
		// request
		authT, hasAuth := req.Headers().First("Authorization")
		if !hasAuth || !strings.HasPrefix(authT, "Bearer ") {
			return routeit.ErrUnauthorized()
		}

		claims, err := auth.ParseAccessToken(authT[7:])
		if err != nil {
			return routeit.ErrUnauthorized().WithCause(err)
		}
		if claims.IsExpired() {
			return routeit.ErrUnauthorized()
		}

		user, found, err := repo.GetUserById(req.Context(), claims.Subject)
		if !found {
			return routeit.ErrUnauthorized().WithCause(routeit.ErrNotFound())
		}
		if err != nil {
			return routeit.ErrUnauthorized().WithCause(err)
		}

		req.NewContextValue("user", user)
		return c.Proceed(rw, req)
	}
}
