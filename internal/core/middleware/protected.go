package middleware

import (
	"net/http"
	"presentator/internal/features/auth/token"

	"github.com/gin-gonic/gin"
)

func ProtectedAdmin(j *JWT.ServingJWT) gin.HandlerFunc {
	return func(c *gin.Context) {

		cookie, err := c.Cookie("Pres-Access")
		if err != nil {
			c.Redirect(http.StatusSeeOther, "/auth/admin.html")
			c.Abort()
			return
		}

		token, ok := j.ValidateToken(cookie)
		if !ok {
			c.Redirect(http.StatusSeeOther, "/auth/admin.html")
			c.Abort()
			return
		}

		ok = j.CheckAccess(c.Request.Context(), token, c.ClientIP(), "admin")
		if !ok {
			c.Redirect(http.StatusSeeOther, "/auth/admin.html")
			c.Abort()
			return
		}

		c.Next()
	}
}

func Protected(j *JWT.ServingJWT) gin.HandlerFunc {
	return func(c *gin.Context) {

		cookie, err := c.Cookie("Pres-Access")
		if err != nil {
			c.Redirect(http.StatusSeeOther, "/auth")
			c.Abort()
			return
		}

		token, ok := j.ValidateToken(cookie)
		if !ok {
			c.Redirect(http.StatusSeeOther, "/auth")
			c.Abort()
			return
		}

		ok = j.CheckAccess(c.Request.Context(), token, c.ClientIP(), "guest")
		if !ok {
			c.Redirect(http.StatusSeeOther, "/auth")
			c.Abort()
			return
		}

		c.Next()
	}
}
