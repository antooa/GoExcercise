package main

import (
	"GoExcercise/client"
	"GoExcercise/handler"
	srv "GoExcercise/server"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	addr := flag.String("addr", ":8080", "Server address")
	elastic := flag.String("elastic", "", "ES address")
	index := flag.String("index", "files", "ES Index name")
	sql := flag.String(
		"sql",
		"",
		"Postgres connection string")
	downloads := flag.String("downloads", "/tmp/downloads/", "Downloads folder path")

	flag.Parse()

	var storageClient handler.Storage
	var err error
	if *elastic == "" && *sql == "" {
		log.Fatal("Error: data storage address is not set")

	} else if *sql == "" {
		storageClient, err = client.NewElasticClient(*elastic, *index)
		if err != nil {
			log.Fatal(err)
		}

	} else {
		storageClient, err = client.NewPostgresClient(*sql)
		if err != nil{
			log.Fatal(err)
		}
	}

	server := srv.NewFileServer(storageClient, *addr)

	err = os.MkdirAll(*downloads, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	_, _ = fmt.Fprintf(os.Stderr, "Listen on %v", *addr)
	log.Fatal(server.ListenAndServe())
}
