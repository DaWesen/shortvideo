package handler

import (
	"context"
	video "shortvideo/kitex_gen/video"
)

// VideoServiceImpl implements the last service interface defined in the IDL.
type VideoServiceImpl struct{}

func NewVideoService() *VideoServiceImpl {
	return &VideoServiceImpl{}
}

// PublishVideo implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) PublishVideo(ctx context.Context, req *video.PublishVideoReq) (resp *video.PublishVideoResp, err error) {
	// TODO: Your code here...
	return
}

// GetUserVideoList implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) GetUserVideoList(ctx context.Context, req *video.UserVideoListReq) (resp *video.UserVideoListResp, err error) {
	// TODO: Your code here...
	return
}

// GetFeed implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) GetFeed(ctx context.Context, req *video.FeedReq) (resp *video.FeedResp, err error) {
	// TODO: Your code here...
	return
}

// SearchVideo implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) SearchVideo(ctx context.Context, req *video.SearchVideoReq) (resp *video.SearchVideoResp, err error) {
	// TODO: Your code here...
	return
}

// GetVideoDetail implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) GetVideoDetail(ctx context.Context, req *video.VideoDetailReq) (resp *video.VideoDetailResp, err error) {
	// TODO: Your code here...
	return
}

// BatchGetVideoInfo implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) BatchGetVideoInfo(ctx context.Context, req *video.BatchVideoInfoReq) (resp *video.BatchVideoInfoResp, err error) {
	// TODO: Your code here...
	return
}

// DeleteVideo implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) DeleteVideo(ctx context.Context, req *video.DeleteVideoReq) (resp *video.DeleteVideoResp, err error) {
	// TODO: Your code here...
	return
}

// UpdateVideoInfo implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) UpdateVideoInfo(ctx context.Context, req *video.UpdateVideoInfoReq) (resp *video.UpdateVideoInfoResp, err error) {
	// TODO: Your code here...
	return
}

// GetVideoStats implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) GetVideoStats(ctx context.Context, req *video.VideoStatsReq) (resp *video.VideoStatsResp, err error) {
	// TODO: Your code here...
	return
}

// GetHotVideos implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) GetHotVideos(ctx context.Context, req *video.HotVideoReq) (resp *video.HotVideoResp, err error) {
	// TODO: Your code here...
	return
}
