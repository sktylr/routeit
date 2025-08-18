package handlers

import (
	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/auth"
	"github.com/sktylr/routeit/examples/todo/db"
)

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse tokenResponse

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func RefreshTokenHandler(repo *db.UsersRepository) routeit.Handler {
	return routeit.Post(func(rw *routeit.ResponseWriter, req *routeit.Request) error {
		var input RefreshTokenRequest
		if err := req.BodyFromJson(&input); err != nil {
			return err
		}

		if input.RefreshToken == "" {
			return routeit.ErrUnprocessableContent()
		}

		claims, err := auth.ParseRefreshToken(input.RefreshToken)
		if err != nil {
			return routeit.ErrUnauthorized().WithCause(err)
		}
		if claims.IsExpired() {
			return routeit.ErrUnauthorized()
		}

		// As a sanity check, make sure the refresh token is for a valid user.
		user, found, err := repo.GetUserById(req.Context(), claims.Subject)
		if !found && err == nil {
			return routeit.ErrUnauthorized()
		}
		if err != nil {
			return err
		}

		// A stronger implementation would also invalidate the current refresh
		// token, to ensure that it cannot be distributed to many different
		// clients and used throughout its lifetime. For our purposes, it is
		// sufficient to just verify the token and generate a new (access,
		// refresh) pair.
		tokens, err := auth.GenerateTokens(user.Id)
		if err != nil {
			return err
		}

		response := RefreshTokenResponse{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		}
		return rw.Json(response)
	})
}
