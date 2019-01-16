// Package handler provides http.Handler for handling http requests
package handler

import (
	cli "GoExcercise/client"
	"GoExcercise/namegen"
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
func NewHandler(elasticClient *elastic.Client) http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/upload", UploadHandler(elasticClient)).Methods("POST")
	router.HandleFunc("/download/{id}", DownloadHandler(elasticClient)).Methods("GET")
	router.HandleFunc("/delete/{id}", DeleteHandler(elasticClient)).Methods("DELETE")
	router.HandleFunc("/rename/{id}/new/{newName}", RenameHandler(elasticClient)).Methods("PUT")
	router.HandleFunc("/description/{id}", NewDescriptionHandler(elasticClient)).Methods("PUT")
	router.HandleFunc("/description/{id}", GetDescriptionHandler(elasticClient)).Methods("GET")
	return router
}

// GetDescriptionHandler provides a handler which returns a file description in response.
func GetDescriptionHandler(client *elastic.Client) func(http.ResponseWriter, *http.Request) {

	return func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		id := vars["id"]

		file, err := cli.ReadDoc(client, id)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusNotFound)
		}

		_, err = writer.Write([]byte(file.Description))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	}
}

// NewDescriptionHandler provides a handler which changes the Description field of the doc stored in ES.
//
// Returns a new description in response.
func NewDescriptionHandler(client *elastic.Client) http.HandlerFunc {

	return func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		id := vars["id"]
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
		}

		oldFile, err := cli.ReadDoc(client, id)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusNotFound)
		}

		file := cli.File{
			Name:        oldFile.Name,
			Url:         oldFile.Url,
			Description: string(body),
		}

		err = cli.UpdateDoc(client, id, file)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}

		changedFile, err := cli.ReadDoc(client, id)
		_, err = writer.Write([]byte(changedFile.Description))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	}
}

// RenameHandler provides a handler which changes the Name field of the doc, stored in ES.
//
// Returns a new name in response.
func RenameHandler(client *elastic.Client) func(http.ResponseWriter, *http.Request) {

	return func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		newName := vars["newName"]
		id := vars["id"]

		//get the file
		file, err := cli.ReadDoc(client, id)
		oldName := file.Name

		//check if the file exists
		if _, err := os.Stat(DownloadFolder + oldName); os.IsNotExist(err) {
			http.Error(writer, err.Error(), http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := os.Stat(DownloadFolder + newName); !os.IsNotExist(err) {
			http.Error(writer, err.Error(), http.StatusConflict)
			return
		}

		//change file name to the new one
		file.Name = newName

		// update using the old file with a new file name
		err = cli.UpdateDoc(client, id, file)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		//rename in os
		err = os.Rename(DownloadFolder+oldName, DownloadFolder+newName)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		//send new name in response
		_, err = writer.Write([]byte(newName))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// DeleteHandler provides an ability to delete a file from server
//
// Simply delete the file from the local storage and if the file was successfully deleted
// send the deleted filename in the response
func DeleteHandler(client *elastic.Client) func(http.ResponseWriter, *http.Request) {

	return func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		id := vars["id"]

		file, err := cli.ReadDoc(client, id)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := os.Stat(DownloadFolder + file.Name); os.IsNotExist(err) {
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		err = os.Remove(DownloadFolder + file.Name)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := os.Stat(DownloadFolder + file.Name); os.IsNotExist(err) {
			_, err = writer.Write([]byte("Deleted file: " + file.Name))
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		err = cli.DeleteDoc(client, id)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// DownloadHandler provides an ability to download the file from server using browser
// Firstly DownloadHandler check if file is exist. If the check is successful, the file is copied to writer
func DownloadHandler(client *elastic.Client) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		id := vars["id"]

		doc, err := cli.ReadDoc(client, id)
		if err != nil{
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}

		if _, err := os.Stat(DownloadFolder + doc.Name); os.IsNotExist(err) {
			http.Error(writer, err.Error(), http.StatusNotFound)
			return
		}
		file, err := os.Open(DownloadFolder + doc.Name)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Disposition", "attachment; filename="+doc.Name)
		writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))
		_, err = io.Copy(writer, file)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
// UploadHandler parses query from the request
// and provides uploading file to the server local storage
// with the randomly generated file name using namegen package
// If downloading was successful, UploadHandler sends the file name in response
func UploadHandler(client *elastic.Client) http.HandlerFunc {

	return func(writer http.ResponseWriter, request *http.Request) {
		uri := request.FormValue("uri")
		fileName := namegen.GenerateFileName(10)
		file := cli.File{
			Name:        fileName,
			Url:         uri,
			Description: "",
		}

		id, err := cli.CreateDoc(client, file)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		err = UploadFile(path.Join(DownloadFolder, fileName), uri)
		if err != nil {
			delErr := cli.DeleteDoc(client, id)
			if delErr != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
			}
			http.Error(writer, err.Error(), http.StatusNotFound)
			return
		}

		_, err = writer.Write([]byte(id))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
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
