package main

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"time"

	"github.com/sktylr/routeit"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

func GetServer() *routeit.Server {
	srv := routeit.NewServer(routeit.ServerConfig{Debug: true, WriteDeadline: 2 * time.Second})
	srv.RegisterErrorHandlers(map[routeit.HttpStatus]routeit.ErrorResponseHandler{
		routeit.StatusUnauthorized: BaseErrorHandler("unauthorised", func(req *routeit.Request) string {
			return "Provide a valid access token"
		}),
		routeit.StatusNotFound: BaseErrorHandler("not_found", func(req *routeit.Request) string {
			return fmt.Sprintf("No matching route found for %s", req.Path())
		}),
		routeit.StatusInternalServerError: func(erw *routeit.ErrorResponseWriter, req *routeit.Request) {
			var sb strings.Builder
			sb.WriteString("An internal error has occurred. We are aware and are investigating. Please try again later or reach out support if it persists.")
			if err, hasErr := erw.Error(); hasErr {
				sb.WriteRune(' ')
				sb.WriteString(err.Error())
			}
			res := ErrorResponse{
				Error: ErrorDetail{
					Message: sb.String(),
					Code:    "internal_server_error",
				},
			}
			erw.Json(res)
		},
		routeit.StatusServiceUnavailable: func(erw *routeit.ErrorResponseWriter, req *routeit.Request) {
			res := ErrorResponse{
				Error: ErrorDetail{
					Message: "Our service is currently experiencing issues and is unavailable. Please try again in a few minutes.",
					Code:    "service_unavailable",
				},
			}
			erw.Json(res)
		},
	})
	srv.RegisterRoutes(routeit.RouteRegistry{
		"/no-auth": routeit.MultiMethod(routeit.MultiMethodHandler{
			Get: func(rw *routeit.ResponseWriter, req *routeit.Request) error {
				return routeit.ErrUnauthorized()
			},
			Post: func(rw *routeit.ResponseWriter, req *routeit.Request) error {
				return routeit.ErrUnauthorized()
			},
		}),
		"/crash": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			return routeit.ErrInternalServerError().WithCause(errors.New("uh oh we crashed"))
		}),
		"/panic": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			panic(errors.New("oops"))
		}),
		"/custom-error": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			return errors.New("this custom error will be mapped to a 500: Internal Server Error")
		}),
		"/manual-status": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			rw.Status(routeit.StatusInternalServerError)
			return nil
		}),
		"/bad-request": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			return routeit.ErrBadRequest()
		}),
		"/not-found": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			panic(fs.ErrNotExist)
		}),
		"/slow": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			time.Sleep(2*time.Second + 100*time.Millisecond)
			return nil
		}),
	})
	return srv
}

func BaseErrorHandler(code string, msg func(req *routeit.Request) string) routeit.ErrorResponseHandler {
	return func(erw *routeit.ErrorResponseWriter, req *routeit.Request) {
		res := ErrorResponse{
			Error: ErrorDetail{
				Message: msg(req),
				Code:    code,
			},
		}
		erw.Json(res)
	}
}

func main() {
	GetServer().StartOrPanic()
}
