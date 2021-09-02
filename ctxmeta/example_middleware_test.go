package ctxmeta_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/DoNewsCode/core/ctxmeta"
)

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bag := ctxmeta.GetBaggage(r.Context())
	bag.Set("method", r.Method)
	bag.Set("path", r.URL.Path)
	bag.Set("content_length", r.ContentLength)
	fmt.Fprintln(w, "OK")
}

type Middleware struct {
	next http.Handler
}

func NewMiddleware(next http.Handler) *Middleware {
	return &Middleware{next: next}
}

func (mw *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bag, ctx := ctxmeta.Inject(r.Context())

	defer func() {
		for _, kv := range bag.Slice() {
			fmt.Printf("%s: %v\n", kv.Key, kv.Val)
		}
	}()

	mw.next.ServeHTTP(w, r.WithContext(ctx))
}

func Example_middleware() {
	server := NewServer()
	middleware := NewMiddleware(server)
	testserver := httptest.NewServer(middleware)
	defer testserver.Close()
	http.Post(testserver.URL+"/path", "text/plain; charset=utf-8", strings.NewReader("hello world"))

	// Output:
	// method: POST
	// path: /path
	// content_length: 11
}
