package fileservingTransport

import (
	"net/http"
	"presentator/internal/core/entity"

	"github.com/gin-gonic/gin"
)

func response(c *gin.Context, res entity.Response) {

	if res.Status == 0 {
		if res.Err != nil {
			code := entity.FindStatus(res.Err)
			c.JSON(code, res.Data)
			return
		}

		code := http.StatusOK
		c.JSON(code, res.Data)
		return
	}

	c.JSON(res.Status, res.Data)
	return
}
