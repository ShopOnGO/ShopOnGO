package middleware

import (
	"net/http"

	"github.com/ShopOnGO/ShopOnGO/prod/pkg/logger"
)

func CheckRole(next http.Handler, requiredRole string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		roles, ok := r.Context().Value(ContextRolesKey).([]string)
		if !ok {
			logger.Error("❌ Roles not found in context")
			writeUnauthed(w)
			return
		}

		for _, role := range roles {
			if role == requiredRole {
				next.ServeHTTP(w, r)
				return
			}
		}

		logger.Error("❌ Your role does not have sufficient access rights:", requiredRole)
		http.Error(w, "Forbidden", http.StatusForbidden)
	})
}
