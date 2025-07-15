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
}

func TestHandleGet(t *testing.T) {
	h := Get(func(rw *ResponseWriter, req *Request) error {
		rw.Text("From inside the handler")
		return nil
	})
	req := requestWithUrlAndMethod("/foo", GET)
	rw := newResponse(StatusOK)
	wantMsg := "From inside the handler"

	err := h.handle(rw, req)

	if err != nil {
		t.Errorf("did not want error to be present, was %#q", err.Error())
	}
	if string(rw.bdy) != wantMsg {
		t.Errorf(`body = %#q, wanted %#q`, string(rw.bdy), wantMsg)
	}
	cType, found := rw.hdrs["Content-Type"]
	if !found {
		t.Error("expected Content-Type header to be present")
	}
	if cType != "text/plain" {
		t.Errorf(`Content-Type = %#q, wanted "text/plain"`, cType)
	}
}

func TestHandleHead(t *testing.T) {
	h := Get(func(rw *ResponseWriter, req *Request) error {
		rw.Text("From inside the handler")
		return nil
	})
	req := requestWithUrlAndMethod("/foo", HEAD)
	rw := newResponse(StatusOK)
	wantLen := uint(len("From inside the handler"))

	err := h.handle(rw, req)

	if err != nil {
		t.Errorf("did not want error to be present, was %#q", err.Error())
	}
	if len(rw.bdy) != 0 {
		t.Errorf("did not want body to be present, was %#q", string(rw.bdy))
	}
	cLen := rw.hdrs.contentLength()
	if cLen != wantLen {
		t.Errorf("content length = %d, wanted %d", cLen, wantLen)
	}
	cType, found := rw.hdrs["Content-Type"]
	if !found {
		t.Error("expected Content-Type header to be present")
	}
	if cType != "text/plain" {
		t.Errorf(`Content-Type = %#q, wanted "text/plain"`, cType)
	}
}

func TestHandlePost(t *testing.T) {
	h := Post(func(rw *ResponseWriter, req *Request) error {
		rw.Text("From inside the handler")
		return nil
	})
	req := requestWithUrlAndMethod("/foo", POST)
	rw := newResponse(StatusOK)
	wantMsg := "From inside the handler"

	err := h.handle(rw, req)

	if err != nil {
		t.Errorf("did not want error to be present, was %#q", err.Error())
	}
	if string(rw.bdy) != wantMsg {
		t.Errorf(`body = %#q, wanted %#q`, string(rw.bdy), wantMsg)
	}
	cType, found := rw.hdrs["Content-Type"]
	if !found {
		t.Error("expected Content-Type header to be present")
	}
	if cType != "text/plain" {
		t.Errorf(`Content-Type = %#q, wanted "text/plain"`, cType)
	}
}

func TestHandleUnsupportedMethod(t *testing.T) {
	h := Get(func(rw *ResponseWriter, req *Request) error {
		rw.Text("From inside the handler")
		return nil
	})
	req := requestWithUrlAndMethod("/foo", POST)
	rw := newResponse(StatusOK)

	err := h.handle(rw, req)

	if err == nil {
		t.Error("expected error, was nil")
	}
	hErr, ok := err.(*HttpError)
	if !ok {
		t.Fatalf("expected HttpError, got %T, %v", err, err)
	}
	if hErr.status != StatusMethodNotAllowed {
		t.Errorf(`status = %v, wanted "405: Method Not Allowed"`, hErr.status)
	}
	allow, found := hErr.headers["Allow"]
	if !found {
		t.Error(`expected "Allow" header to be present, was not found`)
	}
	if allow != "GET, HEAD" {
		t.Errorf(`headers["Allow"] = %#q, wanted "GET, HEAD"`, allow)
	}
}
