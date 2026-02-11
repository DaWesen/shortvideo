package middleware

import (
	"context"
	"net/http"
	"strings"

	"shortvideo/kitex_gen/user/userservice"

	"github.com/cloudwego/hertz/pkg/app"
)

// 认证中间件
type AuthMiddleware struct {
	userClient userservice.Client
}

// 创建认证中间件
func NewAuthMiddleware(userClient userservice.Client) *AuthMiddleware {
	return &AuthMiddleware{
		userClient: userClient,
	}
}

// 处理认证
func (m *AuthMiddleware) Handle() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		authHeader := string(ctx.GetHeader("Authorization"))
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, map[string]string{"message": "Authorization头信息是必需的"})
			ctx.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			ctx.JSON(http.StatusUnauthorized, map[string]string{"message": "Authorization头信息格式必须是Bearer {token}"})
			ctx.Abort()
			return
		}

		token := parts[1]

		userID, err := m.userClient.VerifyToken(c, token)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, map[string]string{"message": "无效或过期的token"})
			ctx.Abort()
			return
		}

		newCtx := context.WithValue(c, "user_id", userID)
		ctx.Next(newCtx)
	}
}
