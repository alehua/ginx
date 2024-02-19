package jwt

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

/*
 * redis实现, 使用redis主要用于验证logout状态.
 * 调用Logout 退出登录, 清除登录状态
 * CheckSession 判断用户使用已经登录
 */

type RedisJWTHandler struct {
	Cmd            redis.Cmdable
	SigningMethod  jwt.SigningMethod
	AccessTokenKey []byte
}

func NewRedisHandler(cmd redis.Cmdable,
	SigningMethod jwt.SigningMethod, AccessTokenKey []byte) Handler {
	return &RedisJWTHandler{
		Cmd:            cmd,
		SigningMethod:  SigningMethod,
		AccessTokenKey: AccessTokenKey,
	}
}

func (r *RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid int64, expireAt time.Duration) error {
	ssid := uuid.New().String()
	return r.setJWTToken(ctx, uid, ssid, expireAt)
}

func (r *RedisJWTHandler) setJWTToken(ctx *gin.Context,
	uid int64, ssid string, expireAt time.Duration) error {
	token := jwt.NewWithClaims(r.SigningMethod, LoginClaims{
		Id:   uid,
		Ssid: ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireAt)), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),               // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),               // 生效时间å
		},
	})
	tokenStr, err := token.SignedString(r.AccessTokenKey)
	if err != nil {
		return err
	}
	// 返回token
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

// ExtractTokenString 解析token, 获取用户信息
func (r *RedisJWTHandler) ExtractTokenString(ctx *gin.Context) string {
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

// Logout 退出登录, 清除登录状态
func (r *RedisJWTHandler) Logout(ctx *gin.Context, expireAt time.Duration) error {
	ctx.Header("x-jwt-token", "")
	// 这里不可能拿不到
	uc := ctx.MustGet("user").(LoginClaims)
	// expireAt设置很大也可以
	return r.Cmd.Set(ctx, r.key(uc.Ssid), "", expireAt).Err()
}

// CheckSession 判断用户使用已经登录
func (r *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	logout, err := r.Cmd.Exists(ctx, r.key(ssid)).Result()
	if err != nil {
		return err
	}
	if logout > 0 {
		return errors.New("用户已经退出登录")
	}
	return nil
}

func (r *RedisJWTHandler) key(ssid string) string {
	return fmt.Sprintf("users:Ssid:%s", ssid)
}
