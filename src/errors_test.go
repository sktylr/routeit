package routeit

import "testing"

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

func newErrorHandlerNoMapper() *errorHandler {
	return newErrorHandler(func(e error) *HttpError { return nil })
}
