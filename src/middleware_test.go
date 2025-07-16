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
		mwares  []Middleware
		last    Middleware
		start   uint
		wantErr error
	}{
		{
			name:    "empty chain, last only with error",
			mwares:  []Middleware{},
			last:    good,
			start:   0,
			wantErr: errShouldHappen,
		},
		{
			name:    "empty, over capacity",
			mwares:  []Middleware{},
			last:    bad,
			start:   1,
			wantErr: nil,
		},
		{
			name:    "singleton with error",
			mwares:  []Middleware{good},
			last:    bad,
			wantErr: errShouldHappen,
			start:   0,
		},
		{
			name:    "singleton chain, at last",
			mwares:  []Middleware{bad},
			last:    good,
			wantErr: errShouldHappen,
			start:   1,
		},
		{
			name:    "singleton chain, at first",
			mwares:  []Middleware{good},
			last:    bad,
			wantErr: errShouldHappen,
			start:   0,
		},
		{
			name:    "singleton, over capacity",
			mwares:  []Middleware{bad},
			last:    bad,
			wantErr: nil,
			start:   2,
		},
		{
			name:    "multiple, first returns error",
			mwares:  []Middleware{good, bad},
			last:    bad,
			start:   0,
			wantErr: errShouldHappen,
		},
		{
			name:    "multiple, middle returns error",
			mwares:  []Middleware{bad, good},
			last:    bad,
			start:   1,
			wantErr: errShouldHappen,
		},
		{
			name:    "multiple, last returns error",
			mwares:  []Middleware{bad, bad},
			last:    good,
			start:   2,
			wantErr: errShouldHappen,
		},
		{
			name:    "multiple at capacity",
			mwares:  []Middleware{bad, bad},
			last:    bad,
			start:   3,
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mware := newMiddleware(tc.last)
			mware.Register(tc.mwares...)
			c := mware.NewChain()
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
