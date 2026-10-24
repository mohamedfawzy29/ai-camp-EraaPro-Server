package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
)

// AdminAuth validates that the request contains the correct admin secret key.
// It checks in this order:
// 1. Query parameter: ?secret=xxx or ?key=xxx (ideal for direct browser download links)
// 2. Header: X-Admin-Key: xxx
// 3. Header: Authorization: Bearer xxx
func AdminAuth(expectedSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			providedSecret := ""

			// 1. Check Query parameter
			if qSecret := r.URL.Query().Get("secret"); qSecret != "" {
				providedSecret = qSecret
			} else if qKey := r.URL.Query().Get("key"); qKey != "" {
				providedSecret = qKey
			}

			// 2. Check X-Admin-Key Header
			if providedSecret == "" {
				providedSecret = r.Header.Get("X-Admin-Key")
			}

			// 3. Check Authorization: Bearer Header
			if providedSecret == "" {
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					providedSecret = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if providedSecret == "" || providedSecret != expectedSecret {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "غير مصرح لك بالوصول (كلمة السر غير صحيحة أو مفقودة)",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
