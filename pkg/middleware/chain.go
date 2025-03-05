package middleware

import (
	"net/http"
)

type Middleware func(http.Handler) http.Handler

func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		} // сначала последняя оборачивается предпоследней, и оборачивается всеми. выполнение идёт с первой
		return next
	}
}
