package routeit

import (
	"reflect"
	"testing"
)

func TestHandlerConstructors(t *testing.T) {
	fn := func(rw *ResponseWriter, req *Request) error { return nil }

	type wantMethods struct {
		get, head, post, put, delete, patch, options, trace bool
	}

	tests := []struct {
		name    string
		handler Handler
		want    wantMethods
	}{
		{
			name:    "GET",
			handler: Get(fn),
			want:    wantMethods{get: true, head: true, options: true, trace: true},
		},
		{
			name:    "POST",
			handler: Post(fn),
			want:    wantMethods{post: true, options: true, trace: true},
		},
		{
			name:    "PUT",
			handler: Put(fn),
			want:    wantMethods{put: true, options: true, trace: true},
		},
		{
			name:    "DELETE",
			handler: Delete(fn),
			want:    wantMethods{delete: true, options: true, trace: true},
		},
		{
			name:    "PATCH",
			handler: Patch(fn),
			want:    wantMethods{patch: true, options: true, trace: true},
		},
		{
			name:    "OPTIONS",
			handler: Post(fn),
			want:    wantMethods{post: true, options: true, trace: true},
		},
		{
			name:    "TRACE",
			handler: Put(fn),
			want:    wantMethods{put: true, options: true, trace: true},
		},
		{
			name: "MultiMethod GET only",
			handler: MultiMethod(MultiMethodHandler{
				Get: fn,
			}),
			want: wantMethods{get: true, head: true, options: true},
		},
		{
			name: "MultiMethod POST only",
			handler: MultiMethod(MultiMethodHandler{
				Post: fn,
			}),
			want: wantMethods{post: true, options: true},
		},
		{
			name: "MultiMethod GET + POST",
			handler: MultiMethod(MultiMethodHandler{
				Get:  fn,
				Post: fn,
			}),
			want: wantMethods{get: true, head: true, post: true, options: true},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			checks := []struct {
				name     string
				actual   HandlerFunc
				expected bool
			}{
				{"get", tc.handler.get, tc.want.get},
				{"head", tc.handler.head, tc.want.head},
				{"post", tc.handler.post, tc.want.post},
				{"put", tc.handler.put, tc.want.put},
				{"delete", tc.handler.delete, tc.want.delete},
				{"options", tc.handler.options, tc.want.options},
			}

			for _, check := range checks {
				if (check.actual != nil) != check.expected {
					t.Errorf("handler.%s != nil = %t; want %t", check.name, check.actual != nil, check.expected)
				}
			}
		})
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
		wantAllow  []string
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
			method:     PATCH,
			h:          Patch(fn),
			wantBody:   "From inside the handler",
			wantStatus: StatusOK,
			wantCType:  "text/plain",
			wantCLen:   23,
		},
		{
			method:     OPTIONS,
			h:          Get(fn),
			wantStatus: StatusNoContent,
			wantAllow:  []string{"GET", "HEAD", "OPTIONS"},
		},
		{
			method:     TRACE,
			h:          Get(fn),
			wantStatus: StatusOK,
			wantCType:  "message/http",
		},
		{
			name:       "unsupported method",
			method:     POST,
			h:          Get(fn),
			wantStatus: StatusMethodNotAllowed,
			wantAllow:  []string{"GET", "HEAD", "OPTIONS"},
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
				allow, found := hErr.headers.All("Allow")
				if found != (len(tc.wantAllow) != 0) {
					t.Errorf("Allow present = %t, wanted %t", found, len(tc.wantAllow) != 0)
				}
				if !(len(allow) == 0 && tc.wantAllow == nil) && !reflect.DeepEqual(allow, tc.wantAllow) {
					t.Errorf(`Allow = %v, wanted %v`, allow, tc.wantAllow)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if rw.s != tc.wantStatus {
				t.Errorf(`status = [%d, %s], wanted [%d, %s]`, rw.s.code, rw.s.msg, tc.wantStatus.code, tc.wantStatus.msg)
			}
			wantCType := tc.wantCType != ""
			cType, found := rw.headers.headers.All("Content-Type")
			if found != wantCType {
				t.Errorf("Content-Type present = %t, wanted %t", found, tc.wantCType != "")
			}
			if wantCType && len(cType) != 1 {
				t.Errorf(`Content-Type = %+v, wanted only 1 element`, cType)
			}
			if wantCType && cType[0] != tc.wantCType {
				t.Errorf(`Content-Type = %#q, wanted %#q`, cType[0], tc.wantCType)
			}
			if cLen := rw.headers.headers.ContentLength(); cLen != tc.wantCLen {
				t.Errorf("content length = %d, wanted %d", cLen, tc.wantCLen)
			}
			allow, found := rw.headers.headers.All("Allow")
			if found != (len(tc.wantAllow) != 0) {
				t.Errorf("Allow present = %t, wanted %t", found, len(tc.wantAllow) != 0)
			}
			if !(len(allow) == 0 && tc.wantAllow == nil) && !reflect.DeepEqual(allow, tc.wantAllow) {
				t.Errorf(`Allow = %v, wanted %v`, allow, tc.wantAllow)
			}
			if string(rw.bdy) != tc.wantBody {
				t.Errorf(`body = %#q, wanted %#q`, string(rw.bdy), tc.wantBody)
			}
		})
	}
}
