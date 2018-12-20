package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

func main() {

	link := flag.String("file", "", "Text to parse.")
	flag.Parse()
	u := url.URL{
		Scheme:     "http",
		User:       nil,
		Host:       "localhost:8080",
		Path:    "/download",
		RawQuery:   url.Values{"uri":[]string{*link}}.Encode(),
	}
	
	fmt.Println(u.String())
	resp, err := http.Get(u.String())
	if err != nil{
		log.Fatal(err)
	}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil{
			log.Fatal(err)
		}

		bodyString := string(bodyBytes)
		fmt.Println(bodyString)
	}

}