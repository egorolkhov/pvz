package logger

import (
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

var Log *zap.Logger = zap.NewNop()

type responseData struct {
	StatusCode int
	Size       int
}

type LogResponseWriter struct {
	http.ResponseWriter
	Log *responseData
}

func (L *LogResponseWriter) Write(b []byte) (int, error) {
	size, err := L.ResponseWriter.Write(b)
	if err != nil {
		return 0, err
	}
	L.Log.Size = size
	return size, nil
}

func (L *LogResponseWriter) WriteHeader(statusCode int) {
	L.ResponseWriter.WriteHeader(statusCode)
	L.Log.StatusCode = statusCode
}

func InitLogger() error {
	cfg := zap.NewProductionConfig()
	cfg.Level.SetLevel(zap.InfoLevel)

	logger, err := cfg.Build()
	if err != nil {
		return err
	}
	defer logger.Sync()

	Log = logger

	return nil
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		data := &responseData{}
		lw := &LogResponseWriter{ResponseWriter: w, Log: data}

		next.ServeHTTP(lw, r)

		duration := time.Since(startTime)
		Log.Info("HTTP Request",
			zap.String("method", r.Method),
			zap.String("status code", strconv.Itoa(data.StatusCode)),
			zap.String("URI", r.RequestURI),
			zap.String("size", strconv.Itoa(data.Size)),
			zap.String("duration", duration.String()),
		)
	})
}
