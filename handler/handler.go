// Package handler provides http.Handler for handling http requests
package handler

import (
	"GoExcercise/namegen"
	"errors"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

const DownloadFolder = "/tmp/downloads/"

type File struct {
	Name        string `json:"name"`
	Url         string `json:"url"`
	Description string `json:"description"`
}

type Storage interface {
	Create(file File) (id string, err error)
	Read(id string) (file File, err error)
	Update(id string, newFile File) error
	Delete(id string) error
}

// NewHandler creates a new gorilla mux router and provides pattern-handler mapping
func NewHandler(storage Storage) http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/upload", RecoveryMiddleware(UploadHandler(storage))).Methods("POST")
	router.HandleFunc("/download/{id}", RecoveryMiddleware(DownloadHandler(storage))).Methods("GET")
	router.HandleFunc("/delete/{id}", RecoveryMiddleware(DeleteHandler(storage))).Methods("DELETE")
	router.HandleFunc("/rename/{id}/new/{newName}", RenameHandler(storage)).Methods("PUT")
	router.HandleFunc("/description/{id}", RecoveryMiddleware(NewDescriptionHandler(storage))).Methods("PUT")
	router.HandleFunc("/description/{id}", RecoveryMiddleware(GetDescriptionHandler(storage))).Methods("GET")
	return router
}

func RecoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request){
		defer func() {
			if r := recover(); r != nil {
				err := r.(error)
				http.Error(writer, err.Error(), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(writer, request)
	}
}

// GetDescriptionHandler provides a handler which returns a file description in response.
func GetDescriptionHandler(storage Storage) func(http.ResponseWriter, *http.Request) {

	return func(writer http.ResponseWriter, request *http.Request) {

		vars := mux.Vars(request)
		id := vars["id"]

		file, err := storage.Read(id)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusNotFound)
			return
		}

		_, err = writer.Write([]byte(file.Description))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// NewDescriptionHandler provides a handler which changes the Description field of the doc stored in ES.
//
// Returns a new description in response.
func NewDescriptionHandler(storage Storage) http.HandlerFunc {

	return func(writer http.ResponseWriter, request *http.Request) {

		vars := mux.Vars(request)
		id := vars["id"]

		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		defer request.Body.Close()

		oldFile, err := storage.Read(id)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusNotFound)
			return
		}

		file := File{
			Name:        oldFile.Name,
			Url:         oldFile.Url,
			Description: string(body),
		}
		err = storage.Update(id, file)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		changedFile, err := storage.Read(id)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = writer.Write([]byte(changedFile.Description))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// RenameHandler provides a handler which changes the Name field of the doc, stored in ES.
//
// Returns a new name in response.
func RenameHandler(storage Storage) func(http.ResponseWriter, *http.Request) {

	return func(writer http.ResponseWriter, request *http.Request) {

		vars := mux.Vars(request)
		newName := vars["newName"]
		id := vars["id"]

		//get the file
		file, err := storage.Read(id)
		oldName := file.Name

		//check if the old file exists
		if _, err := os.Stat(DownloadFolder + oldName); err != nil {
			http.Error(writer, err.Error(), http.StatusNotFound)
			return
		}
		//check if the new file doesn't exist
		if _, err := os.Stat(DownloadFolder + newName); err == nil {
			http.Error(writer,
				errors.New("File with name " + newName + " already exists").Error(),
				http.StatusConflict)
			return
		}

		//change file name to the new one
		file.Name = newName

		// update using the old file with a new file name
		err = storage.Update(id, file)
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
func DeleteHandler(storage Storage) func(http.ResponseWriter, *http.Request) {

	return func(writer http.ResponseWriter, request *http.Request) {

		vars := mux.Vars(request)
		id := vars["id"]

		file, err := storage.Read(id)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := os.Stat(DownloadFolder + file.Name); os.IsNotExist(err) {
			http.Error(writer, err.Error(), http.StatusNotFound)
			return
		} else if err != nil{
			http.Error(writer, err.Error(), http.StatusInternalServerError)
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

		err = storage.Delete(id)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// DownloadHandler provides an ability to download the file from server using browser
// Firstly DownloadHandler check if file is exist. If the check is successful, the file is copied to writer
func DownloadHandler(storage Storage) func(http.ResponseWriter, *http.Request) {

	return func(writer http.ResponseWriter, request *http.Request) {

		vars := mux.Vars(request)
		id := vars["id"]

		doc, err := storage.Read(id)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := os.Stat(DownloadFolder + doc.Name); os.IsNotExist(err) {
			http.Error(writer, err.Error(), http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
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
func UploadHandler(storage Storage) http.HandlerFunc {

	return func(writer http.ResponseWriter, request *http.Request) {

		uri := request.FormValue("uri")
		fileName := namegen.GenerateFileName(10)
		file := File{
			Name:        fileName,
			Url:         uri,
			Description: "",
		}

		id, err := storage.Create(file)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		err = UploadFile(path.Join(DownloadFolder, fileName), uri)
		if err != nil {
			delErr := storage.Delete(id)
			if delErr != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
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
func UploadFile(filepath string, url string) (err error) {

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
