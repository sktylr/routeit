package main

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/sktylr/routeit"
)

var (
	ErrInvalid             = errors.New("this will map to an invalid HttpError and therefore a 500: Internal Server Error")
	ErrIncorrectPassword   = errors.New("password is incorrect")
	ErrMissingInformation  = errors.New("required information not provided")
	ErrPasswordNotProvided = fmt.Errorf("password not provided: %w", ErrMissingInformation)
	ErrUsernameNotProvided = fmt.Errorf("username not provided: %w", ErrMissingInformation)
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func GetServer() *routeit.Server {
	srv := routeit.NewServer(routeit.ServerConfig{
		ErrorMapper: ErrorMapper,
	})
	srv.RegisterRoutes(routeit.RouteRegistry{
		"/invalid": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			return ErrInvalid
		}),
		"/login": routeit.Post(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			var inBody LoginRequest
			err := req.BodyFromJson(&inBody)
			if err != nil {
				return err
			}

			if inBody.Username == "" {
				return ErrUsernameNotProvided
			}
			if inBody.Password == "" {
				return ErrPasswordNotProvided
			}
			if inBody.Password != "Password123!" {
				return ErrIncorrectPassword
			}

			out := LoginResponse{
				AccessToken:  "access_123",
				RefreshToken: "refresh_123",
			}
			rw.Status(routeit.StatusOK)
			return rw.Json(out)
		}),
		"/forbidden": routeit.Get(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
			// This handling is done by routeit implicitly - no need for
			// explicit handling by the integrator
			return fs.ErrPermission
		}),
	})
	return srv
}

func ErrorMapper(err error) *routeit.HttpError {
	if errors.Is(err, ErrInvalid) {
		// Although we technically return a [HttpError] here, it is invalid
		// since it cannot be interpreted by the server. This is because it has
		// been instantiated directly instead of using the appropriate
		// constructors. In reality, this will fall back to a 500: Internal
		// Server Error
		return &routeit.HttpError{}
	}

	if errors.Is(err, ErrIncorrectPassword) {
		return routeit.ErrBadRequest()
	}

	if errors.Is(err, ErrMissingInformation) {
		return routeit.ErrUnprocessableContent()
	}

	// Routeit will handle the rest of the error mapping using sensible
	// defaults
	return nil
}

func main() {
	GetServer().StartOrPanic()
}
