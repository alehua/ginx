package auth

import (
	"github.com/ecodeclub/ekit/set"
	"github.com/gin-gonic/gin"
)

type PathAccess struct {
	publicPaths set.Set[string]
}

func NewURLPathAccess(urls []string) gin.HandlerFunc {
	n := len(urls)
	if n == 0 {
		return func(c *gin.Context) {
		}
	}
	s := set.NewMapSet[string](n)
	for i := 0; i < n; i++ {
		s.Add(urls[i])
	}
	return func(ctx *gin.Context) {
		if s.Exist(ctx.Request.URL.Path) {
			return
		}
		ctx.Next()
	}
}
