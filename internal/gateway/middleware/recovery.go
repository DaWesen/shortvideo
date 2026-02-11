package middleware

import (
	"context"
	"net/http"
	"runtime/debug"

	"shortvideo/pkg/logger"

	"github.com/cloudwego/hertz/pkg/app"
)

// 恢复中间件
type RecoveryMiddleware struct{}

// 创建恢复中间件
func NewRecoveryMiddleware() *RecoveryMiddleware {
	return &RecoveryMiddleware{}
}

// 处理恢复
func (m *RecoveryMiddleware) Handle() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		defer func() {
			if err := recover(); err != nil {
				//记录错误信息
				logger.Error("HTTP请求 panic",
					logger.AnyField("error", err),
					logger.StringField("stack", string(debug.Stack())),
					logger.StringField("method", string(ctx.Method())),
					logger.StringField("path", string(ctx.Path())),
				)

				//返回500错误
				ctx.JSON(http.StatusInternalServerError, map[string]string{"message": "服务器内部错误"})
			}
		}()

		ctx.Next(c)
	}
}
