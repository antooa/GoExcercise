// Package server provides Server generation
package server

import (
	"GoExcercise/handler"
	"github.com/olivere/elastic"
	"net/http"
	"time"
)

// NewFileServer returns a new Server. Also creates new Handler for the Server
func NewFileServer(elasticClient *elastic.Client) *http.Server {
	myHandler := handler.NewHandler(elasticClient)
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

