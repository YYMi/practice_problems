package middleware

import (
	"net/http"
	"practice_problems/global" // 确保路径正确
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JwtSecret 密钥
var JwtSecret = []byte("YOUR_SUPER_SECRET_KEY_CHANGE_ME")

// MyClaims 自定义载荷
type MyClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	UserCode string `json:"user_code"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 Token
func GenerateToken(userID int, username string, userCode string) (string, error) {
	claims := MyClaims{
		UserID:   userID,
		Username: username,
		UserCode: userCode,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 30 * time.Hour)), // 30天有效期
			Issuer:    "practice_system",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JwtSecret)
}

// JWTAuthMiddleware 鉴权中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求路径和方法，用于日志
		requestPath := c.Request.URL.Path
		requestMethod := c.Request.Method

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			global.GetLog(c).Warnf("鉴权失败(无Token): %s %s", requestMethod, requestPath)
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "请求未携带 Token"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			global.GetLog(c).Warnf("鉴权失败(格式错误): %s %s", requestMethod, requestPath)
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Token 格式错误"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// ==========================================
		// 🔥 核心逻辑：检查内存白名单
		// ==========================================
		exists, storedUserCode := global.VerifyToken(tokenString)
		if !exists {
			// 记录哪个接口被拒绝了
			global.GetLog(c).Warnf("鉴权失败(失效/已登出): %s %s", requestMethod, requestPath)
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "登录已失效，请重新登录"})
			c.Abort()
			return
		}

		// 解析 Token
		token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
			return JwtSecret, nil
		})

		if err != nil || !token.Valid {
			global.GetLog(c).Warnf("鉴权失败(解析错误): %s %s - %v", requestMethod, requestPath, err)
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Token 解析失败"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*MyClaims); ok {
			c.Set("userID", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("userCode", storedUserCode)

			// ★★★ 记录访问日志 (Debug级别，防止生产环境刷屏) ★★★
			// 如果你想在生产环境看，可以改成 global.GetLog(c).Infof
			global.GetLog(c).Debugf("[%s] %s %s", storedUserCode, requestMethod, requestPath)
		}

		c.Next()
	}
}

// ParseToken 解析 Token 并返回 Claims
func ParseToken(tokenString string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return JwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
