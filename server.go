package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	// Chi
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	// Logging
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic("failed to initialize logger")
	}

	r := chi.NewRouter()

	r.Use(
		middleware.Heartbeat("/health"),
		NewLoggerMiddleware(logger.Sugar()),
	)

	r.Get("/*", NewFileServer("/static", "index.html"))

	fmt.Println("Serving on port 3000")
	if err := http.ListenAndServe(":3000", r); err != nil {
		fmt.Println("BryrupTeater-Web quit unexpectedly")
	}
}

var loggerCtxKey = "logger"

func GetLogger(r *http.Request) (*zap.SugaredLogger, error) {
	logger, ok := r.Context().Value(loggerCtxKey).(*zap.SugaredLogger)
	if !ok {
		return nil, errors.New("no logger configured")
	}
	return logger, nil
}

func NewLoggerMiddleware(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func (w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), loggerCtxKey,logger))

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}


func NewFileServer(root, index string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger, err := GetLogger(r)
		if err != nil {
			panic(err)
		}
		p := filepath.Join(root, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(p); os.IsNotExist(err) {
			logger.Infow("file not found, serving index",
				zap.String("path", p),
				zap.String("index", filepath.Join(root, index)),
				)
			http.ServeFile(w, r, filepath.Join(root, index))
			return
		} else if info.IsDir() {
			logger.Infow("dir found, serving index",
				zap.String("path", p),
				zap.String("index", filepath.Join(root, index)),
			)
			http.ServeFile(w, r, filepath.Join(root, index))
			return
		}

		logger.Infow("serving file", zap.String("path", p))
		http.ServeFile(w, r, p)
	}
}