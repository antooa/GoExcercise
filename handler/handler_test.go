package handler_test

import (
	"GoExcercise/handler"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var SampleUrl = "https://www.google.com/images/branding/googlelogo/2x/googlelogo_color_120x44dp.png"

func TestDeleteHandler(t *testing.T) {
	const filename = "old"

	err := handler.UploadFile(handler.DownloadFolder+filename, SampleUrl)
	if err != nil {
		t.Fatal("Error while uploading file: ", err.Error())
	}

	req, err := http.NewRequest("DELETE", "/delete/"+filename, nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	handler.NewHandler().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: %v want %v", status, http.StatusOK)
	}

	if _, err := os.Stat(handler.DownloadFolder + filename); err == nil {
		t.Fatal("File has not been deleted")
	}

}

func TestUploadHandler(t *testing.T) {

	req, err := http.NewRequest("POST", "/upload?uri="+SampleUrl, nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	handler.NewHandler().ServeHTTP(recorder, req)
	filename := recorder.Body.String()

	if status := recorder.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: %v want %v", status, http.StatusOK)
	}

	if _, err := os.Stat(handler.DownloadFolder + filename); os.IsNotExist(err) {
		t.Fatal(err)
	}

	err = os.Remove(handler.DownloadFolder + filename)
	if err != nil {
		t.Fatal(err)
	}

}

func TestRenameHandler(t *testing.T) {
	const oldName = "old"
	const newName = "new"

	err := handler.UploadFile(handler.DownloadFolder+oldName, SampleUrl)
	if err != nil {
		t.Fatalf("Error while uploading file: %v", err.Error())
	}

	req, err := http.NewRequest("PUT", "/rename/"+oldName+"/new/"+newName, nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	handler.NewHandler().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: %v, wanted: %v", status, http.StatusOK)
	}

	expected := newName
	if recorder.Body.String() != expected {
		t.Fatalf("handler returned wrong status code: %v, wanted: %v ", recorder.Body.String(), expected)
	}

	err = os.Remove(handler.DownloadFolder + newName)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUploadFile(t *testing.T) {

	err := handler.UploadFile( handler.DownloadFolder+"testfile1", "")
	expected := `Get : unsupported protocol scheme ""`
	if err.Error() != expected {
		t.Fatal("For", `""`, "expected", expected, "got", err.Error())
	}

	err = handler.UploadFile(handler.DownloadFolder+"testfile2", SampleUrl)
	expected = string("<nil>")
	if err != nil {
		t.Fatal("For", `google pic`, "expected", expected, "got", err.Error())
	}
	err = os.Remove(handler.DownloadFolder+"testfile1")
	if err != nil {
		t.Fatal(err)
	}
	err = os.Remove(handler.DownloadFolder+"testfile2")
	if err != nil {
		t.Fatal(err)
	}
}
