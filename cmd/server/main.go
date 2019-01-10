package main

import (
	"GoExcercise/client"
	srv "GoExcercise/server"
	"fmt"
	"log"
	"os"
)

func main() {
	elasticClient, err := client.NewElasticClient()
	if err != nil{
		log.Fatal(err)
	}
	server := srv.NewFileServer(elasticClient)
	_, _ = fmt.Fprintf(os.Stderr, "Listen on %v", server.Addr)
	log.Fatal(server.ListenAndServe())
}
