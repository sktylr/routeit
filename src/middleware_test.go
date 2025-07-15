package routeit

import (
	"errors"
	"testing"
)

func TestChainProceed(t *testing.T) {
	errShouldHappen := errors.New("should happen")
	errShouldNotHappen := errors.New("should not happen")
	good := middlewareReturning(errShouldHappen)
	bad := middlewareReturning(errShouldNotHappen)
	tests := []struct {
		name    string
		m       middlewareRegistry
		start   uint
		wantErr error
	}{
		{
			name:    "empty chain",
			m:       middlewareRegistry{},
			wantErr: nil,
		},
		{
			name:    "singleton with error",
			m:       middlewareRegistry{good},
			wantErr: errShouldHappen,
		},
		{
			name:    "singleton at capacity",
			m:       middlewareRegistry{bad},
			start:   1,
			wantErr: nil,
		},
		{
			name:    "multiple, first returns error",
			m:       middlewareRegistry{good, bad},
			wantErr: errShouldHappen,
		},
		{
			name:    "multiple, middle returns error",
			m:       middlewareRegistry{bad, good, bad},
			start:   1,
			wantErr: errShouldHappen,
		},
		{
			name:    "multiple, last returns error",
			m:       middlewareRegistry{bad, bad, good},
			start:   2,
			wantErr: errShouldHappen,
		},
		{
			name:    "multiple at capacity",
			m:       middlewareRegistry{bad, bad, bad},
			start:   3,
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := newChain(&middlewareRegistry{})
			*c.reg = append(*c.reg, tc.m...)
			c.i = tc.start

			err := c.Proceed(nil, nil)

			if got, want := err, tc.wantErr; !errorsEqual(got, want) {
				t.Errorf("err = %v, want %v", got, want)
			}
		})
	}
}

func middlewareReturning(err error) Middleware {
	return func(c *Chain, rw *ResponseWriter, req *Request) error {
		return err
	}
}

func errorsEqual(a, b error) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Error() == b.Error()
}
