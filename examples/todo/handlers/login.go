package handlers

import (
	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/auth"
	"github.com/sktylr/routeit/examples/todo/db"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse tokenResponse

func LoginHandler(repo *db.UsersRepository) routeit.Handler {
	return routeit.Post(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
		var input LoginRequest
		err := req.BodyFromJson(&input)
		if err != nil {
			return err
		}

		if input.Email == "" || input.Password == "" {
			return routeit.ErrUnprocessableContent()
		}

		user, found, err := repo.GetUserByEmail(req.Context(), input.Email)
		if !found && err == nil {
			return routeit.ErrNotFound()
		}
		if err != nil {
			return routeit.ErrServiceUnavailable().WithCause(err)
		}

		if !auth.ComparePassword(user.Password, input.Password) {
			return routeit.ErrBadRequest()
		}

		tokens, err := auth.GenerateTokens(user.Id)
		if err != nil {
			return err
		}

		response := LoginResponse{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		}
		return rw.Json(response)
	})
}
