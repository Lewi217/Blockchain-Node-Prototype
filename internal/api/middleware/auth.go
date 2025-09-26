package middleware

import (
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"
)

// BasicAuthMiddleware provides HTTP Basic Authentication
func BasicAuthMiddleware(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the Authorization header
			auth := r.Header.Get("Authorization")
			
			if auth == "" {
				requireAuth(w)
				return
			}
			
			// Check if it's Basic auth
			if !strings.HasPrefix(auth, "Basic ") {
				requireAuth(w)
				return
			}
			
			// Decode the credentials
			payload, err := base64.StdEncoding.DecodeString(auth[6:])
			if err != nil {
				requireAuth(w)
				return
			}
			
			// Split username and password
			pair := strings.SplitN(string(payload), ":", 2)
			if len(pair) != 2 {
				requireAuth(w)
				return
			}
			
			// Verify credentials using constant time comparison to prevent timing attacks
			if subtle.ConstantTimeCompare([]byte(pair[0]), []byte(username)) != 1 ||
				subtle.ConstantTimeCompare([]byte(pair[1]), []byte(password)) != 1 {
				requireAuth(w)
				return
			}
			
			// Authentication successful, continue
			next.ServeHTTP(w, r)
		})
	}
}

// requireAuth sends a 401 Unauthorized response
func requireAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Blockchain Node"`)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Unauthorized"))
}

// APIKeyAuthMiddleware provides API key authentication
func APIKeyAuthMiddleware(validAPIKeys []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check for API key in header
			apiKey := r.Header.Get("X-API-Key")
			
			if apiKey == "" {
				// Check for API key in query parameter as fallback
				apiKey = r.URL.Query().Get("api_key")
			}
			
			if apiKey == "" {
				http.Error(w, "API key required", http.StatusUnauthorized)
				return
			}
			
			// Verify API key
			valid := false
			for _, validKey := range validAPIKeys {
				if subtle.ConstantTimeCompare([]byte(apiKey), []byte(validKey)) == 1 {
					valid = true
					break
				}
			}
			
			if !valid {
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
				return
			}
			
			// Authentication successful, continue
			next.ServeHTTP(w, r)
		})
	}
}

// OptionalAuthMiddleware provides optional authentication (continues even if auth fails)
func OptionalAuthMiddleware(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc
	}
}