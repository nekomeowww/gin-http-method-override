package ginhttpmethodoverride

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("Each", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		type testResponse struct {
			Method   string `json:"method"`
			Override string `json:"override"`
		}

		handler := func(c *gin.Context) {
			method := c.Request.Header.Get(XHTTPMethodOverrideHeader)
			c.JSON(http.StatusOK, testResponse{Method: c.Request.Method, Override: method})
		}

		mHandlers := map[string]map[string]gin.HandlerFunc{
			http.MethodGet:     {"/get": handler},
			http.MethodPost:    {"/post": handler},
			http.MethodPatch:   {"/patch": handler},
			http.MethodPut:     {"/put": handler},
			http.MethodDelete:  {"/delete": handler},
			http.MethodHead:    {"/head": handler},
			http.MethodOptions: {"/options": handler},
			http.MethodConnect: {"/connect": handler},
			http.MethodTrace:   {"/trace": handler},
		}

		router := gin.Default()
		// use the middleware
		router.Use(New(router))

		for method, handlers := range mHandlers {
			for endpoint, handler := range handlers {
				router.Handle(method, endpoint, handler)
			}
		}

		for method, handlers := range mHandlers {
			for endpoint := range handlers {
				// use recorder to capture the response
				recorder := httptest.NewRecorder()

				// all the request are made with the POST method
				request, err := http.NewRequest(http.MethodPost, endpoint, nil)
				require.NoError(err)

				// override method through the X-HTTP-Method-Override sheader
				request.Header = http.Header{}
				request.Header.Set(XHTTPMethodOverrideHeader, method)

				// start the server
				router.ServeHTTP(recorder, request)

				require.Equal(http.StatusOK, recorder.Code)

				var res testResponse
				// unmarshal the response
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(err)

				assert.Equal(method, res.Method)
				assert.Equal(method, res.Override)
			}
		}
	})

	t.Run("SameEndpointWithDifferentMethod", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		type testResponse struct {
			Endpoint string `json:"endpoint"`
			Method   string `json:"method"`
			Override string `json:"override"`
		}

		handler := func(c *gin.Context) {
			method := c.Request.Header.Get(XHTTPMethodOverrideHeader)
			c.JSON(http.StatusOK, testResponse{Endpoint: c.Request.URL.String(), Method: c.Request.Method, Override: method})
		}

		mHandlers := map[string]map[string]gin.HandlerFunc{
			"/test": {
				http.MethodGet:     handler,
				http.MethodPost:    handler,
				http.MethodPatch:   handler,
				http.MethodPut:     handler,
				http.MethodDelete:  handler,
				http.MethodHead:    handler,
				http.MethodOptions: handler,
				http.MethodConnect: handler,
				http.MethodTrace:   handler,
			},
		}

		router := gin.Default()
		// use the middleware
		router.Use(New(router))

		for endpoint, handlers := range mHandlers {
			for method, handler := range handlers {
				router.Handle(method, endpoint, handler)
			}
		}

		for endpoint, handlers := range mHandlers {
			for method := range handlers {
				// use recorder to capture the response
				recorder := httptest.NewRecorder()

				// all the request are made with the POST method
				request, err := http.NewRequest(http.MethodPost, endpoint, nil)
				require.NoError(err)

				// override method through the X-HTTP-Method-Override sheader
				request.Header = http.Header{}
				request.Header.Set(XHTTPMethodOverrideHeader, method)

				// start the server
				router.ServeHTTP(recorder, request)

				require.Equal(http.StatusOK, recorder.Code)

				var res testResponse
				// unmarshal the response
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(err)

				assert.Equal(endpoint, res.Endpoint)
				assert.Equal(method, res.Method)
				assert.Equal(method, res.Override)
			}
		}
	})

	t.Run("MiddlewaresSideEffects", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		newMiddlewares := func(c *gin.Context) {
			c.Set("test_key", "test")
		}

		type testResponse struct {
			Key      string `json:"key"`
			Method   string `json:"method"`
			Override string `json:"override"`
		}

		handler := func(c *gin.Context) {
			method := c.Request.Header.Get(XHTTPMethodOverrideHeader)
			key, _ := c.Get("test_key")
			keyStr, _ := key.(string)
			c.JSON(http.StatusOK, testResponse{Key: keyStr, Method: c.Request.Method, Override: method})
		}

		router := gin.Default()
		// use the middleware
		router.Use(New(router))
		router.Use(newMiddlewares)

		mHandlers := map[string]map[string]gin.HandlerFunc{
			http.MethodGet:     {"/get": handler},
			http.MethodPost:    {"/post": handler},
			http.MethodPatch:   {"/patch": handler},
			http.MethodPut:     {"/put": handler},
			http.MethodDelete:  {"/delete": handler},
			http.MethodHead:    {"/head": handler},
			http.MethodOptions: {"/options": handler},
			http.MethodConnect: {"/connect": handler},
			http.MethodTrace:   {"/trace": handler},
		}

		for method, handlers := range mHandlers {
			for endpoint, handler := range handlers {
				router.Handle(method, endpoint, handler)
			}
		}

		for method, handlers := range mHandlers {
			for endpoint := range handlers {
				// use recorder to capture the response
				recorder := httptest.NewRecorder()

				// all the request are made with the POST method
				request, err := http.NewRequest(http.MethodPost, endpoint, nil)
				require.NoError(err)

				// override method through the X-HTTP-Method-Override sheader
				request.Header = http.Header{}
				request.Header.Set(XHTTPMethodOverrideHeader, method)

				// start the server
				router.ServeHTTP(recorder, request)

				require.Equal(http.StatusOK, recorder.Code)

				var res testResponse
				// unmarshal the response
				err = json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(err)

				assert.Equal("test", res.Key)
				assert.Equal(method, res.Method)
				assert.Equal(method, res.Override)
			}
		}
	})

	t.Run("OnlyWorksForPOSTMethod", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		type testResponse struct {
			Method   string `json:"method"`
			Override string `json:"override"`
		}

		handler := func(c *gin.Context) {
			method := c.Request.Header.Get(XHTTPMethodOverrideHeader)
			c.JSON(http.StatusOK, testResponse{Method: c.Request.Method, Override: method})
		}

		router := gin.Default()
		// use the middleware
		router.Use(New(router))

		mHandlers := map[string]map[string]gin.HandlerFunc{
			http.MethodGet:     {"/get": handler},
			http.MethodPost:    {"/post": handler},
			http.MethodPatch:   {"/patch": handler},
			http.MethodPut:     {"/put": handler},
			http.MethodDelete:  {"/delete": handler},
			http.MethodHead:    {"/head": handler},
			http.MethodOptions: {"/options": handler},
			http.MethodConnect: {"/connect": handler},
			http.MethodTrace:   {"/trace": handler},
		}

		for method, handlers := range mHandlers {
			for endpoint, handler := range handlers {
				router.Handle(method, endpoint, handler)
			}
		}

		for method, handlers := range mHandlers {
			for endpoint := range handlers {
				// use recorder to capture the response
				recorder := httptest.NewRecorder()

				// all the request are made with the POST method
				request, err := http.NewRequest(http.MethodPut, endpoint, nil)
				require.NoError(err)

				// override method through the X-HTTP-Method-Override sheader
				request.Header = http.Header{}
				request.Header.Set(XHTTPMethodOverrideHeader, method)

				// start the server
				router.ServeHTTP(recorder, request)

				// success if the request was a PUT request
				if method == http.MethodPut {
					assert.Equal(http.StatusOK, recorder.Code)
				} else {
					assert.Equal(http.StatusNotFound, recorder.Code)
				}
			}
		}
	})
}
