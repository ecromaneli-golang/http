package tests

import (
	"net/http"
	"testing"

	"github.com/ecromaneli-golang/http/webserver"
	"github.com/stretchr/testify/assert"
)

func TestShouldHandleComplexURL(t *testing.T) {
	// When
	test := WebServerTest{
		ServerPattern: "/static1/static2/{p1}/static3/*/{p2}/{o?}/**",
		RequestPath:   "/static1/static2/param1/static3/anything/param2/optional/anything2/anything3/anything4",
	}

	// Then
	test.ServerHandler = func(req *webserver.Request, res *webserver.Response) {
		assert.Equal(t, "param1", req.Param("p1"))
		assert.Equal(t, "param2", req.Param("p2"))
		assert.Equal(t, "optional", req.Param("o"))
	}

	panicIfNotNil(test.Do())
}

func TestShouldNotNeedOptionalParameters(t *testing.T) {
	// When
	test := WebServerTest{ServerPattern: "/static1/{p1}/{o1?}/{o2?}", RequestPath: "/static1/param1"}

	// Then
	test.ServerHandler = func(req *webserver.Request, res *webserver.Response) {
		assert.Equal(t, "param1", req.Param("p1"))
		assert.Equal(t, "", req.Param("o1"))
		assert.Equal(t, "", req.Param("o2"))
	}

	panicIfNotNil(test.Do())
}

func TestShouldNotNeedFinalSlash1(t *testing.T) {
	// When
	test := WebServerTest{ServerPattern: "/static1/static2/", RequestPath: "/static1/static2"}

	// Then
	panicIfNotNil(test.Do())
}

func TestShouldNotNeedFinalSlash2(t *testing.T) {
	// When
	test := WebServerTest{ServerPattern: "/static1/static2", RequestPath: "/static1/static2/"}

	// Then
	panicIfNotNil(test.Do())
}

func TestShouldNotNeedFinalSlash3(t *testing.T) {
	// When
	test := WebServerTest{ServerPattern: "/{p1}/static1/", RequestPath: "/param1/static1"}

	// Then
	panicIfNotNil(test.Do())
}

func TestShouldNotNeedFinalSlash4(t *testing.T) {
	// When
	test := WebServerTest{ServerPattern: "/{p1}/static1", RequestPath: "/param1/static1/"}

	// Then
	panicIfNotNil(test.Do())
}

func TestShouldNotNeedFinalSlash5(t *testing.T) {
	// When
	test := WebServerTest{ServerPattern: "/static1/{p1}/", RequestPath: "/static1/param1"}

	// Then
	panicIfNotNil(test.Do())
}

func TestShouldNotNeedFinalSlash6(t *testing.T) {
	// When
	test := WebServerTest{ServerPattern: "/static1/{p1}", RequestPath: "/static1/param1/"}

	// Then
	panicIfNotNil(test.Do())
}

func TestShouldNotNeedFinalSlash7(t *testing.T) {
	// When
	test := WebServerTest{ServerPattern: "/static1/{o1?}/{o2?}/", RequestPath: "/static1"}

	// Then
	panicIfNotNil(test.Do())
}

func TestShouldReturnNotFound(t *testing.T) {
	// When
	test := WebServerTest{ServerPattern: "/", RequestPath: "/static1"}

	// Then
	assert.ErrorContains(t, test.Do(), http.StatusText(http.StatusNotFound))
}

func TestShouldReturnBadRequest(t *testing.T) {
	// When
	test := WebServerTest{ServerPattern: "/{id}", RequestPath: "/"}

	// Then
	assert.ErrorContains(t, test.Do(), http.StatusText(http.StatusBadRequest))
}

func TestShouldParseParams(t *testing.T) {
	// When
	test := WebServerTest{
		ServerMethod:  http.MethodPost,
		ServerPattern: "/{pathParam}",

		RequestMethod:      http.MethodPost,
		RequestContentType: webserver.ContentTypeFormUrlEncoded,
		RequestPath:        "/pathValue?param1=value1&param2=value2&param3",
		RequestBody:        []byte("bodyParam=bodyValue"),
	}

	// Then
	test.ServerHandler = func(req *webserver.Request, res *webserver.Response) {
		assert.Equal(t, "pathValue", req.Param("pathParam"))
		assert.Equal(t, "bodyValue", req.Param("bodyParam"))
		assert.Equal(t, "value1", req.Param("param1"))
		assert.Equal(t, "value2", req.Param("param2"))
		assert.Equal(t, "", req.Param("param3"))
	}

	panicIfNotNil(test.Do())
}

func TestShouldRefuseWrongHost(t *testing.T) {
	// When
	test1 := WebServerTest{
		ServerHost:  "localhostwrong",
		RequestHost: "localhost",
	}

	test2 := WebServerTest{
		ServerHost:  "localhost",
		RequestHost: "localhostwrong",
	}

	test3 := WebServerTest{
		ServerHost:  "localhost",
		RequestHost: "localhost",
	}

	assert.NotNil(t, test1.Do())
	assert.NotNil(t, test2.Do())
	panicIfNotNil(test3.Do())
}

/*
func Test(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("localhost", func(w http.ResponseWriter, r *http.Request) {
		t.FailNow()
	})
	go http.ListenAndServe("localhost:8080", mux)

	req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080/", nil)
	http.DefaultClient.Do(req)
}
*/

func panicIfNotNil(err error) {
	if err != nil {
		panic(err)
	}
}
