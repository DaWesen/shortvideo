package middleware

import (
	"context"
	"time"

	"shortvideo/pkg/logger"

	"github.com/cloudwego/hertz/pkg/app"
)

// 日志中间件
type LoggerMiddleware struct{}

// 创建日志中间件
func NewLoggerMiddleware() *LoggerMiddleware {
	return &LoggerMiddleware{}
}

// 处理日志
func (m *LoggerMiddleware) Handle() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		//记录请求开始时间
		start := time.Now()

		//记录请求信息
		logger.Info("HTTP Request",
			logger.StringField("method", string(ctx.Method())),
			logger.StringField("path", string(ctx.Path())),
			logger.StringField("query", ctx.QueryArgs().String()),
			logger.StringField("ip", ctx.ClientIP()),
			logger.StringField("user_agent", string(ctx.Request.Header.UserAgent())),
		)

		//处理请求
		ctx.Next(c)

		//计算请求处理时间
		duration := time.Since(start)

		//记录响应信息
		logger.Info("HTTP Response",
			logger.StringField("method", string(ctx.Method())),
			logger.StringField("path", string(ctx.Path())),
			logger.IntField("status", ctx.Response.StatusCode()),
			logger.DurationField("duration", duration),
		)
	}
}
