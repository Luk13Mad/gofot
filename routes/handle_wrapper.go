package routes

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"
)

type HandleWrapper struct {
	templates     map[string]*template.Template
	db            *sql.DB
	Logger        *log.Logger
	NextRequestID func() string
}

func NewHandleWrapper(templates map[string]*template.Template, database *sql.DB, Logger *log.Logger, NextRequestID func() string) *HandleWrapper {
	return &HandleWrapper{
		templates:     templates,
		db:            database,
		Logger:        Logger,
		NextRequestID: NextRequestID,
	}
}

func (hw *HandleWrapper) Logging(hdlr http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func(start time.Time) {
			requestID := w.Header().Get("X-Request-Id")
			if requestID == "" {
				requestID = "unknown"
			}
			hw.Logger.Println("Processed request: ", requestID, " Duration: ", time.Since(start))
		}(time.Now())
		hdlr.ServeHTTP(w, req)
	})
}

func (hw *HandleWrapper) Tracing(hdlr http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestID := req.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = hw.NextRequestID()
		}
		w.Header().Set("X-Request-Id", requestID)
		hw.Logger.Println("Received request: ", requestID, req.Method, req.URL.Path, req.RemoteAddr)
		hdlr.ServeHTTP(w, req)
	})
}
