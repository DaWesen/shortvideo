package router

import (
	"context"
	"net/http"

	"shortvideo/internal/gateway/handler"
	"shortvideo/internal/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// 注册路由
func RegisterRoutes(
	srv *server.Hertz,
	httpHandler *handler.HTTPHandler,
	authMiddleware *middleware.AuthMiddleware,
	corsMiddleware *middleware.CORSMiddleware,
	loggerMiddleware *middleware.LoggerMiddleware,
	recoveryMiddleware *middleware.RecoveryMiddleware,
) {
	//应用全局中间件
	srv.Use(
		corsMiddleware.Handle(),
		loggerMiddleware.Handle(),
		recoveryMiddleware.Handle(),
	)

	//健康检查
	srv.GET("/health", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(http.StatusOK, map[string]string{"status": "OK"})
	})

	//API路由组
	api := srv.Group("/api")

	//公开路由
	public := api.Group("/")
	{
		//用户相关
		public.POST("/user/register", httpHandler.Register)
		public.POST("/user/login", httpHandler.Login)

		//视频相关
		public.GET("/video/feed", httpHandler.GetVideoFeed)
		public.GET("/video/detail", httpHandler.GetVideoByID)
		public.GET("/search", httpHandler.Search)

		//交互相关
		public.GET("/interaction/comments", httpHandler.GetComments)

		//弹幕相关
		public.GET("/danmu/list", httpHandler.GetDanmuList)

		//直播相关
		public.GET("/live/list", httpHandler.GetLiveList)
	}

	//需要认证的路由
	protected := api.Group("/auth")
	protected.Use(authMiddleware.Handle())
	{
		//用户相关
		protected.GET("/user/profile", httpHandler.GetUserProfile)
		protected.PUT("/user/update", httpHandler.UpdateUser)

		//社交相关
		protected.POST("/social/follow", httpHandler.FollowUser)
		protected.POST("/social/unfollow", httpHandler.UnfollowUser)
		protected.GET("/social/following", httpHandler.GetFollowingList)
		protected.GET("/social/follower", httpHandler.GetFollowerList)

		//交互相关
		protected.POST("/interaction/like", httpHandler.LikeVideo)
		protected.POST("/interaction/unlike", httpHandler.UnlikeVideo)
		protected.POST("/interaction/comment", httpHandler.CommentVideo)

		//消息相关
		protected.POST("/message/send", httpHandler.SendMessage)
		protected.GET("/message/list", httpHandler.GetMessageList)

		//直播相关
		protected.POST("/live/start", httpHandler.StartLive)
		protected.POST("/live/stop", httpHandler.StopLive)

		//弹幕相关
		protected.POST("/danmu/send", httpHandler.SendDanmu)

		//推荐相关
		protected.GET("/recommend/videos", httpHandler.GetRecommendedVideos)
	}

	hlog.Info("路由注册完成")
}
