package routeit

import (
	"testing"
)

func TestGet(t *testing.T) {
	h := Get(func(rw *ResponseWriter, req *Request) error { return nil })

	if h.get == nil {
		t.Error("did not expect handler.get() to be nil")
	}
	if h.head == nil {
		t.Error("did not expect handler.head() to be nil")
	}
	if h.post != nil {
		t.Error("expected handler.post() to be nil")
	}
	if h.put != nil {
		t.Error("expected handler.put() to be nil")
	}
	if h.delete != nil {
		t.Error("expected handler.delete() to be nil")
	}
	if h.options == nil {
		t.Error("did not expect handler.options() to be nil")
	}
}

func TestPost(t *testing.T) {
	h := Post(func(rw *ResponseWriter, req *Request) error { return nil })

	if h.post == nil {
		t.Error("did not expect handler.post() to be nil")
	}
	if h.get != nil {
		t.Error("expected handler.get() to be nil")
	}
	if h.head != nil {
		t.Error("expected handler.head() to be nil")
	}
	if h.put != nil {
		t.Error("expected handler.put() to be nil")
	}
	if h.delete != nil {
		t.Error("expected handler.delete() to be nil")
	}
	if h.options == nil {
		t.Error("did not expect handler.options() to be nil")
	}
}

func TestPut(t *testing.T) {
	h := Put(func(rw *ResponseWriter, req *Request) error { return nil })

	if h.put == nil {
		t.Error("did not expect handler.put() to be nil")
	}
	if h.get != nil {
		t.Error("expected handler.get() to be nil")
	}
	if h.head != nil {
		t.Error("expected handler.head() to be nil")
	}
	if h.post != nil {
		t.Error("expected handler.post() to be nil")
	}
	if h.delete != nil {
		t.Error("expected handler.delete() to be nil")
	}
	if h.options == nil {
		t.Error("did not expect handler.options() to be nil")
	}
}

func TestDelete(t *testing.T) {
	h := Delete(func(rw *ResponseWriter, req *Request) error { return nil })

	if h.delete == nil {
		t.Error("did not expect handler.delete() to be nil")
	}
	if h.get != nil {
		t.Error("expected handler.get() to be nil")
	}
	if h.head != nil {
		t.Error("expected handler.head() to be nil")
	}
	if h.post != nil {
		t.Error("expected handler.post() to be nil")
	}
	if h.put != nil {
		t.Error("expected handler.put() to be nil")
	}
	if h.options == nil {
		t.Error("did not expect handler.options() to be nil")
	}
}

func TestMultiMethodOnlyGet(t *testing.T) {
	h := MultiMethod(MultiMethodHandler{
		Get: func(rw *ResponseWriter, req *Request) error { return nil },
	})

	if h.get == nil {
		t.Error("did not expect handler.get() to be nil")
	}
	if h.head == nil {
		t.Error("did not expect handler.head() to be nil")
	}
	if h.post != nil {
		t.Error("expected handler.post() to be nil")
	}
	if h.put != nil {
		t.Error("expected handler.put() to be nil")
	}
	if h.delete != nil {
		t.Error("expected handler.delete() to be nil")
	}
	if h.options == nil {
		t.Error("did not expect handler.options() to be nil")
	}
}

func TestMultiMethodOnlyPost(t *testing.T) {
	h := MultiMethod(MultiMethodHandler{
		Post: func(rw *ResponseWriter, req *Request) error { return nil },
	})

	if h.post == nil {
		t.Error("did not expect handler.post() to be nil")
	}
	if h.get != nil {
		t.Error("expected handler.get() to be nil")
	}
	if h.head != nil {
		t.Error("expected handler.head() to be nil")
	}
	if h.put != nil {
		t.Error("expected handler.put() to be nil")
	}
	if h.delete != nil {
		t.Error("expected handler.delete() to be nil")
	}
	if h.options == nil {
		t.Error("did not expect handler.options() to be nil")
	}
}

func TestMultiMethod(t *testing.T) {
	h := MultiMethod(MultiMethodHandler{
		Get:  func(rw *ResponseWriter, req *Request) error { return nil },
		Post: func(rw *ResponseWriter, req *Request) error { return nil },
	})

	if h.get == nil {
		t.Error("did not expect handler.get() to be nil")
	}
	if h.head == nil {
		t.Error("did not expect handler.head() to be nil")
	}
	if h.post == nil {
		t.Error("did not expect handler.post() to be nil")
	}
	if h.put != nil {
		t.Error("expected handler.put() to be nil")
	}
	if h.delete != nil {
		t.Error("expected handler.delete() to be nil")
	}
	if h.options == nil {
		t.Error("did not expect handler.options() to be nil")
	}
}

func TestHandle(t *testing.T) {
	fn := func(rw *ResponseWriter, req *Request) error {
		rw.Text("From inside the handler")
		return nil
	}
	tests := []struct {
		name       string
		method     HttpMethod
		h          Handler
		wantBody   string
		wantStatus HttpStatus
		wantCType  string
		wantCLen   uint
		wantAllow  string
		wantErr    bool
	}{
		{
			method:     GET,
			h:          Get(fn),
			wantBody:   "From inside the handler",
			wantStatus: StatusOK,
			wantCType:  "text/plain",
			wantCLen:   23,
		},
		{
			method:     HEAD,
			h:          Get(fn),
			wantBody:   "",
			wantStatus: StatusOK,
			wantCType:  "text/plain",
			wantCLen:   23,
		},
		{
			method:     POST,
			h:          Post(fn),
			wantBody:   "From inside the handler",
			wantStatus: StatusCreated,
			wantCType:  "text/plain",
			wantCLen:   23,
		},
		{
			method:     PUT,
			h:          Put(fn),
			wantBody:   "From inside the handler",
			wantStatus: StatusOK,
			wantCType:  "text/plain",
			wantCLen:   23,
		},
		{
			method:     DELETE,
			h:          Delete(fn),
			wantBody:   "From inside the handler",
			wantStatus: StatusNoContent,
			wantCType:  "text/plain",
			wantCLen:   23,
		},
		{
			method:     OPTIONS,
			h:          Get(fn),
			wantStatus: StatusNoContent,
			wantAllow:  "GET, HEAD, OPTIONS",
		},
		{
			name:       "unsupported method",
			method:     POST,
			h:          Get(fn),
			wantStatus: StatusMethodNotAllowed,
			wantAllow:  "GET, HEAD, OPTIONS",
			wantErr:    true,
		},
		{
			name:   "returns error",
			method: GET,
			h: Get(func(rw *ResponseWriter, req *Request) error {
				return ErrBadRequest()
			}),
			wantStatus: StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		name := tc.name
		if name == "" {
			name = tc.method.name
		}
		t.Run(name, func(t *testing.T) {
			req := requestWithUrlAndMethod("/foo", tc.method)
			rw := newResponseForMethod(tc.method)

			err := tc.h.handle(rw, req)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				hErr, ok := err.(*HttpError)
				if !ok {
					t.Fatalf("expected HttpError, got %T", err)
				}
				if hErr.status != tc.wantStatus {
					t.Errorf("status = %v, wanted %v", hErr.status, tc.wantStatus)
				}
				allow, found := hErr.headers.Get("Allow")
				if found != (tc.wantAllow != "") {
					t.Errorf("Allow present = %t, wanted %t", found, tc.wantAllow != "")
				}
				if allow != tc.wantAllow {
					t.Errorf(`Allow = %#q, wanted %#q`, allow, tc.wantAllow)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if rw.s != tc.wantStatus {
				t.Errorf(`status = [%d, %s], wanted [%d, %s]`, rw.s.code, rw.s.msg, tc.wantStatus.code, tc.wantStatus.msg)
			}
			cType, found := rw.hdrs.Get("Content-Type")
			if found != (tc.wantCType != "") {
				t.Errorf("Content-Type present = %t, wanted %t", found, tc.wantCType != "")
			}
			if cType != tc.wantCType {
				t.Errorf(`Content-Type = %#q, wanted %#q`, cType, tc.wantCType)
			}
			if cLen := rw.hdrs.ContentLength(); cLen != tc.wantCLen {
				t.Errorf("content length = %d, wanted %d", cLen, tc.wantCLen)
			}
			allow, found := rw.hdrs.Get("Allow")
			if found != (tc.wantAllow != "") {
				t.Errorf("Allow present = %t, wanted %t", found, tc.wantAllow != "")
			}
			if allow != tc.wantAllow {
				t.Errorf(`Allow = %#q, wanted %#q`, allow, tc.wantAllow)
			}
			if string(rw.bdy) != tc.wantBody {
				t.Errorf(`body = %#q, wanted %#q`, string(rw.bdy), tc.wantBody)
			}
		})
	}
}
