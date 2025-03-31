package middleware

import "net/http"

// Middleware is a function that wraps a http.Handler.
type Middleware func(next http.Handler) http.Handler