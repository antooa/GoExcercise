package main

import (
	"GoExcercise/client"
	srv "GoExcercise/server"
	"flag"
	"fmt"
	"log"
	"net/http"
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

	var server *http.Server
	if *elastic == "" && *sql == "" {
		log.Fatal("Error: data storage address is not set")

	} else if *sql == "" {
		elasticClient, err := client.NewElasticClient(*elastic, *index)
		if err != nil {
			log.Fatal(err)
		}
		server = srv.NewFileServer(elasticClient, *addr)

	} else {
		sqlClient, err := client.NewPostgresClient(*sql)
		if err != nil{
			log.Fatal(err)
		}
		server = srv.NewFileServer(sqlClient, *addr)
	}

	err := os.MkdirAll(*downloads, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	_, _ = fmt.Fprintf(os.Stderr, "Listen on %v", *addr)
	log.Fatal(server.ListenAndServe())
}
