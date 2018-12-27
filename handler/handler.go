// Package handler provides http.Handler for handling http requests
package handler

import (
	"GoExcercise/namegen"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

const downloadFolder = "./downloads/"

// NewHandler creates a new ServeMux and provides pattern-handler mapping
func NewHandler() http.Handler{
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", UploadHandler)
	mux.HandleFunc("/download/", DownloadHandler)
	mux.HandleFunc("/delete/", DeleteHandler)
	return mux
}

// DeleteHandler provies an ability to delete a file from server
// Simply delete the file from local storage and if the file was successfully deleted
// send the deleted filename in the response
func DeleteHandler(writer http.ResponseWriter, request *http.Request) {
	filename := path.Base(request.URL.Path)

	if _, err := os.Stat(downloadFolder + filename); os.IsNotExist(err) {
		writer.WriteHeader(404)
		return
	}
	err := os.Remove(downloadFolder + filename)
	if err != nil{
		log.Fatal(err)
	}

	if _, err := os.Stat(downloadFolder + filename); os.IsNotExist(err) {
		_, err = writer.Write([]byte("File Deleted: " + filename))
	}
}

// DownloadHandler provides an ability to download the file from server using browser
// Firstly DownloadHandler check if file is exist. If the check is successful, the file is copied to writer
func DownloadHandler(writer http.ResponseWriter, request *http.Request) {
	filename := path.Base(request.URL.Path)

	if _, err := os.Stat(downloadFolder + filename); os.IsNotExist(err) {
		writer.WriteHeader(404)
		return
	}
	file, err := os.Open(downloadFolder + filename)
	if err != nil{
		log.Fatal(err)
	}

	writer.Header().Set("Content-Disposition", "attachment; filename=" + filename)
	writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))
	_, err = io.Copy(writer, file)
	if err !=  nil {
		log.Fatal(err)
	}
}

// UploadHandler parses query from the request
// and provides uploading file to the server local storage
// with the randomly generated file name using namegen package
// If downloading was successful, UploadHandler sends the file name in response
func UploadHandler(writer http.ResponseWriter, request *http.Request) {

	uri := request.FormValue("uri")
	fileName := namegen.GenerateFileName(10)
	err := UploadFile(path.Join(downloadFolder, fileName), uri)
	if err != nil {
		_, _ = io.WriteString(writer, err.Error())
	}

	_, err = writer.Write([]byte(fileName))
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
