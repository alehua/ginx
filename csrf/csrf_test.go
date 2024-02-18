package csrf

import (
	"bytes"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
)

type CsrfMiddlewareTestSuite struct {
	suite.Suite

	server *gin.Engine
	token  string
}

func (s *CsrfMiddlewareTestSuite) SetupSuite() {
	server := gin.Default()
	store := cookie.NewStore([]byte("secret"))   // secret是加密密钥
	server.Use(sessions.Sessions("ssid", store)) // session的名字是ssid
	s.server = server
	s.server.GET("/login", func(c *gin.Context) {
		token, err := GetToken(c)
		assert.NoError(s.T(), err)
		s.token = token
	})
	// 配置中间件, 中间件要在登录校验后配置，否则无法登录
	server.Use(NewCsrfMiddlewareOption("secret", func(c *gin.Context) {
		c.String(400, "CSRF token mismatch")
		c.Abort()
	}))
	s.server.POST("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
}

func (s *CsrfMiddlewareTestSuite) TearDownTest() {
}

// TestCsrfMiddleware 启动测试
func TestCsrfMiddleware(t *testing.T) {
	suite.Run(t, &CsrfMiddlewareTestSuite{})
}

func (s *CsrfMiddlewareTestSuite) request(method, url string,
	header map[string]string, body string) *http.Request {
	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(body)))
	assert.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")
	if header != nil {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}
	return req
}

func (s *CsrfMiddlewareTestSuite) TestCsrfMiddleware() {
	t := s.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		method, url string
		body        string
		header      map[string]string

		wantErr  error
		wantCode int
	}{
		{
			// TODO: 测试跑不过, 请求未携带session
			name: "正确使用token, 放到header中",
			before: func(t *testing.T) {
				req := s.request(http.MethodGet, "/login", nil, `{}`)
				recorder := httptest.NewRecorder()
				s.server.ServeHTTP(recorder, req)
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
			wantCode: http.StatusOK,
			method:   "POST",
			header:   map[string]string{"X-CSRF-Token": s.token},
			url:      "/test",
			body:     `{}`,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.before != nil {
				tc.before(t)
			}
			if tc.after != nil {
				defer tc.after(t)
			}
			req := s.request(tc.method, tc.url, tc.header, tc.body)
			// 创建一个 http 的响应体 <==> 等价于 http.ResponseWriter
			recorder := httptest.NewRecorder()
			s.server.ServeHTTP(recorder, req)

			code := recorder.Code
			// 判断响应状态码是否正确
			assert.Equal(t, tc.wantCode, code)
		})
	}
}
