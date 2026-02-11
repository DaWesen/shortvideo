package handler

import (
	"context"
	"shortvideo/internal/recommend/service"
	"shortvideo/kitex_gen/common"
	recommend "shortvideo/kitex_gen/recommend"
	"shortvideo/pkg/logger"
)

// RecommendServiceImpl implements the last service interface defined in the IDL.
type RecommendServiceImpl struct {
	recommendService service.RecommendService
}

func NewRecommendService(recommendService service.RecommendService) *RecommendServiceImpl {
	return &RecommendServiceImpl{
		recommendService: recommendService,
	}
}

// GetRecommendVideos implements the RecommendServiceImpl interface.
func (s *RecommendServiceImpl) GetRecommendVideos(ctx context.Context, req *recommend.GetRecommendVideosReq) (resp *recommend.GetRecommendVideosResp, err error) {
	logger.Info("GetRecommendVideos request",
		logger.Int64Field("user_id", req.UserId),
		logger.AnyField("page_size", req.PageSize),
		logger.AnyField("offset", req.Offset))

	successMsg := "成功"
	resp = &recommend.GetRecommendVideosResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Videos:     []*common.Video{},
		NextOffset: 0,
	}

	videos, nextOffset, err := s.recommendService.GetRecommendVideos(ctx, req.UserId, req.PageSize, req.Offset)
	if err != nil {
		logger.Error("GetRecommendVideos failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.Videos = videos
	resp.NextOffset = nextOffset
	logger.Info("GetRecommendVideos success",
		logger.Int64Field("user_id", req.UserId),
		logger.IntField("video_count", len(videos)),
		logger.Int64Field("next_offset", nextOffset))
	return resp, nil
}

// GetRecommendUsers implements the RecommendServiceImpl interface.
func (s *RecommendServiceImpl) GetRecommendUsers(ctx context.Context, req *recommend.GetRecommendUsersReq) (resp *recommend.GetRecommendUsersResp, err error) {
	logger.Info("GetRecommendUsers request",
		logger.Int64Field("user_id", req.UserId),
		logger.AnyField("count", req.Count))

	successMsg := "成功"
	resp = &recommend.GetRecommendUsersResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Users: []*common.User{},
	}

	users, err := s.recommendService.GetRecommendUsers(ctx, req.UserId, req.Count)
	if err != nil {
		logger.Error("GetRecommendUsers failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.Users = users
	logger.Info("GetRecommendUsers success",
		logger.Int64Field("user_id", req.UserId),
		logger.IntField("user_count", len(users)))
	return resp, nil
}

// RecordUserAction implements the RecommendServiceImpl interface.
func (s *RecommendServiceImpl) RecordUserAction(ctx context.Context, req *recommend.UserActionReq) (resp *recommend.UserActionResp, err error) {
	logger.Info("RecordUserAction request",
		logger.Int64Field("user_id", req.UserId),
		logger.Int64Field("item_id", req.ItemId),
		logger.StringField("item_type", req.ItemType),
		logger.StringField("action_type", req.ActionType),
		logger.StringField("timestamp", req.Timestamp))

	successMsg := "成功"
	resp = &recommend.UserActionResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
	}

	err = s.recommendService.RecordUserAction(ctx, req.UserId, req.ItemId, req.ItemType, req.ActionType, req.Timestamp, req.Duration, req.Score)
	if err != nil {
		logger.Error("RecordUserAction failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	logger.Info("RecordUserAction success",
		logger.Int64Field("user_id", req.UserId),
		logger.Int64Field("item_id", req.ItemId))
	return resp, nil
}

// GetHotTags implements the RecommendServiceImpl interface.
func (s *RecommendServiceImpl) GetHotTags(ctx context.Context, req *recommend.GetHotTagsReq) (resp *recommend.GetHotTagsResp, err error) {
	logger.Info("GetHotTags request", logger.AnyField("count", req.Count))

	successMsg := "成功"
	resp = &recommend.GetHotTagsResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Tags: []*recommend.TagInfo{},
	}

	tags, err := s.recommendService.GetHotTags(ctx, req.Count)
	if err != nil {
		logger.Error("GetHotTags failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.Tags = tags
	logger.Info("GetHotTags success", logger.IntField("tag_count", len(tags)))
	return resp, nil
}

// GetTagVideos implements the RecommendServiceImpl interface.
func (s *RecommendServiceImpl) GetTagVideos(ctx context.Context, req *recommend.GetTagVideosReq) (resp *recommend.GetTagVideosResp, err error) {
	logger.Info("GetTagVideos request",
		logger.StringField("tag", req.Tag),
		logger.Int64Field("user_id", req.UserId),
		logger.AnyField("page_size", req.PageSize))

	successMsg := "成功"
	resp = &recommend.GetTagVideosResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Videos: []*common.Video{},
	}

	videos, err := s.recommendService.GetTagVideos(ctx, req.Tag, req.UserId, req.PageSize)
	if err != nil {
		logger.Error("GetTagVideos failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.Videos = videos
	logger.Info("GetTagVideos success",
		logger.StringField("tag", req.Tag),
		logger.IntField("video_count", len(videos)))
	return resp, nil
}

// GetPersonalizedFeed implements the RecommendServiceImpl interface.
func (s *RecommendServiceImpl) GetPersonalizedFeed(ctx context.Context, req *recommend.GetPersonalizedFeedReq) (resp *recommend.GetPersonalizedFeedResp, err error) {
	logger.Info("GetPersonalizedFeed request",
		logger.Int64Field("user_id", req.UserId),
		logger.AnyField("page_size", req.PageSize),
		logger.AnyField("last_video_id", req.LastVideoId))

	successMsg := "成功"
	resp = &recommend.GetPersonalizedFeedResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Videos:          []*common.Video{},
		NextLastVideoId: 0,
	}

	videos, nextLastVideoId, err := s.recommendService.GetPersonalizedFeed(ctx, req.UserId, req.PageSize, req.LastVideoId)
	if err != nil {
		logger.Error("GetPersonalizedFeed failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.Videos = videos
	resp.NextLastVideoId = nextLastVideoId
	logger.Info("GetPersonalizedFeed success",
		logger.Int64Field("user_id", req.UserId),
		logger.IntField("video_count", len(videos)),
		logger.Int64Field("next_last_video_id", nextLastVideoId))
	return resp, nil
}
