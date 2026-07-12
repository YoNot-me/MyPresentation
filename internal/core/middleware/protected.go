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

		claims := token.Claims.(*JWT.JWT)
		if !j.IsExist(c.Request.Context(), "sess:"+claims.ID) {
			c.Redirect(http.StatusSeeOther, "/auth/admin.html")
			c.Abort()
			return
		}

		ok = j.CheckAdminAccess(token, "sess:"+claims.ID, "admin")
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

		claims := token.Claims.(*JWT.JWT)
		if !j.IsExist(c.Request.Context(), "sess:"+claims.ID) {
			c.Redirect(http.StatusSeeOther, "/auth/admin.html")
			c.Abort()
			return
		}

		ok = j.CheckAccess(token, "sess:"+claims.ID)
		if !ok {
			c.Redirect(http.StatusSeeOther, "/auth")
			c.Abort()
			return
		}

		c.Next()
	}
}
