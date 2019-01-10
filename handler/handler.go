// Package handler provides http.Handler for handling http requests
package handler

import (
	cli "GoExcercise/client"
	"GoExcercise/namegen"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/olivere/elastic"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
)

const DownloadFolder = "/tmp/downloads/"

func init() {
	err := os.MkdirAll(DownloadFolder, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

// NewHandler creates a new gorilla mux router and provides pattern-handler mapping
func NewHandler(elasticClient *elastic.Client ) http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/upload", UploadHandler)
	router.HandleFunc("/download/", DownloadHandler)
	router.HandleFunc("/delete/{filename}", DeleteHandler)
	router.HandleFunc("/rename/{oldName}/new/{newName}", RenameHandler)
	router.HandleFunc("/description/{filename}", NewDescriptionHandler(elasticClient)).Methods("POST")
	router.HandleFunc("/description/{filename}", GetDescriptionHandler(elasticClient)).Methods("GET")
	return router
}

func GetDescriptionHandler(client *elastic.Client) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		filename := vars["filename"]
		// Search with a term query
		termQuery := elastic.NewTermQuery("name", filename)
		searchResult, err := client.Search().
			Index(cli.IndexName).            // search in index "files"
			Query(termQuery).           // specify the query
			Sort("name.keyword", true). // sort by "name" field, ascending
			From(0).Size(10).           // take documents 0-9
			Pretty(true).               // pretty print request and response JSON
			Do(context.Background())    // execute
		if err != nil {
			http.Error(writer, err.Error(), http.StatusNotFound)
		}

		//var file cli.File
		//for _, item := range searchResult.Each(reflect.TypeOf(file)) {
		//	if t, ok := item.(cli.File); ok {
		//		_, err = writer.Write([]byte(t.Name + " " + t.Description) )
		//		if err != nil{
		//			http.Error(writer, err.Error(), http.StatusInternalServerError)
		//		}
		//	}
		//}
		if searchResult.Hits.TotalHits > 0 {
			fmt.Printf("Found a total of %d tweets\n", searchResult.Hits.TotalHits)

			// Iterate through results
			for _, hit := range searchResult.Hits.Hits {
				// hit.Index contains the name of the index

				// Deserialize hit.Source into a cli.File (could also be just a map[string]interface{}).
				var t cli.File
				err := json.Unmarshal(*hit.Source, &t)
				if err != nil {
					// Deserialization failed
					http.Error(writer, err.Error(), http.StatusInternalServerError)
				}

				// Work with cli.File
				_, err =writer.Write([]byte(t.Name + " " + t.Description))
				if err != nil{
					http.Error(writer, err.Error(), http.StatusInternalServerError)
				}
			}
		} else {
			// No hits
			_, err =writer.Write([]byte("Found no description\n"))
			if err != nil{
				http.Error(writer, err.Error(), http.StatusInternalServerError)
			}
		}

	}
}

func NewDescriptionHandler(client *elastic.Client) http.HandlerFunc{
	return func (writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		filename := vars["filename"]
		body, err := ioutil.ReadAll(request.Body)
		if err != nil{
			http.Error(writer, err.Error(), http.StatusBadRequest)
		}

		file := cli.File{
			Name:        filename,
			Description: string(body),
		}
		_, err = client.Index().
			Index(cli.IndexName).
			Type("doc").
			BodyJson(file).
			Refresh("wait_for").
			Do(context.Background())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}

		_, err = writer.Write([]byte(filename + " " + string(body)))
		if err != nil{
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	}
}



func RenameHandler(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	oldName := vars["oldName"]
	newName := vars["newName"]
	if _, err := os.Stat(DownloadFolder + oldName); os.IsNotExist(err) {
		writer.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err := os.Stat(DownloadFolder + newName); !os.IsNotExist(err) {
		writer.WriteHeader(http.StatusConflict)
		return
	}
	err := os.Rename(DownloadFolder+oldName, DownloadFolder+newName)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = writer.Write([]byte(newName))
}

// DeleteHandler provides an ability to delete a file from server
//
// Simply delete the file from the local storage and if the file was successfully deleted
// send the deleted filename in the response
func DeleteHandler(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	filename := vars["filename"]

	if _, err := os.Stat(DownloadFolder + filename); os.IsNotExist(err) {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	err := os.Remove(DownloadFolder + filename)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := os.Stat(DownloadFolder + filename); os.IsNotExist(err) {
		_, err = writer.Write([]byte("File Deleted: " + filename))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// DownloadHandler provides an ability to download the file from server using browser
// Firstly DownloadHandler check if file is exist. If the check is successful, the file is copied to writer
func DownloadHandler(writer http.ResponseWriter, request *http.Request) {
	filename := path.Base(request.URL.Path)

	if _, err := os.Stat(DownloadFolder + filename); os.IsNotExist(err) {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	file, err := os.Open(DownloadFolder + filename)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Disposition", "attachment; filename="+filename)
	writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))
	_, err = io.Copy(writer, file)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

// UploadHandler parses query from the request
// and provides uploading file to the server local storage
// with the randomly generated file name using namegen package
// If downloading was successful, UploadHandler sends the file name in response
func UploadHandler(writer http.ResponseWriter, request *http.Request) {

	uri := request.FormValue("uri")
	fileName := namegen.GenerateFileName(10)
	err := UploadFile(path.Join(DownloadFolder, fileName), uri)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusNotFound)
		return
	}

	_, err = writer.Write([]byte(fileName))
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

// UploadFile copies the body of the response into the new file on the server
func UploadFile(filepath string, url string) error {

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
