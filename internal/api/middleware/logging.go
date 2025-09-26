package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a custom ResponseWriter to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		
		// Call the next handler
		next.ServeHTTP(wrapped, r)
		
		// Log the request
		duration := time.Since(start)
		log.Printf(
			"%s %s %d %v %s %s",
			r.Method,
			r.RequestURI,
			wrapped.statusCode,
			duration,
			r.RemoteAddr,
			r.UserAgent(),
		)
	})
}

// responseWriter is a wrapper around http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// DetailedLoggingMiddleware provides more detailed logging including request body size
func DetailedLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a custom ResponseWriter to capture status code and response size
		wrapped := &detailedResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			responseSize:   0,
		}
		
		// Get request size
		requestSize := r.ContentLength
		if requestSize < 0 {
			requestSize = 0
		}
		
		// Call the next handler
		next.ServeHTTP(wrapped, r)
		
		// Log the request with more details
		duration := time.Since(start)
		log.Printf(
			"[%s] %s %s %d %v - Request: %d bytes, Response: %d bytes, IP: %s, User-Agent: %s",
			time.Now().Format("2006-01-02 15:04:05"),
			r.Method,
			r.RequestURI,
			wrapped.statusCode,
			duration,
			requestSize,
			wrapped.responseSize,
			r.RemoteAddr,
			r.UserAgent(),
		)
	})
}

// detailedResponseWriter captures both status code and response size
type detailedResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseSize int64
}

// WriteHeader captures the status code
func (rw *detailedResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response size
func (rw *detailedResponseWriter) Write(data []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(data)
	rw.responseSize += int64(size)
	return size, err
}