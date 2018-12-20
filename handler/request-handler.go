// Package handler provides http.Handler for handling http requests
package handler

import (
	"GoExcercise/namegen"
	"io"
	"net/http"
	"os"
	"path"
)

const downloadFolder = "./downloads/"

// NewHandler creates a new ServeMux and provides pattern-handler mapping
func NewHandler() http.Handler{
	mux := http.NewServeMux()
	mux.HandleFunc("/download", DownloadHandler)
	return mux
}

// DownloadHandler parses query from the request
// and provides downloading file to the local storage
// with the randomly generated file name using namegen package
// If downloading was successful, DownloadHandler sends the file name in response
func DownloadHandler(writer http.ResponseWriter, request *http.Request) {

	uri := request.FormValue("uri")
	fileName := namegen.GenerateFileName(10)
	err := DownloadFile(path.Join(downloadFolder, fileName), uri)
	if err != nil {
		_, _ = io.WriteString(writer, err.Error())
	}

	_, err = writer.Write([]byte(fileName))
}

// DownloadFile copies the body of the response into the new file in the local storage
func DownloadFile(filepath string, url string) error {

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
