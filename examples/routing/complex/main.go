package main

import "github.com/sktylr/routeit"

type HelloResponse struct {
	IncomingUrl  string `json:"incoming_url"`
	HandlerRoute string `json:"handler_route"`
	PathParam    string `json:"path_param"`
}

func GetServer() *routeit.Server {
	srv := routeit.NewServer(routeit.ServerConfig{Debug: true})
	rreg := routeit.RouteRegistry{}
	routes := []string{
		// The most general route, this will match against /<anything>, but has
		// the lowest priority of anything coming in on that route.
		"/:path",
		// This route has equal specificity as /:path|suffix, which is more
		// specific than /:path, but less specific than /:path|prefix|suffix
		"/:path|prefix",
		"/:path||suffix",
		// This is the most specific of the above matchers, and requires that
		// both the prefix and suffix are included, plus at least 1 character
		// in between them.
		"/:path|prefix|suffix",
	}
	for _, r := range routes {
		rreg[r] = Get(r)
	}
	srv.RegisterRoutes(rreg)
	return srv
}

func Get(route string) routeit.Handler {
	return routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
		param := req.PathParam("path")
		out := HelloResponse{
			IncomingUrl:  req.Path(),
			HandlerRoute: route,
			PathParam:    param,
		}

		return rw.Json(out)
	})
}

func main() {
	GetServer().StartOrPanic()
}
