package error_log

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

func NewErrorLogMiddleWareFunc(l Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			httpStatus := c.Writer.Status()
			if httpStatus == 404 {
				l.Error(fmt.Sprintf("http status error:%d", httpStatus))
			} else if httpStatus > 499 {
				// 只打印服务端错误
				l.Error(fmt.Sprintf("http status error:%d", httpStatus))
			}
		}()
		c.Next()
	}
}
