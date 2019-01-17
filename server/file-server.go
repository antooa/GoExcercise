// Package server provides Server generation
package server

import (
	"GoExcercise/handler"
	"net/http"
	"time"
)

// NewFileServer returns a new Server. Also creates new Handler for the Server
func NewFileServer(storage handler.Storage) *http.Server {
	myHandler := handler.NewHandler(storage)
	return &http.Server{
		Addr:              ":8080",
		Handler:           myHandler,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 0,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       0,
		MaxHeaderBytes:    1 << 20,
	}
}
