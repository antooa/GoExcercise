package main

import (
	"GoExcercise/client"
	srv "GoExcercise/server"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	addr := flag.String("addr", ":8080", "Server address")
	elastic := flag.String("elastic", "http://elastic:9200", "ES address")
	downloads := flag.String("downloads", "/tmp/downloads/", "Downloads folder path")
	index := flag.String("index", "files", "ES Index name")
	flag.Parse()

	elasticClient, err := client.NewElasticClient(*elastic, *index)
	if err != nil {
		log.Fatal(err)
	}

	server := srv.NewFileServer(elasticClient, *addr)

	err = os.MkdirAll(*downloads, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	_, _ = fmt.Fprintf(os.Stderr, "Listen on %v", *addr)
	log.Fatal(server.ListenAndServe())
}
