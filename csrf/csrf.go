package csrf

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	csrfToken = "csrf_token"
)

// NewCsrfMiddlewareOption
/*
 * 初始化CSRF中间件，并设置错误处理函数。
 * @param secret 密钥
 * @param errorFunc 错误处理函数
 */
func NewCsrfMiddlewareOption(secret string, errorFunc func(ctx *gin.Context)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		recordToken, ok := session.Get(csrfToken).(string)
		// 判断是否存在
		// 如果不存在，则有人在搞你
		if !ok || len(recordToken) == 0 {
			errorFunc(ctx)
			return
		}
		// 判断是否正确
		// 不正确，则有人篡改了token
		reqToken := extractToken(ctx)
		if reqToken != recordToken {
			errorFunc(ctx)
			return
		}
		ctx.Next()
	}
}

// GetToken 获取 CSRF token.
/*
 * 登录时候调用，初始化CSRF Token并将token返回给前端。
 * 前端可以将其存储在cookie中，也可以将其存储在隐藏的表单字段中，或者放在url 参数中。
 */
func GetToken(c *gin.Context) (string, error) {
	session := sessions.Default(c)
	// 已经设置过token，直接返回
	if t, ok := c.Get(csrfToken); ok {
		return t.(string), nil
	}

	key := uuid.New().String()
	session.Set(csrfToken, key)
	err := session.Save()
	if err != nil {
		return "", err
	}
	return key, nil
}

// extractToken
// 从请求中获取token
func extractToken(ctx *gin.Context) string {
	r := ctx.Request
	if t := r.FormValue("csrf"); len(t) > 0 {
		return t
	} else if t := r.URL.Query().Get("csrf"); len(t) > 0 {
		return t
	} else if t := r.Header.Get("X-CSRF-TOKEN"); len(t) > 0 {
		return t
	}
	return ""
}
