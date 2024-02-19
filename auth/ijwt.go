package auth

import (
	ijwt "github.com/alehua/ginx/auth/jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

type JwtMiddleware struct {
	Secret  []byte
	Handler ijwt.Handler
}

func NewJwtMiddleware(secret []byte, handler ijwt.Handler) *JwtMiddleware {
	return &JwtMiddleware{
		Secret:  secret,
		Handler: handler,
		// ijwt.NewDefaultHandler(jwt.SigningMethodES256, secret),
		// ijwt.NewRedisHandler(reids.Cmd, jwt.SigningMethodES256, secret),
	}
}

func (j *JwtMiddleware) Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenStr := j.Handler.ExtractTokenString(ctx)
		uc := ijwt.LoginClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return j.Secret, nil
		})

		// 不正确的 token
		if err != nil || !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		expireTime, err := uc.GetExpirationTime()
		// 拿不到过期时间
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// 已经过期
		if expireTime.Before(time.Now()) {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		err = j.Handler.CheckSession(ctx, uc.Ssid)
		if err != nil {
			// 已经退出登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}
