package errors

import (
	"errors"

	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/db"
)

func ErrorMapper(e error) *routeit.HttpError {
	switch {
	case errors.Is(e, db.ErrDuplicateKey):
		return routeit.ErrBadRequest().WithCause(e)
	case errors.Is(e, db.ErrItemNotFound), errors.Is(e, db.ErrListNotFound):
		return routeit.ErrNotFound().WithCause(e)
	case errors.Is(e, db.ErrDatabaseIssue):
		return routeit.ErrServiceUnavailable().WithCause(e)
	default:
		return nil
	}
}
