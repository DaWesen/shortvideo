package handler

import (
	"context"
	"net/http"
	"strconv"

	"shortvideo/kitex_gen/danmu"
	"shortvideo/kitex_gen/interaction"
	"shortvideo/kitex_gen/live"
	"shortvideo/kitex_gen/message"
	"shortvideo/kitex_gen/recommend"
	"shortvideo/kitex_gen/social"
	"shortvideo/kitex_gen/user"
	"shortvideo/kitex_gen/video"

	"github.com/cloudwego/hertz/pkg/app"
)

// 处理HTTP请求
type HTTPHandler struct {
	clients *ServiceClients
}

// 创建新的HTTP处理器
func NewHTTPHandler(clients *ServiceClients) *HTTPHandler {
	return &HTTPHandler{
		clients: clients,
	}
}

// 响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 成功响应
func (h *HTTPHandler) success(ctx *app.RequestContext, data interface{}) {
	resp := Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	}
	h.writeResponse(ctx, resp)
}

// 错误响应
func (h *HTTPHandler) error(ctx *app.RequestContext, code int, message string) {
	resp := Response{
		Code:    code,
		Message: message,
	}
	h.writeResponse(ctx, resp)
}

// 写入响应
func (h *HTTPHandler) writeResponse(ctx *app.RequestContext, resp Response) {
	ctx.JSON(resp.Code, resp)
}

// 用户注册
func (h *HTTPHandler) Register(c context.Context, ctx *app.RequestContext) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Avatar   string `json:"avatar"`
		About    string `json:"about"`
	}
	if err := ctx.Bind(&req); err != nil {
		h.error(ctx, http.StatusBadRequest, "请求体无效")
		return
	}

	if h.clients.UserClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "用户服务不可用")
		return
	}

	registerReq := &user.RegisterReq{
		Username: req.Username,
		Password: req.Password,
	}

	if req.Avatar != "" {
		registerReq.Avatar = &req.Avatar
	}
	if req.About != "" {
		registerReq.About = &req.About
	}

	resp, err := h.clients.UserClient.Register(c, registerReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, map[string]interface{}{
		"user":  resp.User,
		"token": resp.Token,
	})
}

// 用户登录
func (h *HTTPHandler) Login(c context.Context, ctx *app.RequestContext) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := ctx.Bind(&req); err != nil {
		h.error(ctx, http.StatusBadRequest, "请求体无效")
		return
	}

	if h.clients.UserClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "用户服务不可用")
		return
	}

	loginReq := &user.LoginReq{
		Username: req.Username,
		Password: req.Password,
	}

	resp, err := h.clients.UserClient.Login(c, loginReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, map[string]interface{}{
		"user":  resp.User,
		"token": resp.Token,
	})
}

// 获取视频流
func (h *HTTPHandler) GetVideoFeed(c context.Context, ctx *app.RequestContext) {
	pageSize, _ := strconv.Atoi(ctx.Query("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}

	latestTime, _ := strconv.ParseInt(ctx.Query("latest_time"), 10, 64)

	userID, _ := c.Value("user_id").(int64)

	if h.clients.VideoClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "视频服务不可用")
		return
	}

	feedReq := &video.FeedReq{
		UserId:     userID,
		LatestTime: latestTime,
		PageSize:   int32(pageSize),
	}

	resp, err := h.clients.VideoClient.GetFeed(c, feedReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, map[string]interface{}{
		"videos":    resp.Videos,
		"next_time": resp.NextTime,
	})
}

// 获取视频详情
func (h *HTTPHandler) GetVideoByID(c context.Context, ctx *app.RequestContext) {
	videoID, err := strconv.ParseInt(ctx.Query("id"), 10, 64)
	if err != nil {
		h.error(ctx, http.StatusBadRequest, "无效的视频ID")
		return
	}

	userID, _ := c.Value("user_id").(int64)

	if h.clients.VideoClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "视频服务不可用")
		return
	}

	detailReq := &video.VideoDetailReq{
		VideoId:       videoID,
		CurrentUserId: userID,
	}

	resp, err := h.clients.VideoClient.GetVideoDetail(c, detailReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, resp.Video)
}

// 搜索视频
func (h *HTTPHandler) Search(c context.Context, ctx *app.RequestContext) {
	keyword := ctx.Query("keyword")
	page, _ := strconv.Atoi(ctx.Query("page"))
	pageSize, _ := strconv.Atoi(ctx.Query("page_size"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	userID, _ := c.Value("user_id").(int64)

	if h.clients.VideoClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "视频服务不可用")
		return
	}

	searchReq := &video.SearchVideoReq{
		Keyword:       keyword,
		CurrentUserId: userID,
		Page:          int32(page),
		PageSize:      int32(pageSize),
	}

	resp, err := h.clients.VideoClient.SearchVideo(c, searchReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, map[string]interface{}{
		"videos": resp.Videos,
		"total":  resp.TotalCount,
		"page":   page,
		"size":   pageSize,
	})
}

// 获取用户信息
func (h *HTTPHandler) GetUserProfile(c context.Context, ctx *app.RequestContext) {
	userID, _ := c.Value("user_id").(int64)

	if h.clients.UserClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "用户服务不可用")
		return
	}

	infoReq := &user.UserInfoReq{
		UserId:        userID,
		CurrentUserId: userID,
	}

	resp, err := h.clients.UserClient.GetUserInfo(c, infoReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, resp.User)
}

// 更新用户信息
func (h *HTTPHandler) UpdateUser(c context.Context, ctx *app.RequestContext) {
	userID, _ := c.Value("user_id").(int64)

	var req struct {
		Avatar      string `json:"avatar"`
		About       string `json:"about"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := ctx.Bind(&req); err != nil {
		h.error(ctx, http.StatusBadRequest, "请求体无效")
		return
	}

	if h.clients.UserClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "用户服务不可用")
		return
	}

	updateReq := &user.UpdateUserReq{
		UserId: userID,
	}

	if req.Avatar != "" {
		updateReq.Avatar = &req.Avatar
	}
	if req.About != "" {
		updateReq.About = &req.About
	}
	if req.OldPassword != "" {
		updateReq.OldPassword = &req.OldPassword
	}
	if req.NewPassword != "" {
		updateReq.NewPassword_ = &req.NewPassword
	}

	_, err := h.clients.UserClient.UpdateUser(c, updateReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, nil)
}

// 关注用户
func (h *HTTPHandler) FollowUser(c context.Context, ctx *app.RequestContext) {
	userID, _ := c.Value("user_id").(int64)

	var req struct {
		TargetUserId int64 `json:"target_user_id"`
	}
	if err := ctx.Bind(&req); err != nil {
		h.error(ctx, http.StatusBadRequest, "请求体无效")
		return
	}

	if h.clients.SocialClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "社交服务不可用")
		return
	}

	followReq := &social.FollowActionReq{
		UserId:       userID,
		TargetUserId: req.TargetUserId,
		Action:       true,
	}

	_, err := h.clients.SocialClient.FollowAction(c, followReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, nil)
}

// 取消关注
func (h *HTTPHandler) UnfollowUser(c context.Context, ctx *app.RequestContext) {
	userID, _ := c.Value("user_id").(int64)

	var req struct {
		TargetUserId int64 `json:"target_user_id"`
	}
	if err := ctx.Bind(&req); err != nil {
		h.error(ctx, http.StatusBadRequest, "请求体无效")
		return
	}

	if h.clients.SocialClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "社交服务不可用")
		return
	}

	followReq := &social.FollowActionReq{
		UserId:       userID,
		TargetUserId: req.TargetUserId,
		Action:       false,
	}

	_, err := h.clients.SocialClient.FollowAction(c, followReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, nil)
}

// 获取关注列表
func (h *HTTPHandler) GetFollowingList(c context.Context, ctx *app.RequestContext) {
	userID, _ := c.Value("user_id").(int64)
	page, _ := strconv.Atoi(ctx.Query("page"))
	pageSize, _ := strconv.Atoi(ctx.Query("page_size"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	if h.clients.SocialClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "社交服务不可用")
		return
	}

	followingReq := &social.FollowListReq{
		UserId:   userID,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	resp, err := h.clients.SocialClient.GetFollowList(c, followingReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, resp.Users)
}

// 获取粉丝列表
func (h *HTTPHandler) GetFollowerList(c context.Context, ctx *app.RequestContext) {
	userID, _ := c.Value("user_id").(int64)
	page, _ := strconv.Atoi(ctx.Query("page"))
	pageSize, _ := strconv.Atoi(ctx.Query("page_size"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	if h.clients.SocialClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "社交服务不可用")
		return
	}

	followerReq := &social.FollowerListReq{
		UserId:   userID,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	resp, err := h.clients.SocialClient.GetFollowerList(c, followerReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, resp.Users)
}

// 点赞视频
func (h *HTTPHandler) LikeVideo(c context.Context, ctx *app.RequestContext) {
	userID, _ := c.Value("user_id").(int64)

	var req struct {
		VideoId int64 `json:"video_id"`
	}
	if err := ctx.Bind(&req); err != nil {
		h.error(ctx, http.StatusBadRequest, "请求体无效")
		return
	}

	if h.clients.InteractionClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "交互服务不可用")
		return
	}

	likeReq := &interaction.LikeActionReq{
		UserId:  userID,
		VideoId: req.VideoId,
		Action:  true,
	}

	_, err := h.clients.InteractionClient.LikeAction(c, likeReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, nil)
}

// 取消点赞
func (h *HTTPHandler) UnlikeVideo(c context.Context, ctx *app.RequestContext) {
	userID, _ := c.Value("user_id").(int64)

	var req struct {
		VideoId int64 `json:"video_id"`
	}
	if err := ctx.Bind(&req); err != nil {
		h.error(ctx, http.StatusBadRequest, "请求体无效")
		return
	}

	if h.clients.InteractionClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "交互服务不可用")
		return
	}

	likeReq := &interaction.LikeActionReq{
		UserId:  userID,
		VideoId: req.VideoId,
		Action:  false,
	}

	_, err := h.clients.InteractionClient.LikeAction(c, likeReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, nil)
}

// 评论视频
func (h *HTTPHandler) CommentVideo(c context.Context, ctx *app.RequestContext) {
	userID, _ := c.Value("user_id").(int64)

	var req struct {
		VideoId int64  `json:"video_id"`
		Content string `json:"content"`
	}
	if err := ctx.Bind(&req); err != nil {
		h.error(ctx, http.StatusBadRequest, "请求体无效")
		return
	}

	if h.clients.InteractionClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "交互服务不可用")
		return
	}

	commentReq := &interaction.CommentActionReq{
		UserId:  userID,
		VideoId: req.VideoId,
		Content: req.Content,
	}

	resp, err := h.clients.InteractionClient.CommentAction(c, commentReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, resp.Comment)
}

// 获取评论列表
func (h *HTTPHandler) GetComments(c context.Context, ctx *app.RequestContext) {
	videoID, err := strconv.ParseInt(ctx.Query("video_id"), 10, 64)
	if err != nil {
		h.error(ctx, http.StatusBadRequest, "无效的视频ID")
		return
	}

	page, _ := strconv.Atoi(ctx.Query("page"))
	pageSize, _ := strconv.Atoi(ctx.Query("page_size"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	if h.clients.InteractionClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "交互服务不可用")
		return
	}

	commentsReq := &interaction.CommentListReq{
		VideoId:  videoID,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	resp, err := h.clients.InteractionClient.GetCommentList(c, commentsReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, resp.Comments)
}

// 发送消息
func (h *HTTPHandler) SendMessage(c context.Context, ctx *app.RequestContext) {
	userID, _ := c.Value("user_id").(int64)

	var req struct {
		ReceiverId int64  `json:"receiver_id"`
		Content    string `json:"content"`
	}
	if err := ctx.Bind(&req); err != nil {
		h.error(ctx, http.StatusBadRequest, "请求体无效")
		return
	}

	if h.clients.MessageClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "消息服务不可用")
		return
	}

	sendReq := &message.SendMessageReq{
		SenderId:   userID,
		ReceiverId: req.ReceiverId,
		Content:    req.Content,
	}

	resp, err := h.clients.MessageClient.SendMessage(c, sendReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, map[string]interface{}{
		"message_id": resp.MessageId,
	})
}

// 获取消息列表
func (h *HTTPHandler) GetMessageList(c context.Context, ctx *app.RequestContext) {
	userID, _ := c.Value("user_id").(int64)
	otherUserID, _ := strconv.ParseInt(ctx.Query("other_user_id"), 10, 64)

	if h.clients.MessageClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "消息服务不可用")
		return
	}

	messageReq := &message.GetChatHistoryReq{
		UserId1: userID,
		UserId2: otherUserID,
	}

	resp, err := h.clients.MessageClient.GetChatHistory(c, messageReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, resp.Messages)
}

// 开始直播
func (h *HTTPHandler) StartLive(c context.Context, ctx *app.RequestContext) {
	userID, _ := c.Value("user_id").(int64)

	var req struct {
		RoomId  int64  `json:"room_id"`
		RtmpUrl string `json:"rtmp_url"`
	}
	if err := ctx.Bind(&req); err != nil {
		h.error(ctx, http.StatusBadRequest, "请求体无效")
		return
	}

	if h.clients.LiveClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "直播服务不可用")
		return
	}

	startReq := &live.StartLiveReq{
		HostId:  userID,
		RoomId:  req.RoomId,
		RtmpUrl: req.RtmpUrl,
	}

	_, err := h.clients.LiveClient.StartLive(c, startReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, nil)
}

// 停止直播
func (h *HTTPHandler) StopLive(c context.Context, ctx *app.RequestContext) {
	userID, _ := c.Value("user_id").(int64)

	var req struct {
		RoomId int64 `json:"room_id"`
	}
	if err := ctx.Bind(&req); err != nil {
		h.error(ctx, http.StatusBadRequest, "请求体无效")
		return
	}

	if h.clients.LiveClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "直播服务不可用")
		return
	}

	stopReq := &live.StopLiveReq{
		HostId: userID,
		RoomId: req.RoomId,
	}

	_, err := h.clients.LiveClient.StopLive(c, stopReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, nil)
}

// 获取直播列表
func (h *HTTPHandler) GetLiveList(c context.Context, ctx *app.RequestContext) {
	userID, _ := c.Value("user_id").(int64)
	page, _ := strconv.Atoi(ctx.Query("page"))
	pageSize, _ := strconv.Atoi(ctx.Query("page_size"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	if h.clients.LiveClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "直播服务不可用")
		return
	}

	liveReq := &live.GetLiveRoomsReq{
		UserId:   userID,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	resp, err := h.clients.LiveClient.GetLiveRooms(c, liveReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, resp.Rooms)
}

// 发送弹幕
func (h *HTTPHandler) SendDanmu(c context.Context, ctx *app.RequestContext) {
	userID, _ := c.Value("user_id").(int64)

	var req struct {
		LiveId   int64  `json:"live_id"`
		Content  string `json:"content"`
		Color    string `json:"color"`
		Position int32  `json:"position"`
	}
	if err := ctx.Bind(&req); err != nil {
		h.error(ctx, http.StatusBadRequest, "请求体无效")
		return
	}

	if h.clients.DanmuClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "弹幕服务不可用")
		return
	}

	danmuReq := &danmu.SendDanmuReq{
		UserId:  userID,
		LiveId:  req.LiveId,
		Content: req.Content,
	}

	if req.Color != "" {
		danmuReq.Color = &req.Color
	}
	if req.Position > 0 {
		danmuReq.Position = &req.Position
	}

	resp, err := h.clients.DanmuClient.SendDanmu(c, danmuReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, map[string]interface{}{
		"danmu_id": resp.DanmuId,
	})
}

// 获取弹幕列表
func (h *HTTPHandler) GetDanmuList(c context.Context, ctx *app.RequestContext) {
	liveID, err := strconv.ParseInt(ctx.Query("live_id"), 10, 64)
	if err != nil {
		h.error(ctx, http.StatusBadRequest, "无效的直播ID")
		return
	}

	if h.clients.DanmuClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "弹幕服务不可用")
		return
	}

	danmuReq := &danmu.GetDanmuHistoryReq{
		LiveId: liveID,
	}

	resp, err := h.clients.DanmuClient.GetDanmuHistory(c, danmuReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, resp.Danmus)
}

// 获取推荐视频
func (h *HTTPHandler) GetRecommendedVideos(c context.Context, ctx *app.RequestContext) {
	userID, _ := c.Value("user_id").(int64)
	pageSize, _ := strconv.Atoi(ctx.Query("page_size"))

	if pageSize <= 0 {
		pageSize = 10
	}

	if h.clients.RecommendClient == nil {
		h.error(ctx, http.StatusServiceUnavailable, "推荐服务不可用")
		return
	}

	recommendReq := &recommend.GetRecommendVideosReq{
		UserId:   userID,
		PageSize: int32(pageSize),
	}

	resp, err := h.clients.RecommendClient.GetRecommendVideos(c, recommendReq)
	if err != nil {
		h.error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(ctx, resp.Videos)
}
