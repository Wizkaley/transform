package main

import (
	"bytes"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"transform/primitive"

	"github.com/stretchr/testify/assert"
)

func TestM(m *testing.T) {
	tempListenAndServe := listenAndServe

	defer func() {
		listenAndServe = tempListenAndServe
	}()
	listenAndServe = func(addr string, handler http.Handler) error {
		panic("Failed")
	}

	assert.PanicsWithValuef(m, "Failed", main, "Expected %v got %v", "Failed", main)
}

var tst = []struct {
	in   []string
	stat bool
}{
	{[]string{"txt", "fine"}, false},
	{[]string{"creteFile", "/"}, true},
	{[]string{"/+/", "/"}, true},
}

func TestTempFile(t *testing.T) {
	for _, item := range tst {
		_, err := tempfile(item.in[0], item.in[1])
		assert.Equalf(t, item.stat, err != nil, "Expected %v but got %v", item.stat, err != nil)
	}
}

func cont() http.Handler {
	srv := http.NewServeMux()
	srv.HandleFunc("/", BasePath)
	srv.HandleFunc("/upload/", Upload)
	srv.HandleFunc("/modify/", Modify)
	return srv
}
func TestBaseURL(t *testing.T) {

	server := httptest.NewServer(cont())

	req, err := http.NewRequest("GET", server.URL+"/", nil)
	if err != nil {
		log.Printf("Error while creating request :: %v\n", err)
		return
	}

	//server.ServeHTTP(res, req)

	res, err := http.DefaultClient.Do(req)
	// if err != nil {
	// 	log.Printf("Erro while creating request :: %v\n", err)
	// 	return
	// }

	if res.StatusCode != 200 {
		t.Errorf("Expected %v but got %v", http.StatusOK, res.Status)
	}

}

var testModifyInput = []struct {
	url    string
	status int
}{
	{"/modify/307198887..jpeg?mode=3", 200},
	{"/modify/invalid.jpeg?mode=2&n=10", 400},
	{"/modify/307198887..jpeg", 200},
	{"/modify/307198887..jpeg?mode=abc", 400},
	{"/modify/307198887..jpeg?mode=3&n=abc", 400},
}

func TestModify(t *testing.T) {
	for _, item := range testModifyInput {
		srv := httptest.NewServer(cont())
		req, _ := http.NewRequest("GET", srv.URL+item.url, nil)
		// rr := httptest.NewRecorder()
		res, _ := http.DefaultClient.Do(req)
		assert.Equalf(t, item.status, res.StatusCode, "Expected %d but got %d", item.status, res.StatusCode)
	}
}

var testUploadhandlerInput = []struct {
	testCase string
	param    string
	filename string
	stat     int
}{
	{"tc1", "image", "img/rnm.png", 302},
	{"tc2", "non-image", "img/rnm.png", 400},
	{"tc3", "image", "img/rnm.png", 500},
	{"tc4", "image", "img/rnm.png", 500},
}

func TestUploadHanlder(t *testing.T) {
	tempIOCopyVar := IOCopyVar
	tempCreateTempFileVar := tempFileVar
	for _, item := range testUploadhandlerInput {

		imgFile, err := os.OpenFile(item.filename, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			t.Errorf("Error While Opening Image file : %v", err)
		}

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile(item.param, imgFile.Name())
		if err != nil {
			t.Errorf("Error while creating form file : %v", err)
		}

		_, err = io.Copy(part, imgFile)
		if err != nil {
			t.Errorf("Error while copying file : %v", err)
		}

		err = writer.Close()
		if err != nil {
			t.Errorf("Error while closing writer : %v", err)
		}

		//srv := httptest.NewServer(cont())

		req, _ := http.NewRequest("POST", "/upload", body)
		rr := httptest.NewRecorder()
		if item.testCase == "tc3" {
			tempFileVar = func(prefix, ext string) (*os.File, error) {
				return nil, errors.New("main: failed to create temp input file")
			}
		}

		if item.testCase == "tc4" {
			IOCopyVar = func(dst io.Writer, src io.Reader) (written int64, err error) {
				return -1, errors.New("IO Copy Failed")
			}
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())
		//res, err := http.DefaultClient.Do(req)
		handler := http.HandlerFunc(Upload)
		handler.ServeHTTP(rr, req)
		assert.Equalf(t, item.stat, rr.Code, "Expected %v but got %v", item.stat, rr.Code)
		IOCopyVar = tempIOCopyVar
		tempFileVar = tempCreateTempFileVar
		imgFile.Close()
	}
}

var tstRenderInput = []struct {
	testCase     string
	inp          []string
	expectedStat int
}{
	{"tc1", []string{"tmp/rnm.png", "png"}, 200},
	{"tc2", []string{"tmp/rnm.png", "png"}, 500},
	//{"tc3", []string{"tmp/rnm.png", "png"}, 200},
	{"tc4", []string{"tmp/rnm.png", "png"}, 200},
}

func TestRenderNumShapes(t *testing.T) {

	tempTplExecuteBool := tplExecuteBool
	for _, item := range tstRenderInput {
		req, err := http.NewRequest("GET", "localhost:5000", nil)
		if err != nil {
			t.Errorf("Error while creatinhg test request : %v ", err)
		}

		rr := httptest.NewRecorder()
		file, err := os.OpenFile(item.inp[0], os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			t.Errorf("Error while opening test file : %v", err)
		}
		if item.testCase == "tc2" {
			file.Close()
		}

		if item.testCase == "tc4" {
			tplExecuteBool = true
		}

		renderNumShapeChoices(rr, req, file, item.inp[1], primitive.ModeCircle)
		assert.Equalf(t, item.expectedStat, rr.Code, "Expected %v but got %v", item.expectedStat, rr.Code)

		tplExecuteBool = tempTplExecuteBool
		file.Close()
	}
}

func TestRenderMode(t *testing.T) {
	tempTplExecuteBool := tplExecuteBool
	for _, item := range tstRenderInput {
		req, err := http.NewRequest("GET", "localhost:5000", nil)
		if err != nil {
			t.Errorf("Error while creatinhg test request : %v ", err)
		}

		rr := httptest.NewRecorder()
		file, err := os.OpenFile(item.inp[0], os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			t.Errorf("Error while opening test file : %v", err)
		}
		if item.testCase == "tc2" {
			file.Close()
		}

		if item.testCase == "tc4" {
			tplExecuteBool = true
		}

		renderModeChoices(rr, req, file, item.inp[1])
		assert.Equalf(t, item.expectedStat, rr.Code, "Expected %v but got %v", item.expectedStat, rr.Code)

		tplExecuteBool = tempTplExecuteBool
		file.Close()
	}
}

func TestGenerateImages(t *testing.T) {
	var tstInp = []struct {
		testCase string
		input    []string
		isErr    bool
	}{
		{"tc1", []string{"tmp/rnm.png", "png"}, false},
		{"tc2", []string{"tmp/rnmasd.png", "png"}, true},
		{"tc3", []string{"tmp/rasda.png", "/"}, true},
		{"tc4", []string{"tmp/rnm.png", "png"}, true},
	}

	option := []genOpts{
		{N: 10, M: primitive.ModeBeziers},
	}
	for _, item := range tstInp {
		file, err := os.OpenFile(item.input[0], os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			t.Errorf("Error While Opening Test file : %v", err)
		}

		if item.testCase == "tc4" {
			file.Close()
		}

		_, err = genImages(file, item.input[1], option...)
		assert.Equalf(t, item.isErr, err != nil, "Expected %v but got %v", item.isErr, err != nil)
		file.Close()

	}

}

func TestGenImage(t *testing.T) {

	var test = []struct {
		testCase string
		in       []string
		isErr    bool
	}{
		{"tc1", []string{"tmp/rnm.png", "png"}, false},
		{"tc2", []string{"tmp/rnm.png", "png"}, true},
		{"tc3", []string{"tmp/rnm.png", "png"}, true},
		{"tc4", []string{"tmp/rnm.png", "png"}, true},
	}

	tempVarName := TempName
	tempIOCopyVar := IOCopyVar

	for _, item := range test {

		file, err := os.OpenFile(item.in[0], os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			t.Errorf("Error while opening opening test file : %v", err)
		}

		if item.testCase == "tc2" {
			TempName = "/+/"
		}
		if item.testCase == "tc3" {
			IOCopyVar = func(dst io.Writer, src io.Reader) (written int64, err error) {
				return -1, errors.New("IO Copy failed")
			}
		}

		if item.testCase == "tc4" {
			file.Close()
		}
		_, err = genImage(file, item.in[1], 5, primitive.ModeCombo)
		assert.Equalf(t, item.isErr, err != nil, "Expected %v but got %v", item.isErr, err != nil)
		TempName = tempVarName
		IOCopyVar = tempIOCopyVar
		file.Close()
	}

}
