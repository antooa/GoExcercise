package main

import (
	srv "GoExcercise/server"
	"fmt"
	"log"
	"os"
)

func main() {
	server := srv.NewFileServer()
	_, _ = fmt.Fprintf(os.Stderr, "Listen on %v", server.Addr)
	log.Fatal(server.ListenAndServe())
}
