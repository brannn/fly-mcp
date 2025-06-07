package server

import (
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a response writer wrapper to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Call next handler
		next.ServeHTTP(wrapper, r)
		
		// Log the request
		s.logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Str("user_agent", r.UserAgent()).
			Int("status_code", wrapper.statusCode).
			Dur("duration", time.Since(start)).
			Msg("HTTP request")
	})
}

// corsMiddleware handles CORS headers
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// Check if origin is allowed
		if s.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// rateLimitMiddleware implements rate limiting
func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	// Create a rate limiter
	limiter := rate.NewLimiter(rate.Limit(s.config.Security.RateLimitRPS), s.config.Security.RateLimitRPS*2)
	
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check rate limit
		if !limiter.Allow() {
			s.logger.Warn().
				Str("remote_addr", r.RemoteAddr).
				Str("path", r.URL.Path).
				Msg("Rate limit exceeded")
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "rate limit exceeded"}`))
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// isOriginAllowed checks if an origin is allowed based on configuration
func (s *Server) isOriginAllowed(origin string) bool {
	if origin == "" {
		return true // Allow requests without origin (e.g., from curl)
	}
	
	for _, allowed := range s.config.Security.AllowedOrigins {
		if allowed == "*" {
			return true
		}
		
		// Simple wildcard matching for localhost and development
		if strings.Contains(allowed, "*") {
			pattern := strings.Replace(allowed, "*", "", -1)
			if strings.Contains(origin, pattern) {
				return true
			}
		}
		
		if origin == allowed {
			return true
		}
	}
	
	return false
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
