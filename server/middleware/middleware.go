package middleware

import "net/http"

type Middleware struct {
	functions []func(h http.Handler) http.Handler
}

func (m *Middleware) Then(handler http.Handler) http.Handler {
	if handler == nil {
		handler = http.DefaultServeMux
	}
	// h create
	// m.functions [logger, auth, isOwner]
	for i := range m.functions {
		// each middleware returns an http.Handler
		// h = M[n-1](...(M[0](controller)))
		// h = HTTPLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {...})
		handler = m.functions[len(m.functions)-1-i](handler)
	}
	return handler
}

func New(ms ...func(h http.Handler) http.Handler) *Middleware {
	return &Middleware{
		functions: ms,
	}
}
