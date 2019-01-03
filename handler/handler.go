// Package handler provides http.Handler for handling http requests
package handler

import (
	"GoExcercise/namegen"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

const DownloadFolder = "/tmp/downloads/"

func init(){
	err := os.MkdirAll(DownloadFolder, os.ModePerm)
	if err != nil{
		log.Fatal(err)
	}
}

// NewHandler creates a new gorilla mux router and provides pattern-handler mapping
func NewHandler() http.Handler{
	router := mux.NewRouter()
	router.HandleFunc("/upload", UploadHandler)
	router.HandleFunc("/download/", DownloadHandler)
	router.HandleFunc("/delete/{filename}", DeleteHandler)
	router.HandleFunc("/rename/{oldName}/new/{newName}", RenameHandler)
	return router
}

func RenameHandler(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	oldName := vars["oldName"]
	newName := vars["newName"]
	if _, err := os.Stat(DownloadFolder + oldName); os.IsNotExist(err) {
		writer.WriteHeader(404)
		return
	} else if err != nil{
		writer.WriteHeader(500)
		return
	}
	if _, err := os.Stat(DownloadFolder + newName); !os.IsNotExist(err) {
		writer.WriteHeader(409)
		return
	}
	err := os.Rename(DownloadFolder + oldName, DownloadFolder + newName)
	if err != nil{
		writer.WriteHeader(500)
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
		writer.WriteHeader(404)
		return
	}

	err := os.Remove(DownloadFolder + filename)
	if err != nil{
		http.Error(writer, err.Error(), 500)
	}

	if _, err := os.Stat(DownloadFolder + filename); os.IsNotExist(err) {
		_, err = writer.Write([]byte("File Deleted: " + filename))
		if err != nil{
			http.Error(writer, err.Error(), 500)
		}
	}
}

// DownloadHandler provides an ability to download the file from server using browser
// Firstly DownloadHandler check if file is exist. If the check is successful, the file is copied to writer
func DownloadHandler(writer http.ResponseWriter, request *http.Request) {
	filename := path.Base(request.URL.Path)

	if _, err := os.Stat(DownloadFolder + filename); os.IsNotExist(err) {
		writer.WriteHeader(404)
		return
	}
	file, err := os.Open(DownloadFolder + filename)
	if err != nil{
		http.Error(writer, err.Error(), 500)
	}

	writer.Header().Set("Content-Disposition", "attachment; filename=" + filename)
	writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))
	_, err = io.Copy(writer, file)
	if err !=  nil {
		http.Error(writer, err.Error(), 500)
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
		http.Error(writer, err.Error(), 404)
		return
	}

	_, err = writer.Write([]byte(fileName))
	if err != nil{
		http.Error(writer, err.Error(), 500)
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
