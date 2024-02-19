package jwt

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"strings"
	"time"
)

type Handler interface {
	SetLoginToken(ctx *gin.Context, uid int64, expireAt time.Duration) error // 设置登录token
	ExtractTokenString(ctx *gin.Context) string                              // 提取token字符串
	CheckSession(ctx *gin.Context, ssid string) error                        // 判断是否已经退出登录
	Logout(ctx *gin.Context, expireAt time.Duration) error                   // 退出登录
}

type LoginClaims struct {
	Id   int64  // 用户id
	Ssid string // 用户登录状态唯一标识符, 用于验证用户是否登录
	jwt.RegisteredClaims
}

// DefaultHandler 默认实现, 不支持logout检查
type DefaultHandler struct {
	SigningMethod  jwt.SigningMethod
	AccessTokenKey []byte
}

func NewDefaultHandler(signingMethod jwt.SigningMethod, accessTokenKey []byte) *DefaultHandler {
	return &DefaultHandler{SigningMethod: signingMethod, AccessTokenKey: accessTokenKey}
}

func (d *DefaultHandler) SetLoginToken(ctx *gin.Context, uid int64, expireAt time.Duration) error {
	token := jwt.NewWithClaims(d.SigningMethod, LoginClaims{
		Id: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireAt)), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),               // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),               // 生效时间å
		},
	})
	tokenStr, err := token.SignedString(d.AccessTokenKey)
	if err != nil {
		return err
	}
	// 返回token
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (d *DefaultHandler) ExtractTokenString(ctx *gin.Context) string {
	authCode := ctx.GetHeader("Authorization")
	if authCode == "" {
		return ""
	}
	authSegments := strings.SplitN(authCode, " ", 2)
	if len(authSegments) != 2 {
		return ""
	}
	return authSegments[1]
}

func (d *DefaultHandler) CheckSession(ctx *gin.Context, ssid string) error {
	return nil
}

func (d *DefaultHandler) Logout(ctx *gin.Context, expireAt time.Duration) error {
	return nil
}
