package toolkit

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func TestTools_RandomString(t *testing.T) {
	var testTools Tools
	s := testTools.RandomString(10)

	if len(s) != 10 {
		t.Error("Wrong length random string returnned")
	}
}

var uploadTests = []struct {
	name          string
	allowedTypes  []string
	renamedFile   bool
	errorExpected bool
}{
	{name: "allowed no rename", allowedTypes: []string{"image/jpeg", "image/png"}, renamedFile: false, errorExpected: false},
	{name: "allowed rename", allowedTypes: []string{"image/jpeg", "image/png"}, renamedFile: true, errorExpected: false},
	{name: "not allowed", allowedTypes: []string{"image/jpeg"}, renamedFile: false, errorExpected: true},
}

func TestTools_UploadFiles(t *testing.T) {
	for _, e := range uploadTests {
		// set up a pipe to avoid buffering
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer writer.Close()
			defer wg.Done()
			/// Create the form data filed 'file'
			part, err := writer.CreateFormFile("file", "./testdata/img.png")
			if err != nil {
				t.Error(err)
			}
			f, err := os.Open("./testdata/img.png")
			if err != nil {
				t.Error(err)
			}
			defer f.Close()
			img, _, err := image.Decode(f)
			if err != nil {
				t.Error("error decode image", err)
			}

			err = png.Encode(part, img)
			if err != nil {
				t.Error(err)
			}
		}()

		// read from the pipe which receives data

		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools
		testTools.AllowedFileTypes = e.allowedTypes

		uploadedFiles, err := testTools.UploadFiles(request, "./testdata/uploads/", e.renamedFile)
		if err != nil && !e.errorExpected {
			t.Error(err)
		}

		if !e.errorExpected {
			if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName)); os.IsNotExist(err) {
				t.Errorf("%s: expected file to exist: %s", e.name, err.Error())
			}

			// clean up
			_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName))
		}

		if !e.errorExpected && err != nil {
			t.Errorf("%s: error expected but none received", e.name)
		}

		wg.Wait()
	}
}

func TestTools_UploadOneFiles(t *testing.T) {
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer writer.Close()
		/// Create the form data filed 'file'
		part, err := writer.CreateFormFile("file", "./testdata/img.png")
		if err != nil {
			t.Error(err)
		}
		f, err := os.Open("./testdata/img.png")
		if err != nil {
			t.Error(err)
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		if err != nil {
			t.Error("error decode image", err)
		}

		err = png.Encode(part, img)
		if err != nil {
			t.Error(err)
		}
	}()

	// read from the pipe which receives data

	request := httptest.NewRequest("POST", "/", pr)
	request.Header.Add("Content-Type", writer.FormDataContentType())

	var testTools Tools
	uploadedFiles, err := testTools.UploadOneFile(request, "./testdata/uploads/", true)
	if err != nil {
		t.Error(err)
	}

	if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles.NewFileName)); os.IsNotExist(err) {
		t.Errorf("expected file to exist: %s", err.Error())
	}

	// clean up
	_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles.NewFileName))
}
