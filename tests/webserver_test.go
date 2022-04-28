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

func TestShouldNotNeedFinalSlash8(t *testing.T) {
	// When
	test := WebServerTest{ServerPattern: "localhost/static1/{o1?}/{o2?}/", RequestPath: "/static1"}

	// Then
	panicIfNotNil(test.Do())
}

func TestShouldNotAcceptWrongDomain(t *testing.T) {
	// When
	test := WebServerTest{ServerPattern: "wronghost/static1/{o1?}/{o2?}/", RequestPath: "/static1"}
	test2 := WebServerTest{ServerPattern: "localhost/static1/{o1?}/{o2?}/", RequestHost: "127.0.0.1", RequestPath: "/static1"}

	// Then
	assert.ErrorContains(t, test.Do(), http.StatusText(http.StatusNotFound))
	assert.ErrorContains(t, test2.Do(), http.StatusText(http.StatusNotFound))
}

func TestShouldAcceptRightDomain(t *testing.T) {
	// When
	test := WebServerTest{ServerPattern: "localhost/static1/{o1?}/{o2?}/", RequestPath: "/static1"}
	test2 := WebServerTest{ServerPattern: "127.0.0.1/static1/{o1?}/{o2?}/", RequestHost: "127.0.0.1", RequestPath: "/static1"}

	// Then
	panicIfNotNil(test.Do())
	panicIfNotNil(test2.Do())
}

func TestShouldReturnNotFound1(t *testing.T) {
	// When
	test := WebServerTest{ServerPattern: "/", RequestPath: "/static1"}

	// Then
	assert.ErrorContains(t, test.Do(), http.StatusText(http.StatusNotFound))
}

func TestShouldReturnNotFound2(t *testing.T) {
	// When
	test := WebServerTest{ServerPattern: "/{id}", RequestPath: "/"}

	// Then
	assert.ErrorContains(t, test.Do(), http.StatusText(http.StatusNotFound))
}

func TestShouldParseParams(t *testing.T) {
	// When
	test := WebServerTest{
		ServerMethod:  http.MethodPost,
		ServerPattern: "{domain}/{pathParam}",

		RequestMethod:      http.MethodPost,
		RequestContentType: webserver.ContentTypeFormUrlEncoded,
		RequestPath:        "/pathValue?param1=value1&param2=value2&param3",
		RequestBody:        []byte("bodyParam=bodyValue"),
	}

	// Then
	test.ServerHandler = func(req *webserver.Request, res *webserver.Response) {
		assert.Equal(t, "localhost", req.Param("domain"))
		assert.Equal(t, "pathValue", req.Param("pathParam"))
		assert.Equal(t, "bodyValue", req.Param("bodyParam"))
		assert.Equal(t, "value1", req.Param("param1"))
		assert.Equal(t, "value2", req.Param("param2"))
		assert.Equal(t, "", req.Param("param3"))
	}

	panicIfNotNil(test.Do())
}

// Issue fixed on 0.3.2
func TestShouldParseDomainParamEvenWithoutPathParam(t *testing.T) {
	// When
	test := WebServerTest{ServerPattern: "{domain}/", RequestPath: "/"}

	// Then
	test.ServerHandler = func(req *webserver.Request, res *webserver.Response) {
		assert.Equal(t, "localhost", req.Param("domain"))
	}

	panicIfNotNil(test.Do())
}

// Issue fixed on 0.3.3: When the token * was passed to isOptional(token), an index out of range [-1] was thrown
func TestShouldNotPanicWhenPathIsGreaterThenPatternAndNextTokenIsShort(t *testing.T) {
	// When
	test := WebServerTest{ServerPattern: "/static1/*/{opt?}", RequestPath: "/static1"}

	// Then
	panicIfNotNil(test.Do())
}

func TestShouldAcceptWildCard(t *testing.T) {
	// When
	test := WebServerTest{ServerPattern: "/**", RequestPath: "/"}

	// Then
	panicIfNotNil(test.Do())
}

func panicIfNotNil(err error) {
	if err != nil {
		panic(err)
	}
}
