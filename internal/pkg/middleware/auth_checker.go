package middleware

import (
	"betalyr-learning-server/internal/pkg/logger"
	"encoding/base64"
	"encoding/json"
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

// 解析JWT令牌获取用户ID (支持Firebase认证)
func parseJWTToken(token string) string {
	// 检查令牌格式，Firebase令牌格式为 "Bearer xxxxx.yyyyy.zzzzz"
	if !strings.HasPrefix(token, "Bearer ") {
		logger.Warn("令牌格式错误，缺少Bearer前缀")
		return ""
	}

	// 移除"Bearer "前缀
	tokenOnly := strings.TrimPrefix(token, "Bearer ")

	// 按点分割，获取三个部分
	parts := strings.Split(tokenOnly, ".")
	if len(parts) != 3 {
		logger.Warn("令牌格式错误，不是有效的JWT格式")
		return ""
	}

	// 解码payload部分（第二部分）
	payload, err := base64UrlDecode(parts[1])
	if err != nil {
		logger.Error("解码JWT payload失败", zap.Error(err))
		return ""
	}

	// 解析JSON
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		logger.Error("解析JWT payload JSON失败", zap.Error(err))
		return ""
	}

	// 从claims中提取用户ID
	// Firebase通常使用uid或sub字段作为用户ID
	if uid, ok := claims["uid"].(string); ok && uid != "" {
		logger.Info("从Firebase JWT中提取到uid", zap.String("uid", uid))
		return uid
	}

	if sub, ok := claims["sub"].(string); ok && sub != "" {
		logger.Info("从Firebase JWT中提取到sub", zap.String("sub", sub))
		return sub
	}

	// Firebase也可能在user_id字段中存储用户ID
	if userId, ok := claims["user_id"].(string); ok && userId != "" {
		logger.Info("从Firebase JWT中提取到user_id", zap.String("user_id", userId))
		return userId
	}

	// 检查Firebase特有的字段
	if identities, ok := claims["firebase"].(map[string]interface{}); ok {
		if sign_in_provider, exists := identities["sign_in_provider"].(string); exists {
			logger.Info("检测到Firebase认证", zap.String("provider", sign_in_provider))
		}
	}

	// 尝试使用email作为最后的备选
	if email, ok := claims["email"].(string); ok && email != "" {
		logger.Info("使用email作为用户ID", zap.String("email", email))
		return email
	}

	// 记录完整的claims以便调试
	claimsJson, _ := json.Marshal(claims)
	logger.Warn("Firebase JWT中没有找到有效的用户标识", zap.String("claims", string(claimsJson)))
	return ""
}

// base64URL解码
func base64UrlDecode(str string) ([]byte, error) {
	// 添加填充
	padding := 4 - (len(str) % 4)
	if padding < 4 {
		str += strings.Repeat("=", padding)
	}

	// 替换URL安全字符
	str = strings.ReplaceAll(str, "-", "+")
	str = strings.ReplaceAll(str, "_", "/")

	return base64.StdEncoding.DecodeString(str)
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
