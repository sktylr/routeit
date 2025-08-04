package main

import (
	"slices"
	"strings"

	"github.com/sktylr/routeit"
)

func GetServer() *routeit.Server {
	srv := routeit.NewServer(routeit.ServerConfig{Debug: true})
	srv.RegisterRoutes(routeit.RouteRegistry{
		"/scoped": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			// The middleware will ensure that only authenticated users with
			// the correct scopes will reach this endpoint.
			rw.Text(`You are authenticated and have the correct scopes: "fooscope:write", "fooscope:read" and "barscope".`)
			return nil
		}),
		"/scopeless": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			// This endpoint only requires authentication and does not need any scopes.
			scopes, _ := req.ContextValue("scopes")
			rw.Textf("You are authenticated and have the following scopes: %v", scopes)
			return nil
		}),
		"/no-auth": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			// This endpoint does not require authentication at all.
			rw.Text("You do not need to be authenticated to reach this endpoint!")
			return nil
		}),
		"/hello": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			scopes, _ := req.ContextValue("scopes")
			rw.Textf(`You need to be authenticated and have "barscopes" to reach this endpoint. You have %v scopes`, scopes)
			return nil
		}),
	})
	srv.RegisterMiddleware(AuthMiddleware, ScopesMiddleware)
	return srv
}

func AuthMiddleware(c routeit.Chain, rw *routeit.ResponseWriter, req *routeit.Request) error {
	if req.Path() == "/no-auth" {
		return c.Proceed(rw, req)
	}

	auth, hasAuth := req.Headers().First("Authorization")

	if !hasAuth || !strings.HasPrefix(auth, "Bearer ") {
		return routeit.ErrUnauthorized()
	}

	user := strings.TrimPrefix(auth, "Bearer ")
	req.NewContextValue("userId", user)
	req.NewContextValue("scopes", getScopes(user))

	return c.Proceed(rw, req)
}

func ScopesMiddleware(c routeit.Chain, rw *routeit.ResponseWriter, req *routeit.Request) error {
	if req.Path() == "/no-auth" {
		return c.Proceed(rw, req)
	}

	scopesRaw, hasScopes := req.ContextValue("scopes")

	if !hasScopes {
		return routeit.ErrForbidden()
	}

	scopes, ok := scopesRaw.([]string)
	if !ok {
		return routeit.ErrForbidden()
	}

	switch req.Path() {
	case "/scoped":
		if !slices.Contains(scopes, "fooscope:write") || !slices.Contains(scopes, "fooscope:read") {
			return routeit.ErrForbidden()
		}
	case "/scopeless":
	default:
		if !slices.Contains(scopes, "barscope") {
			return routeit.ErrForbidden()
		}
	}

	return c.Proceed(rw, req)
}

func getScopes(user string) []string {
	switch {
	case strings.HasPrefix(user, "user_"):
		return []string{"barscope"}
	case strings.HasPrefix(user, "superuser_"):
		return []string{"fooscope:write", "fooscope:read", "barscope"}
	default:
		return []string{}
	}
}

func main() {
	GetServer().StartOrPanic()
}
