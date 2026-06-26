package middleware

import "github.com/gin-gonic/gin"

// OperationLogFunc 操作日志写入函数，由上层注入具体实现
type OperationLogFunc func(c *gin.Context)

// WithOperationLog 返回一个中间件，在请求完成后记录操作日志
func WithOperationLog(writeLog OperationLogFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 仅记录管理端写操作
		if c.Request.Method != "POST" && c.Request.Method != "PUT" && c.Request.Method != "PATCH" && c.Request.Method != "DELETE" {
			return
		}

		writeLog(c)
	}
}
