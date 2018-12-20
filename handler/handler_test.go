package handler

import (
	"testing"
)

func TestDownloadFile(t *testing.T) {

	 SampleUrl := "https://www.google.com/images/branding/googlelogo/2x/googlelogo_color_120x44dp.png"

	err := DownloadFile("testfile1", "")
	expected := `Get : unsupported protocol scheme ""`
	if err.Error() != expected{
		t.Error("For", `""`, "expected", expected, "got", err.Error())
	}

	err = DownloadFile("testfile2", SampleUrl)
	expected = string("<nil>")
	if err != nil{
		t.Error("For", `google pic`, "expected", expected, "got", err.Error())
	}

}
