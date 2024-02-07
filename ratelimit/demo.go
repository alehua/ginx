package ratelimit

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"time"
)

func Demo() {
	cmd := redis.NewClient(&redis.Options{
		Addr: "redis.addr",
	})
	// 创建限流器 1000/s
	limiter := NewBuilder(cmd, time.Second, 1000).Prefix("demo").Build()
	server := gin.Default()
	server.Use(limiter)
}
