package middleware

import (
	"rvcs/internal/api/response"

	"github.com/gin-gonic/gin"
)

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				response.InternalError(c, "internal server error")
				c.Abort()
			}
		}()
		c.Next()
	}
}
