package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/test-tt/pkg/jwt"
)

// UserIDKey context 中存储用户 ID 的 key
type userIDKey struct{}
type usernameKey struct{}

// JWTAuth JWT 认证中间件
func JWTAuth(jwtConfig *jwt.Config) app.HandlerFunc {
	j := jwt.New(jwtConfig)

	return func(ctx context.Context, c *app.RequestContext) {
		// 从 Header 获取 token
		authHeader := string(c.GetHeader("Authorization"))
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    1002,
				"message": "missing authorization header",
			})
			return
		}

		// 检查 Bearer 前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    1002,
				"message": "invalid authorization format",
			})
			return
		}

		// 解析 token
		claims, err := j.ParseToken(parts[1])
		if err != nil {
			// 安全考虑：返回泛化的错误信息，避免泄露 token 验证细节
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    1002,
				"message": "invalid or expired token",
			})
			return
		}

		// 将用户信息存入 context
		ctx = context.WithValue(ctx, userIDKey{}, claims.UserID)
		ctx = context.WithValue(ctx, usernameKey{}, claims.Username)
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		c.Next(ctx)
	}
}

// GetUserID 从 context 获取用户 ID
func GetUserID(ctx context.Context) uint64 {
	if id, ok := ctx.Value(userIDKey{}).(uint64); ok {
		return id
	}
	return 0
}

// GetUsername 从 context 获取用户名
func GetUsername(ctx context.Context) string {
	if name, ok := ctx.Value(usernameKey{}).(string); ok {
		return name
	}
	return ""
}

// GetUserIDFromContext 从 RequestContext 获取用户 ID（安全版本）
func GetUserIDFromContext(c *app.RequestContext) uint64 {
	if id, exists := c.Get("user_id"); exists {
		if uid, ok := id.(uint64); ok {
			return uid
		}
	}
	return 0
}

// GetUsernameFromContext 从 RequestContext 获取用户名（安全版本）
func GetUsernameFromContext(c *app.RequestContext) string {
	if name, exists := c.Get("username"); exists {
		if uname, ok := name.(string); ok {
			return uname
		}
	}
	return ""
}
