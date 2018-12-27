package handler_test

import (
	"GoExcercise/handler"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var SampleUrl = "https://www.google.com/images/branding/googlelogo/2x/googlelogo_color_120x44dp.png"

func TestRenameHandler(t *testing.T) {
	const oldName = "old"
	const newName = "new"

	err := handler.UploadFile(handler.DownloadFolder+oldName, SampleUrl)
	if err != nil {
		t.Fatal("Error while uploading file: ", err.Error())
	}

	req, err := http.NewRequest("PUT", "/rename/"+oldName+"/new/"+newName, nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	handler.NewHandler().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: %v want %v", status, http.StatusOK)
	}

	expected := newName
	if recorder.Body.String() != expected {
		t.Error("handler returned wrong status code: " + recorder.Body.String() + ", want " + expected)
	}

	err = os.Remove(handler.DownloadFolder + newName)
	if err != nil {
		log.Fatal(err)
	}
}

func TestUploadFile(t *testing.T) {

	err := handler.UploadFile("testfile1", "")
	expected := `Get : unsupported protocol scheme ""`
	if err.Error() != expected {
		t.Error("For", `""`, "expected", expected, "got", err.Error())
	}

	err = handler.UploadFile("testfile2", SampleUrl)
	expected = string("<nil>")
	if err != nil {
		t.Error("For", `google pic`, "expected", expected, "got", err.Error())
	}
	err = os.Remove("testfile1")
	if err != nil {
		log.Fatal(err)
	}
	err = os.Remove("testfile2")
	if err != nil {
		log.Fatal(err)
	}
}
