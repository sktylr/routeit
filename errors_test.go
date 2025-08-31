package routeit

import (
	"errors"
	"testing"
)

func TestRegisterHandler(t *testing.T) {
	t.Run("panics", func(t *testing.T) {
		tests := []struct {
			name string
			in   HttpStatus
		}{
			{
				name: "1xx",
				in:   StatusContinue,
			},
			{
				name: "2xx",
				in:   StatusOK,
			},
			{
				name: "3xx",
				in:   StatusPermanentRedirect,
			},
			{
				name: "no code",
				in:   HttpStatus{},
			},
			{
				name: "> 599",
				in:   HttpStatus{code: 600},
			},
		}
		eh := newErrorHandlerNoMapper()

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				defer func() {
					if r := recover(); r == nil {
						t.Error("wanted panic, got none")
					}
				}()

				eh.RegisterHandler(tc.in, func(erw *ErrorResponseWriter, req *Request) {})
			})
		}
	})

	t.Run("stores handler", func(t *testing.T) {
		tests := []struct {
			name string
			in   HttpStatus
		}{
			{
				name: "400",
				in:   StatusBadRequest,
			},
			{
				name: "4xx",
				in:   StatusNotFound,
			},
			{
				name: "500",
				in:   StatusInternalServerError,
			},
			{
				name: "5xx",
				in:   StatusBadGateway,
			},
			{
				// Although currently impossible due to how status codes must
				// be constructed, this ensures the validity check is sound.
				name: "599",
				in:   HttpStatus{code: 599},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				eh := newErrorHandlerNoMapper()
				want := "from inside test handler"

				eh.RegisterHandler(tc.in, func(erw *ErrorResponseWriter, req *Request) {
					erw.Text(want)
				})

				h, found := eh.handlers[tc.in]
				if !found {
					t.Error("expected handler to be found, was not")
				}
				erw := &ErrorResponseWriter{rw: newResponseWithStatus(tc.in)}
				h(erw, &Request{})
				if string(erw.rw.bdy) != want {
					t.Errorf("Body() = %#q, wanted %#q", string(erw.rw.bdy), want)
				}
			})
		}
	})
}

func TestIs(t *testing.T) {
	tests := []struct {
		name string
		in   error
		cmp  error
		want bool
	}{
		{
			name: "same status code",
			in:   ErrUnauthorized(),
			cmp:  ErrUnauthorized(),
			want: true,
		},
		{
			name: "same status code with different message",
			in:   ErrUnauthorized().WithMessage("foobar"),
			cmp:  ErrUnauthorized(),
			want: true,
		},
		{
			name: "same status code with different cause",
			in:   ErrUnauthorized().WithCause(errors.New("bad error")),
			cmp:  ErrUnauthorized(),
			want: true,
		},
		{
			name: "first nil",
			cmp:  ErrBadRequest(),
		},
		{
			name: "cmp nil",
			in:   ErrBadRequest(),
		},
		{
			name: "different status code",
			in:   ErrUnauthorized(),
			cmp:  ErrNotFound(),
		},
		{
			name: "different status code (manual construction)",
			in:   ErrUnauthorized(),
			cmp: func() error {
				e := ErrUnauthorized()
				e.status = StatusNotFound
				return e
			}(),
		},
		{
			name: "same status code, invalid < 100",
			in:   &HttpError{status: HttpStatus{code: 99}},
			cmp:  &HttpError{status: HttpStatus{code: 99}},
		},
		{
			name: "same status code, invalid >= 600",
			in:   &HttpError{status: HttpStatus{code: 600}},
			cmp:  &HttpError{status: HttpStatus{code: 600}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := errors.Is(tc.in, tc.cmp)
			if res != tc.want {
				t.Errorf(`in = %v, cmp = %v, res = %t, want = %t`, tc.in, tc.cmp, res, tc.want)
			}
		})
	}
}

func newErrorHandlerNoMapper() *errorHandler {
	return newErrorHandler(func(e error) *HttpError { return nil })
}
