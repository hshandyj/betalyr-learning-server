package middleware

import (
	"betalyr-learning-server/internal/pkg/logger"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 自定义Gin上下文键
const (
	UserIDKey   = "user_id"
	AuthTypeKey = "auth_type"
)

// AuthType 表示身份验证类型
type AuthType string

const (
	AuthTypeVirtual AuthType = "virtual"
	AuthTypeJWT     AuthType = "jwt"
)

// GetUserID 从Gin上下文中获取用户ID
// 如果用户ID不存在，返回空字符串和false
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return "", false
	}
	return userID.(string), true
}

// GetAuthType 从Gin上下文中获取身份验证类型
// 如果身份验证类型不存在，返回空字符串和false
func GetAuthType(c *gin.Context) (AuthType, bool) {
	authType, exists := c.Get(AuthTypeKey)
	if !exists {
		return "", false
	}
	return authType.(AuthType), true
}

// 解析JWT令牌获取用户ID（简化版，实际应该使用JWT库解析）
func parseJWTToken(token string) string {
	// 这里仅为示例，实际应用中应该使用JWT库正确解析令牌
	// 假设格式为 "Bearer xxxxx.yyyyy.zzzzz"
	parts := strings.Split(token, ".")
	if len(parts) != 3 || !strings.HasPrefix(token, "Bearer ") {
		return ""
	}

	// 实际应用中，这里应该解码JWT的payload部分并提取用户ID
	// 这里简化处理，返回一个假的ID
	return "jwt-user-id"
}

// AuthChecker 是一个中间件，用于检查请求头中是否包含X-Virtual-User-ID或Authorization字段
// 如果含有任一字段，提取用户ID并存储到上下文中
// 如果两者都没有，则返回401 Unauthorized
func AuthChecker() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求头中的认证信息
		virtualUserId := c.GetHeader("X-Virtual-User-ID")
		authorization := c.GetHeader("Authorization")

		// 检查是否有任意一种认证方式
		if virtualUserId == "" && authorization == "" {
			logger.Warn("请求缺少身份验证",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("ip", c.ClientIP()),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "缺少身份验证，请提供X-Virtual-User-ID或Authorization",
			})
			c.Abort() // 终止请求处理
			return
		}

		// 优先使用JWT令牌（如果两种认证方式都存在）
		var userId string
		var authType AuthType

		if authorization != "" {
			// 解析JWT令牌获取用户ID
			userId = parseJWTToken(authorization)
			if userId != "" {
				authType = AuthTypeJWT
				logger.Info("使用JWT令牌认证",
					zap.String("user_id", userId),
					zap.String("path", c.Request.URL.Path),
				)
			} else {
				// JWT解析失败，尝试使用虚拟用户ID
				if virtualUserId != "" {
					userId = virtualUserId
					authType = AuthTypeVirtual
					logger.Info("JWT令牌无效，使用虚拟用户ID认证",
						zap.String("virtual_user_id", virtualUserId),
						zap.String("path", c.Request.URL.Path),
					)
				} else {
					// 两种认证方式都无效
					c.JSON(http.StatusUnauthorized, gin.H{
						"error": "提供的认证信息无效",
					})
					c.Abort()
					return
				}
			}
		} else if virtualUserId != "" {
			// 只有虚拟用户ID
			userId = virtualUserId
			authType = AuthTypeVirtual
			logger.Info("使用虚拟用户ID认证",
				zap.String("virtual_user_id", virtualUserId),
				zap.String("path", c.Request.URL.Path),
			)
		}

		// 将用户ID和认证类型存储到上下文中，供后续处理函数使用
		c.Set(UserIDKey, userId)
		c.Set(AuthTypeKey, authType)

		// 继续处理请求
		c.Next()
	}
}
