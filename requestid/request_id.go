package requestid

import (
	"github.com/google/uuid"
	"github.com/sktylr/routeit"
)

// Returns a [routeit.RequestIdProvider] that returns a UUIDv7 ID for each
// request. UUIDv7 ID's are guaranteed to be ordered by creation, meaning
// request ID's can be traced more accurately and easily to understand the
// order of request receipt. In the case where the ID cannot be generated, an
// empty string is returned as we do not wish to block the request lifecycle
// purely because a debug identifier cannot be attached to the request.
func NewUuidV7Provider() routeit.RequestIdProvider {
	return func(r *routeit.Request) string {
		id, err := uuid.NewV7()
		if err != nil {
			// Request IDs are nice to have, but not essential. We shouldn't
			// block the request just because we couldn't generate an ID for it
			return ""
		}

		return id.String()
	}
}
