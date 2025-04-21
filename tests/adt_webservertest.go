package tests

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/ecromaneli-golang/http/webserver"
)

var port = 8500

type WebServerTest struct {
	ServerHost    string
	ServerPort    int
	ServerMethod  string
	ServerPattern string
	ServerHandler webserver.Handler

	RequestMethod      string
	RequestContentType string
	RequestPath        string
	RequestHost        string
	RequestPort        int
	RequestBody        []byte
}

func (wst *WebServerTest) SetDefaults() {
	if wst.ServerPort == 0 {
		port++
		wst.ServerPort = port
	}

	if wst.ServerMethod == "" {
		wst.ServerMethod = http.MethodGet
	}

	if wst.ServerPattern == "" {
		wst.ServerPattern = "/"
	}

	if wst.ServerHandler == nil {
		wst.ServerHandler = emptyHandler
	}

	if wst.RequestHost == "" {
		wst.RequestHost = "localhost"
	}

	if wst.RequestPort == 0 {
		wst.RequestPort = wst.ServerPort
	}

	if wst.RequestMethod == "" {
		wst.RequestMethod = http.MethodGet
	}

	if wst.RequestPath == "" {
		wst.RequestPath = "/"
	}
}

func (wst WebServerTest) Do() error {
	_, _, err := wst.DoAndGetDetails()
	return err
}

func (wst WebServerTest) DoAndGetDetails() (req *http.Request, res *http.Response, err error) {

	// Given
	wst.SetDefaults()

	server := webserver.NewServer()
	server.Handle(wst.ServerMethod, wst.ServerPattern, wst.ServerHandler)

	// When
	go func() {
		panic(server.ListenAndServe(wst.ServerHost + ":" + strconv.Itoa(wst.ServerPort)))
	}()

	<-time.After(time.Millisecond)

	var body io.Reader
	if wst.RequestBody != nil {
		body = bytes.NewBuffer(wst.RequestBody)
	}

	// param := make(map[string][]string)
	// param["testBody"] = []string{"valueBody"}

	// res, err = http.PostForm("http://localhost:"+strconv.Itoa(wst.ServerPort)+wst.RequestPath, param)

	req, err = http.NewRequest(wst.RequestMethod, "http://"+wst.RequestHost+":"+strconv.Itoa(wst.RequestPort)+wst.RequestPath, body)

	if err != nil {
		return req, nil, err
	}

	if wst.RequestContentType != "" {
		req.Header.Add(webserver.ContentTypeHeader, wst.RequestContentType)
	}

	res, err = http.DefaultClient.Do(req)

	if err != nil {
		return req, nil, err
	}

	if res.StatusCode != http.StatusOK {
		return req, res, errors.New(res.Status)
	}

	return req, res, nil
}

func emptyHandler(req *webserver.Request, res *webserver.Response) {}
