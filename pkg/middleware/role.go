package middleware

import (
	"net/http"

	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
)

func CheckRole(next http.Handler, allowedRoles []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := r.Context().Value(ContextRolesKey)
		
		logger.Infof("Role found in request context: %s", role)

		if role == "" {
			logger.Error("❌ No role found in request context. Allowed roles:", allowedRoles)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				next.ServeHTTP(w, r)
				return
			}
		}

		logger.Error("❌ Your role does not have sufficient access rights. Allowed roles:", allowedRoles)
		http.Error(w, "Forbidden", http.StatusForbidden)
	})
}