package storage

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type RequestIDKey struct{}

func GenerateRandomID() string {
	return uuid.New().String()
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &responseRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r)

		requestID, ok := r.Context().Value(RequestIDKey{}).(string)
		if !ok {
			requestID = "unknown"
		}
		ipAddress := r.RemoteAddr
		statusCode := recorder.statusCode
		logTime := time.Now()

		log.Printf("Request logged: time=%s, ip=%s, request_id=%s, status_code=%d", logTime, ipAddress, requestID, statusCode)
	})
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.URL.Query().Get("request_id")
		if requestID == "" {
			requestID = GenerateRandomID()
		}
		ctx := context.WithValue(r.Context(), RequestIDKey{}, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
