package logging

import (
	"github.com/rs/zerolog"
	"net/http"
	"time"
)

type responseWrite struct {
	code           int
	internalWriter http.ResponseWriter
}

func (r *responseWrite) Header() http.Header {
	return r.internalWriter.Header()
}

func (r *responseWrite) Write(b []byte) (int, error) {
	return r.internalWriter.Write(b)
}

func (r *responseWrite) WriteHeader(statusCode int) {
	r.code = statusCode
	r.internalWriter.WriteHeader(statusCode)
}

func (r *responseWrite) Code() int {
	return r.code
}

func LoggingMiddleware(next http.Handler, logger *zerolog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		newWriter := responseWrite{internalWriter: w, code: 200}
		next.ServeHTTP(&newWriter, r)
		duration := time.Since(startTime)
		logger.Info().
			Str("host", r.Host).
			Int("code", newWriter.Code()).
			Str("user-agent", r.UserAgent()).
			Int64("duration", duration.Milliseconds()).
			Str("ip", r.RemoteAddr).
			Int64("content-length", r.ContentLength).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg(http.StatusText(newWriter.Code()))
	})
}
