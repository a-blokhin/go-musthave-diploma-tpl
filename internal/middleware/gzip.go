package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
			reader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid gzip data"})
				return
			}
			defer reader.Close()

			decompressed, err := io.ReadAll(reader)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to decompress data"})
				return
			}

			c.Request.Body = io.NopCloser(bytes.NewReader(decompressed))
		}

		acceptEncoding := c.GetHeader("Accept-Encoding")
		if !strings.Contains(acceptEncoding, "gzip") {
			c.Next()
			return
		}

		originalWriter := c.Writer

		var buf bytes.Buffer

		c.Writer = &bodyWriter{
			ResponseWriter: originalWriter,
			body:           &buf,
		}

		c.Next()

		contentType := c.Writer.Header().Get("Content-Type")
		shouldCompress := strings.Contains(contentType, "application/json") ||
			strings.Contains(contentType, "text/html")

		c.Writer = originalWriter

		if shouldCompress && c.Writer.Status() < 300 && c.Writer.Status() >= 200 {
			c.Header("Content-Encoding", "gzip")
			c.Header("Vary", "Accept-Encoding")

			gz := gzip.NewWriter(c.Writer)
			gz.Write(buf.Bytes())
			gz.Close()
		} else {
			c.Writer.Write(buf.Bytes())
		}
	}
}

type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w *bodyWriter) WriteString(s string) (int, error) {
	return w.body.WriteString(s)
}
