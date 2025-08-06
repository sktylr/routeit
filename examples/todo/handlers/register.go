package handlers

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/auth"
	"github.com/sktylr/routeit/examples/todo/db"
)

// This is an oversimplified regex that will match basic emails. For our
// example use case it is sufficient
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type RegisterUserRequest struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type RegisterUserResponse tokenResponse

func RegisterUserHandler(repo *db.UsersRepository) routeit.Handler {
	return routeit.Post(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
		var input RegisterUserRequest
		err := req.BodyFromJson(&input)
		if err != nil {
			return err
		}

		if input.Name == "" || input.Email == "" || input.Password == "" {
			return routeit.ErrUnprocessableContent()
		}

		if !emailRegex.MatchString(input.Email) {
			return routeit.ErrBadRequest().WithMessage("Invalid email address format.")
		}

		if input.Password != input.ConfirmPassword {
			return routeit.ErrBadRequest().WithMessage("Password does not match confirm password.")
		}

		user, err := repo.CreateUser(req.Context(), input.Name, input.Email, input.Password)
		if err != nil {
			fmt.Println(err)
			if errors.Is(err, db.ErrDuplicateKey) {
				// For now, we don't report this specifically back to the user,
				// since it's generally not a good idea to indicate that a user
				// already exists with the given email address as this could be
				// a security risk (if the request is coming from a bad actor).
				return routeit.ErrBadRequest()
			}
			return routeit.ErrServiceUnavailable().WithCause(err)
		}

		tokens, err := auth.GenerateTokens(user.Id)
		if err != nil {
			return err
		}

		response := RegisterUserResponse{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		}
		return rw.Json(response)
	})
}
