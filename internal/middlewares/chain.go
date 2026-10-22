// Package middlewares - encadenador de middlewares
package middlewares

import (
	"net/http"
)

// Middleware contrato de middlwares
type Middleware func(http.Handler) http.Handler

// Chain encadena varios middlewares
func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}
