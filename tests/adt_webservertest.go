package tests

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/ecromaneli-golang/http/webserver"
)

var port = 8000

type WebServerTest struct {
	ServerPort    int
	ServerMethod  string
	ServerPattern string
	ServerHandler webserver.Handler

	RequestMethod      string
	RequestContentType string
	RequestPath        string
	RequestBody        []byte
}

func (this *WebServerTest) SetDefaults() {
	if this.ServerPort == 0 {
		port++
		this.ServerPort = port
	}

	if this.ServerMethod == "" {
		this.ServerMethod = http.MethodGet
	}

	if this.ServerPattern == "" {
		this.ServerPattern = "/"
	}

	if this.ServerHandler == nil {
		this.ServerHandler = emptyHandler
	}

	if this.RequestMethod == "" {
		this.RequestMethod = http.MethodGet
	}

	if this.RequestPath == "" {
		this.RequestPath = "/"
	}
}

func (this WebServerTest) Do() error {
	_, _, err := this.DoAndGetDetails()
	return err
}

func (this WebServerTest) DoAndGetDetails() (req *http.Request, res *http.Response, err error) {

	// Given
	this.SetDefaults()
	go webserver.NewServer().Handle(this.ServerMethod, this.ServerPattern, this.ServerHandler).ListenAndServe(":" + strconv.Itoa(this.ServerPort))

	// When
	var body io.Reader
	if this.RequestBody != nil {
		body = bytes.NewBuffer(this.RequestBody)
	}

	// param := make(map[string][]string)
	// param["testBody"] = []string{"valueBody"}

	// res, err = http.PostForm("http://localhost:"+strconv.Itoa(this.ServerPort)+this.RequestPath, param)

	req, err = http.NewRequest(this.RequestMethod, "http://localhost:"+strconv.Itoa(this.ServerPort)+this.RequestPath, body)

	if err != nil {
		return req, nil, err
	}

	if this.RequestContentType != "" {
		req.Header.Add("Content-Type", this.RequestContentType)
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
