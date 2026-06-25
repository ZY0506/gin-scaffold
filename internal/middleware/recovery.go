package middleware

import (
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ZY0506/gin-scaffold/internal/pkg/errors"
	"github.com/ZY0506/gin-scaffold/internal/pkg/response"
)

// Recovery 自定义 Recovery 中间件，使用 zap 记录 panic 堆栈
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						s := strings.ToLower(se.Error())
						if strings.Contains(s, "broken pipe") ||
							strings.Contains(s, "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				fields := []zap.Field{
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("client_ip", c.ClientIP()),
				}

				if brokenPipe {
					logger.Error("broken pipe", fields...)
					c.Error(err.(error))
					c.Abort()
					return
				}

				fields = append(fields, zap.Any("error", err), zap.Stack("stack"))
				logger.Error("panic recovered", fields...)
				response.Error(c, http.StatusInternalServerError, errors.ErrInternal, "服务器内部错误")
			}
		}()
		c.Next()
	}
}
