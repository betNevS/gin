// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func testRouteOK(method string, t *testing.T) {
	passed := false
	passedAny := false
	r := New()
	r.Any("/test2", func(c *Context) {
		passedAny = true
	})
	r.Handle(method, "/test", func(c *Context) {
		passed = true
	})

	w := performRequest(r, method, "/test")
	assert.True(t, passed)
	assert.Equal(t, w.Code, http.StatusOK)

	performRequest(r, method, "/test2")
	assert.True(t, passedAny)

}

// TestSingleRouteOK tests that POST route is correctly invoked.
func testRouteNotOK(method string, t *testing.T) {
	passed := false
	router := New()
	router.Handle(method, "/test_2", func(c *Context) {
		passed = true
	})

	w := performRequest(router, method, "/test")

	assert.False(t, passed)
	assert.Equal(t, w.Code, http.StatusNotFound)
}

// TestSingleRouteOK tests that POST route is correctly invoked.
func testRouteNotOK2(method string, t *testing.T) {
	passed := false
	router := New()
	var methodRoute string
	if method == "POST" {
		methodRoute = "GET"
	} else {
		methodRoute = "POST"
	}
	router.Handle(methodRoute, "/test", func(c *Context) {
		passed = true
	})

	w := performRequest(router, method, "/test")

	assert.False(t, passed)
	assert.Equal(t, w.Code, http.StatusMethodNotAllowed)
}

func TestRouterGroupRouteOK(t *testing.T) {
	testRouteOK("GET", t)
	testRouteOK("POST", t)
	testRouteOK("PUT", t)
	testRouteOK("PATCH", t)
	testRouteOK("HEAD", t)
	testRouteOK("OPTIONS", t)
	testRouteOK("DELETE", t)
	testRouteOK("CONNECT", t)
	testRouteOK("TRACE", t)
}

// TestSingleRouteOK tests that POST route is correctly invoked.
func TestRouteNotOK(t *testing.T) {
	testRouteNotOK("GET", t)
	testRouteNotOK("POST", t)
	testRouteNotOK("PUT", t)
	testRouteNotOK("PATCH", t)
	testRouteNotOK("HEAD", t)
	testRouteNotOK("OPTIONS", t)
	testRouteNotOK("DELETE", t)
	testRouteNotOK("CONNECT", t)
	testRouteNotOK("TRACE", t)
}

// TestSingleRouteOK tests that POST route is correctly invoked.
func TestRouteNotOK2(t *testing.T) {
	testRouteNotOK2("GET", t)
	testRouteNotOK2("POST", t)
	testRouteNotOK2("PUT", t)
	testRouteNotOK2("PATCH", t)
	testRouteNotOK2("HEAD", t)
	testRouteNotOK2("OPTIONS", t)
	testRouteNotOK2("DELETE", t)
	testRouteNotOK2("CONNECT", t)
	testRouteNotOK2("TRACE", t)
}

// TestContextParamsGet tests that a parameter can be parsed from the URL.
func TestRouteParamsByName(t *testing.T) {
	name := ""
	lastName := ""
	wild := ""
	router := New()
	router.GET("/test/:name/:last_name/*wild", func(c *Context) {
		name = c.Params.ByName("name")
		lastName = c.Params.ByName("last_name")
		wild = c.Params.ByName("wild")

		assert.Equal(t, name, c.ParamValue("name"))
		assert.Equal(t, lastName, c.ParamValue("last_name"))

		assert.Equal(t, name, c.DefaultParamValue("name", "nothing"))
		assert.Equal(t, lastName, c.DefaultParamValue("last_name", "nothing"))
		assert.Equal(t, c.DefaultParamValue("noKey", "default"), "default")
	})

	w := performRequest(router, "GET", "/test/john/smith/is/super/great")

	assert.Equal(t, w.Code, 200)
	assert.Equal(t, name, "john")
	assert.Equal(t, lastName, "smith")
	assert.Equal(t, wild, "/is/super/great")
}

// TestHandleStaticFile - ensure the static file handles properly
func TestRouteStaticFile(t *testing.T) {
	// SETUP file
	testRoot, _ := os.Getwd()
	f, err := ioutil.TempFile(testRoot, "")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("Gin Web Framework")
	f.Close()

	dir, filename := path.Split(f.Name())

	// SETUP gin
	router := New()
	router.Static("/using_static", dir)
	router.StaticFile("/result", f.Name())

	w := performRequest(router, "GET", "/using_static/"+filename)
	w2 := performRequest(router, "GET", "/result")

	assert.Equal(t, w, w2)
	assert.Equal(t, w.Code, 200)
	assert.Equal(t, w.Body.String(), "Gin Web Framework")
	assert.Equal(t, w.HeaderMap.Get("Content-Type"), "text/plain; charset=utf-8")

	w3 := performRequest(router, "HEAD", "/using_static/"+filename)
	w4 := performRequest(router, "HEAD", "/result")

	assert.Equal(t, w3, w4)
	assert.Equal(t, w3.Code, 200)
}

// TestHandleStaticDir - ensure the root/sub dir handles properly
func TestRouteStaticListingDir(t *testing.T) {
	router := New()
	router.StaticFS("/", Dir("./", true))

	w := performRequest(router, "GET", "/")

	assert.Equal(t, w.Code, 200)
	assert.Contains(t, w.Body.String(), "gin.go")
	assert.Equal(t, w.HeaderMap.Get("Content-Type"), "text/html; charset=utf-8")
}

// TestHandleHeadToDir - ensure the root/sub dir handles properly
func TestRouteStaticNoListing(t *testing.T) {
	router := New()
	router.Static("/", "./")

	w := performRequest(router, "GET", "/")

	assert.Equal(t, w.Code, 404)
	assert.NotContains(t, w.Body.String(), "gin.go")
}

func TestRouterMiddlewareAndStatic(t *testing.T) {
	router := New()
	static := router.Group("/", func(c *Context) {
		c.Writer.Header().Add("Last-Modified", "Mon, 02 Jan 2006 15:04:05 MST")
		c.Writer.Header().Add("Expires", "Mon, 02 Jan 2006 15:04:05 MST")
		c.Writer.Header().Add("X-GIN", "Gin Framework")
	})
	static.Static("/", "./")

	w := performRequest(router, "GET", "/gin.go")

	assert.Equal(t, w.Code, 200)
	assert.Contains(t, w.Body.String(), "package gin")
	assert.Equal(t, w.HeaderMap.Get("Content-Type"), "text/plain; charset=utf-8")
	assert.NotEqual(t, w.HeaderMap.Get("Last-Modified"), "Mon, 02 Jan 2006 15:04:05 MST")
	assert.Equal(t, w.HeaderMap.Get("Expires"), "Mon, 02 Jan 2006 15:04:05 MST")
	assert.Equal(t, w.HeaderMap.Get("x-GIN"), "Gin Framework")
}

func TestRouteNotAllowed(t *testing.T) {
	router := New()

	router.POST("/path", func(c *Context) {})
	w := performRequest(router, "GET", "/path")
	assert.Equal(t, w.Code, http.StatusMethodNotAllowed)

	router.NoMethod(func(c *Context) {
		c.String(http.StatusTeapot, "responseText")
	})
	w = performRequest(router, "GET", "/path")
	assert.Equal(t, w.Body.String(), "responseText")
	assert.Equal(t, w.Code, http.StatusTeapot)
}

func TestRouterNotFound(t *testing.T) {
	router := New()
	router.GET("/path", func(c *Context) {})
	router.GET("/dir/", func(c *Context) {})
	router.GET("/", func(c *Context) {})

	testRoutes := []struct {
		route  string
		code   int
		header string
	}{
		{"/path/", 301, "map[Location:[/path]]"},   // TSR -/
		{"/dir", 301, "map[Location:[/dir/]]"},     // TSR +/
		{"", 301, "map[Location:[/]]"},             // TSR +/
		{"/PATH", 301, "map[Location:[/path]]"},    // Fixed Case
		{"/DIR/", 301, "map[Location:[/dir/]]"},    // Fixed Case
		{"/PATH/", 301, "map[Location:[/path]]"},   // Fixed Case -/
		{"/DIR", 301, "map[Location:[/dir/]]"},     // Fixed Case +/
		{"/../path", 301, "map[Location:[/path]]"}, // CleanPath
		{"/nope", 404, ""},                         // NotFound
	}
	for _, tr := range testRoutes {
		w := performRequest(router, "GET", tr.route)
		assert.Equal(t, w.Code, tr.code)
		if w.Code != 404 {
			assert.Equal(t, fmt.Sprint(w.Header()), tr.header)
		}
	}

	// Test custom not found handler
	var notFound bool
	router.NoRoute(func(c *Context) {
		c.AbortWithStatus(404)
		notFound = true
	})
	w := performRequest(router, "GET", "/nope")
	assert.Equal(t, w.Code, 404)
	assert.True(t, notFound)

	// Test other method than GET (want 307 instead of 301)
	router.PATCH("/path", func(c *Context) {})
	w = performRequest(router, "PATCH", "/path/")
	assert.Equal(t, w.Code, 307)
	assert.Equal(t, fmt.Sprint(w.Header()), "map[Location:[/path]]")

	// Test special case where no node for the prefix "/" exists
	router = New()
	router.GET("/a", func(c *Context) {})
	w = performRequest(router, "GET", "/")
	assert.Equal(t, w.Code, 404)
}