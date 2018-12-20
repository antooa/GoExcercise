package main

import (
	srv "GoExcercise/server"
	"log"
)

func main() {
	server := srv.NewFileServer()
	log.Fatal(server.ListenAndServe())
}
