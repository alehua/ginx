package csrf

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	csrfToken = "csrf_token"
)

type GinCsrfMiddleware struct {
	skipConditionFunc func(ctx *gin.Context) bool // 判断是否跳过中间件逻辑
	errorFunc         func(ctx *gin.Context)      // 拦截后的请求处理
}

// NewCsrfMiddleware 创建 CSRF 中间件
func NewCsrfMiddleware() *GinCsrfMiddleware {
	return &GinCsrfMiddleware{
		skipConditionFunc: func(ctx *gin.Context) bool { return false }, // 默认都不跳过
		errorFunc: func(ctx *gin.Context) {
			ctx.AbortWithStatus(403) // 默认终止请求并返回403
		},
	}
}

func (g *GinCsrfMiddleware) SkipCondition(fn func(ctx *gin.Context) bool) *GinCsrfMiddleware {
	g.skipConditionFunc = func(ctx *gin.Context) bool {
		defer func() {
			if err := recover(); err != nil {
				// 保证方法别出错
				fmt.Println(err)
			}
		}()
		return fn(ctx)
	}
	return g
}

func (g *GinCsrfMiddleware) ErrorFunc(fn func(ctx *gin.Context)) *GinCsrfMiddleware {
	g.errorFunc = fn
	return g
}

func (g *GinCsrfMiddleware) Builder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		recordToken, ok := session.Get(csrfToken).(string)
		// 判断是否存在
		// 如果不存在，则有人在搞你
		if !ok || len(recordToken) == 0 {
			g.errorFunc(ctx)
			return
		}
		// 判断是否正确
		// 不正确，则有人篡改了token
		reqToken := g.extractToken(ctx)
		if reqToken != recordToken {
			g.errorFunc(ctx)
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
func (g *GinCsrfMiddleware) extractToken(ctx *gin.Context) string {
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
