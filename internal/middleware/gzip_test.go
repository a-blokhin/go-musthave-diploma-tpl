package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipMiddleware_Compression(t *testing.T) {
	tests := []struct {
		name           string
		acceptEncoding string
		contentType    string
		responseBody   string
		expectGzip     bool
	}{
		{
			name:           "Compress JSON response",
			acceptEncoding: "gzip",
			contentType:    "application/json",
			responseBody:   `{"message":"hello"}`,
			expectGzip:     true,
		},
		{
			name:           "Compress HTML response",
			acceptEncoding: "gzip",
			contentType:    "text/html",
			responseBody:   `<html><body>Hello</body></html>`,
			expectGzip:     true,
		},
		{
			name:           "Don't compress when client doesn't accept gzip",
			acceptEncoding: "",
			contentType:    "application/json",
			responseBody:   `{"message":"hello"}`,
			expectGzip:     false,
		},
		{
			name:           "Don't compress text/plain",
			acceptEncoding: "gzip",
			contentType:    "text/plain",
			responseBody:   "plain text",
			expectGzip:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(GzipMiddleware())

			router.GET("/test", func(c *gin.Context) {
				c.Header("Content-Type", tt.contentType)
				c.String(http.StatusOK, tt.responseBody)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if tt.expectGzip {
				assert.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))
				assert.Equal(t, "Accept-Encoding", resp.Header.Get("Vary"))

				reader, err := gzip.NewReader(resp.Body)
				require.NoError(t, err)
				defer reader.Close()

				decompressed, err := io.ReadAll(reader)
				require.NoError(t, err)
				assert.Equal(t, tt.responseBody, string(decompressed))
			} else {
				assert.Empty(t, resp.Header.Get("Content-Encoding"))

				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.responseBody, string(body))
			}
		})
	}
}

func TestGzipMiddleware_Decompression(t *testing.T) {
	tests := []struct {
		name            string
		contentEncoding string
		requestBody     []byte
		expectedBody    string
		expectedStatus  int
	}{
		{
			name:            "Decompress gzip request",
			contentEncoding: "gzip",
			requestBody:     compressString(`{"message":"hello"}`),
			expectedBody:    `{"message":"hello"}`,
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "Handle uncompressed request",
			contentEncoding: "",
			requestBody:     []byte(`{"message":"hello"}`),
			expectedBody:    `{"message":"hello"}`,
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "Handle invalid gzip data",
			contentEncoding: "gzip",
			requestBody:     []byte("invalid gzip data"),
			expectedBody:    "",
			expectedStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(GzipMiddleware())

			var receivedBody string
			router.POST("/test", func(c *gin.Context) {
				body, _ := io.ReadAll(c.Request.Body)
				receivedBody = string(body)
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("POST", "/test", bytes.NewReader(tt.requestBody))
			if tt.contentEncoding != "" {
				req.Header.Set("Content-Encoding", tt.contentEncoding)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, tt.expectedBody, receivedBody)
			}
		})
	}
}

func compressString(s string) []byte {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write([]byte(s))
	gz.Close()
	return buf.Bytes()
}
