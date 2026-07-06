package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func MaximumSize() gin.HandlerFunc {
	return func(c *gin.Context) {

		const MaxBytes = 200 << 20
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBytes)
		c.Next()
	}
}
