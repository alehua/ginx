package main

import (
	"github.com/alehua/ginx/csrf"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	server := gin.Default()
	store := cookie.NewStore([]byte("secret"))   // secret是加密密钥
	server.Use(sessions.Sessions("ssid", store)) // session的名字是ssid
	server.GET("/login", func(c *gin.Context) {
		token, _ := csrf.GetToken(c)
		println(token)
	})
	// 配置中间件, 中间件要在登录校验后配置，否则无法登录
	server.Use(csrf.NewCsrfMiddlewareOption("secret", func(c *gin.Context) {
		c.String(400, "CSRF token mismatch")
		c.Abort()
	}))
	server.POST("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	server.Run(":8081")
}
