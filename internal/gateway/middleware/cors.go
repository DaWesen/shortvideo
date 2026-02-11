package middleware

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

// CORS中间件
type CORSMiddleware struct{}

// 创建CORS中间件
func NewCORSMiddleware() *CORSMiddleware {
	return &CORSMiddleware{}
}

// 处理CORS
func (m *CORSMiddleware) Handle() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		//设置CORS头
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		ctx.Header("Access-Control-Max-Age", "86400")

		//处理OPTIONS请求
		if string(ctx.Method()) == "OPTIONS" {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next(c)
	}
}
