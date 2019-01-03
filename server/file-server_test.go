package server_test

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

// TestServer is a test for all the functions (except download), provided by server
func TestServer(t *testing.T) {
	u := url.URL{
		Scheme:   "http",
		Host:     "localhost:8080",
		Path:     "/upload",
		RawQuery: url.Values{"uri": []string{"https://www.google.com/images/branding/googlelogo/2x/googlelogo_color_120x44dp.png"}}.Encode(),
	}
	resp, err := http.Get(u.String())
	if err != nil {
		t.Fatalf("Couldn't GET : %v; Got %v", u.String(), err)
	}
	defer resp.Body.Close()

	if status := resp.StatusCode; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: %v want %v", status, http.StatusOK)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		t.Fatalf("Unable to read body")
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
		t.Fatalf("Couldn't GET : %v", renUrl.String())
	}
	defer renResp.Body.Close()

	if status := renResp.StatusCode; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: %v want %v", status, http.StatusOK)
	}

	t.Log(filename)
	delUrl := url.URL{
		Scheme:   "http",
		Host:     "localhost:8080",
		Path:     "/delete/"+newName,
	}

	deleteResp, err := http.Get(delUrl.String())
	if err != nil {
		t.Fatalf("Couldn't GET : %v", delUrl.String())
	}
	defer deleteResp.Body.Close()

	if status := deleteResp.StatusCode; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: %v want %v", status, http.StatusOK)
	}
}

