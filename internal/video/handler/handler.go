package handler

import (
	"context"
	"shortvideo/internal/video/service"
	"shortvideo/kitex_gen/common"
	video "shortvideo/kitex_gen/video"
)

// VideoServiceImpl implements the last service interface defined in the IDL.
type VideoServiceImpl struct {
	videoService service.VideoService
}

func NewVideoService(videoService service.VideoService) *VideoServiceImpl {
	return &VideoServiceImpl{
		videoService: videoService,
	}
}

// PublishVideo implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) PublishVideo(ctx context.Context, req *video.PublishVideoReq) (resp *video.PublishVideoResp, err error) {
	successMsg := "成功"
	resp = &video.PublishVideoResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
	}

	videoID, err := s.videoService.PublishVideo(ctx, req.UserId, req.Title, req.VideoUrl, req.CoverUrl, req.Description)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.VideoId = videoID
	return resp, nil
}

// GetUserVideoList implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) GetUserVideoList(ctx context.Context, req *video.UserVideoListReq) (resp *video.UserVideoListResp, err error) {
	successMsg := "成功"
	resp = &video.UserVideoListResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Videos:     []*common.Video{},
		TotalCount: 0,
	}

	videos, total, err := s.videoService.GetUserVideos(ctx, req.UserId, req.CurrentUserId, int(req.Page), int(req.PageSize))
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	commonVideos := make([]*common.Video, len(videos))
	for i, v := range videos {
		commonVideos[i] = &common.Video{
			Id:           v.ID,
			AuthorId:     v.AuthorID,
			Url:          v.URL,
			CoverUrl:     v.CoverURL,
			Title:        v.Title,
			Description:  v.Description,
			LikeCount:    v.LikeCount,
			CommentCount: v.CommentCount,
			PublishTime:  v.PublishTime,
		}
	}

	resp.Videos = commonVideos
	resp.TotalCount = int32(total)
	return resp, nil
}

// GetFeed implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) GetFeed(ctx context.Context, req *video.FeedReq) (resp *video.FeedResp, err error) {
	successMsg := "成功"
	resp = &video.FeedResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Videos:   []*common.Video{},
		NextTime: 0,
	}

	videos, nextTime, err := s.videoService.GetFeedVideos(ctx, req.UserId, req.LatestTime, int(req.PageSize))
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	commonVideos := make([]*common.Video, len(videos))
	for i, v := range videos {
		commonVideos[i] = &common.Video{
			Id:           v.ID,
			AuthorId:     v.AuthorID,
			Url:          v.URL,
			CoverUrl:     v.CoverURL,
			Title:        v.Title,
			Description:  v.Description,
			LikeCount:    v.LikeCount,
			CommentCount: v.CommentCount,
			PublishTime:  v.PublishTime,
		}
	}

	resp.Videos = commonVideos
	resp.NextTime = nextTime
	return resp, nil
}

// SearchVideo implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) SearchVideo(ctx context.Context, req *video.SearchVideoReq) (resp *video.SearchVideoResp, err error) {
	successMsg := "成功"
	resp = &video.SearchVideoResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Videos:     []*common.Video{},
		TotalCount: 0,
	}

	videos, total, err := s.videoService.SearchVideos(ctx, req.Keyword, req.CurrentUserId, int(req.Page), int(req.PageSize))
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	commonVideos := make([]*common.Video, len(videos))
	for i, v := range videos {
		commonVideos[i] = &common.Video{
			Id:           v.ID,
			AuthorId:     v.AuthorID,
			Url:          v.URL,
			CoverUrl:     v.CoverURL,
			Title:        v.Title,
			Description:  v.Description,
			LikeCount:    v.LikeCount,
			CommentCount: v.CommentCount,
			PublishTime:  v.PublishTime,
		}
	}

	resp.Videos = commonVideos
	resp.TotalCount = int32(total)
	return resp, nil
}

// GetVideoDetail implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) GetVideoDetail(ctx context.Context, req *video.VideoDetailReq) (resp *video.VideoDetailResp, err error) {
	successMsg := "成功"
	resp = &video.VideoDetailResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
	}

	v, err := s.videoService.GetVideoByID(ctx, req.VideoId, req.CurrentUserId)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.Video = &common.Video{
		Id:           v.ID,
		AuthorId:     v.AuthorID,
		Url:          v.URL,
		CoverUrl:     v.CoverURL,
		Title:        v.Title,
		Description:  v.Description,
		LikeCount:    v.LikeCount,
		CommentCount: v.CommentCount,
		PublishTime:  v.PublishTime,
	}

	return resp, nil
}

// BatchGetVideoInfo implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) BatchGetVideoInfo(ctx context.Context, req *video.BatchVideoInfoReq) (resp *video.BatchVideoInfoResp, err error) {
	successMsg := "成功"
	resp = &video.BatchVideoInfoResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Videos: make(map[int64]*common.Video),
	}

	videos, err := s.videoService.BatchGetVideosByIDs(ctx, req.VideoIds, req.CurrentUserId)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	for id, v := range videos {
		resp.Videos[id] = &common.Video{
			Id:           v.ID,
			AuthorId:     v.AuthorID,
			Url:          v.URL,
			CoverUrl:     v.CoverURL,
			Title:        v.Title,
			Description:  v.Description,
			LikeCount:    v.LikeCount,
			CommentCount: v.CommentCount,
			PublishTime:  v.PublishTime,
		}
	}

	return resp, nil
}

// DeleteVideo implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) DeleteVideo(ctx context.Context, req *video.DeleteVideoReq) (resp *video.DeleteVideoResp, err error) {
	successMsg := "成功"
	resp = &video.DeleteVideoResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
	}

	err = s.videoService.DeleteVideo(ctx, req.VideoId, req.UserId)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	return resp, nil
}

// UpdateVideoInfo implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) UpdateVideoInfo(ctx context.Context, req *video.UpdateVideoInfoReq) (resp *video.UpdateVideoInfoResp, err error) {
	successMsg := "成功"
	resp = &video.UpdateVideoInfoResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
	}

	title := ""
	if req.Title != nil {
		title = *req.Title
	}

	description := ""
	if req.Description != nil {
		description = *req.Description
	}

	err = s.videoService.UpdateVideo(ctx, req.VideoId, req.UserId, title, description)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	return resp, nil
}

// GetVideoStats implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) GetVideoStats(ctx context.Context, req *video.VideoStatsReq) (resp *video.VideoStatsResp, err error) {
	successMsg := "成功"
	resp = &video.VideoStatsResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
	}

	stats, err := s.videoService.GetVideoStats(ctx, req.VideoId)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.Stats = &video.VideoStats{
		VideoId:      stats.VideoID,
		ViewCount:    stats.ViewCount,
		LikeCount:    stats.LikeCount,
		CommentCount: stats.CommentCount,
		ShareCount:   stats.ShareCount,
	}

	return resp, nil
}

// GetHotVideos implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) GetHotVideos(ctx context.Context, req *video.HotVideoReq) (resp *video.HotVideoResp, err error) {
	successMsg := "成功"
	resp = &video.HotVideoResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Videos: []*common.Video{},
	}

	videos, err := s.videoService.GetHotVideos(ctx, req.UserId, int(req.PageSize))
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	commonVideos := make([]*common.Video, len(videos))
	for i, v := range videos {
		commonVideos[i] = &common.Video{
			Id:           v.ID,
			AuthorId:     v.AuthorID,
			Url:          v.URL,
			CoverUrl:     v.CoverURL,
			Title:        v.Title,
			Description:  v.Description,
			LikeCount:    v.LikeCount,
			CommentCount: v.CommentCount,
			PublishTime:  v.PublishTime,
		}
	}

	resp.Videos = commonVideos
	return resp, nil
}

// UploadVideo implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) UploadVideo(ctx context.Context, req *video.UploadVideoReq) (resp *video.UploadVideoResp, err error) {
	successMsg := "成功"
	resp = &video.UploadVideoResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
	}

	videoURL, coverURL, err := s.videoService.UploadVideo(ctx, req.UserId, req.VideoData, req.CoverData, req.Title, req.Description)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.VideoUrl = videoURL
	resp.CoverUrl = coverURL
	return resp, nil
}

// GetUserVideoCount implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) GetUserVideoCount(ctx context.Context, req *video.GetUserVideoCountReq) (resp *video.GetUserVideoCountResp, err error) {
	successMsg := "成功"
	resp = &video.GetUserVideoCountResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Count: 0,
	}

	count, err := s.videoService.CountVideosByUserID(ctx, req.UserId)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.Count = count
	return resp, nil
}

// GetTotalVideoCount implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) GetTotalVideoCount(ctx context.Context, req *video.GetTotalVideoCountReq) (resp *video.GetTotalVideoCountResp, err error) {
	successMsg := "成功"
	resp = &video.GetTotalVideoCountResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Count: 0,
	}

	count, err := s.videoService.GetTotalVideoCount(ctx)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.Count = count
	return resp, nil
}
