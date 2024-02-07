package logger

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

type AccessLogBuilder struct {
	logFunc       func(ctx context.Context, al AccessLog)
	allowReqBody  bool
	allowRespBody bool
}

func NewAccessLogBuilder(fn func(ctx context.Context, al AccessLog)) *AccessLogBuilder {
	return &AccessLogBuilder{
		logFunc: fn,
		// 默认不打印
		allowReqBody:  false,
		allowRespBody: false,
	}
}

func (b *AccessLogBuilder) AllowReqBody() *AccessLogBuilder {
	b.allowReqBody = true
	return b
}

func (b *AccessLogBuilder) AllowRespBody() *AccessLogBuilder {
	b.allowRespBody = true
	return b
}

func (b *AccessLogBuilder) Build() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		al := AccessLog{
			Method: c.Request.Method,
			Path:   c.Request.URL.Path,
		}
		if b.allowReqBody && c.Request.Body != nil {
			// 直接忽略 error，不影响程序运行
			reqBodyBytes, _ := c.GetRawData()
			// Request.Body 是一个 Stream（流）对象，所以是只能读取一次的
			// 因此读完之后要放回去，不然后续步骤是读不到的
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
			al.ReqBody = string(reqBodyBytes)
		}

		if b.allowRespBody {
			// 面向切片编程
			c.Writer = responseWriter{
				ResponseWriter: c.Writer,
				al:             &al,
			}
		}

		defer func() {
			duration := time.Since(start)
			al.Duration = duration.String()
			b.logFunc(c, al)
		}()
		c.Next()
	}
}

// AccessLog 可以打印很多的信息，根据需要自己加
type AccessLog struct {
	Method     string `json:"method"`
	Path       string `json:"path"`
	ReqBody    string `json:"req_body"`
	Duration   string `json:"duration"`
	StatusCode int    `json:"status_code"`
	RespBody   string `json:"resp_body"`
}

type responseWriter struct {
	al *AccessLog
	gin.ResponseWriter
}

func (r responseWriter) WriteHeader(statusCode int) {
	r.al.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r responseWriter) Write(data []byte) (int, error) {
	r.al.RespBody = string(data)
	return r.ResponseWriter.Write(data)
}

func (r responseWriter) WriteString(data string) (int, error) {
	r.al.RespBody = data
	return r.ResponseWriter.WriteString(data)
}
