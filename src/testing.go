package routeit

import (
	"context"
	"testing"
)

// By declaring these variables here, we will fail builds if a method is added
// to the [testable] interface that is not implemented by one or both of
// [testing.T] or [testing.B].
var (
	_ testable = (*testing.T)(nil)
	_ testable = (*testing.B)(nil)
)

// The [testable] interface is an interface that contains signatures common
// between [testing.T] and [testing.B]. In doing this, we can accept either a
// [testing.T] or [testing.B] instance to the various helper methods exposed by
// routeit, meaning that the code is testable under unit tests and benchmarking
// conditions. The interface is non-exhaustive and only contains methods used
// within the library.
type testable interface {
	Context() context.Context
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	Helper()
}
