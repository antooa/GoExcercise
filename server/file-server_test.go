package server_test

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

func TestUploadDelete(t *testing.T) {
	u := url.URL{
		Scheme:   "http",
		Host:     "localhost:8080",
		Path:     "/upload",
		RawQuery: url.Values{"uri": []string{"https://www.google.com/images/branding/googlelogo/2x/googlelogo_color_120x44dp.png"}}.Encode(),
	}
	resp, err := http.Get(u.String())
	if err != nil {
		t.Errorf("Couldn't GET : %v; Got %v", u.String(), err)
	}
	defer resp.Body.Close()

	if status := resp.StatusCode; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: %v want %v", status, http.StatusOK)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		t.Errorf("Unable to read body")
	}
	filename := string(bodyBytes)

	newName := "name"

	renUrl := url.URL{
		Scheme:   "http",
		Host:     "localhost:8080",
		Path:     "/rename/"+filename+"/new/"+newName,
	}
	renResp, err := http.Get(renUrl.String())
	if err != nil {
		t.Errorf("Couldn't GET : %v", renUrl.String())
	}

	if status := renResp.StatusCode; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: %v want %v", status, http.StatusOK)
	}

	t.Log(filename)
	delUrl := url.URL{
		Scheme:   "http",
		Host:     "localhost:8080",
		Path:     "/delete/"+newName,
	}

	deleteResp, err := http.Get(delUrl.String())
	if err != nil {
		t.Errorf("Couldn't GET : %v", delUrl.String())
	}

	if status := deleteResp.StatusCode; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: %v want %v", status, http.StatusOK)
	}
}

