package middleware

import (
	"net/http"
)

// CORSMiddleware adds CORS headers to allow cross-origin requests
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Range")
		w.Header().Set("Access-Control-Allow-Credentials", "false")
		
		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// Continue with the next handler
		next.ServeHTTP(w, r)
	})
}

// CORSMiddlewareWithOrigins creates a CORS middleware with specific allowed origins
func CORSMiddlewareWithOrigins(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}
			
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Range")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			
			// Handle preflight OPTIONS request
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			// Continue with the next handler
			next.ServeHTTP(w, r)
		})
	}
}