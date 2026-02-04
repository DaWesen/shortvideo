package handler

import (
	"context"
	recommend "shortvideo/kitex_gen/recommend"
)

// RecommendServiceImpl implements the last service interface defined in the IDL.
type RecommendServiceImpl struct{}

func NewRecommendService() *RecommendServiceImpl {
	return &RecommendServiceImpl{}
}

// GetRecommendVideos implements the RecommendServiceImpl interface.
func (s *RecommendServiceImpl) GetRecommendVideos(ctx context.Context, req *recommend.GetRecommendVideosReq) (resp *recommend.GetRecommendVideosResp, err error) {
	// TODO: Your code here...
	return
}

// GetRecommendUsers implements the RecommendServiceImpl interface.
func (s *RecommendServiceImpl) GetRecommendUsers(ctx context.Context, req *recommend.GetRecommendUsersReq) (resp *recommend.GetRecommendUsersResp, err error) {
	// TODO: Your code here...
	return
}

// RecordUserAction implements the RecommendServiceImpl interface.
func (s *RecommendServiceImpl) RecordUserAction(ctx context.Context, req *recommend.UserActionReq) (resp *recommend.UserActionResp, err error) {
	// TODO: Your code here...
	return
}

// GetHotTags implements the RecommendServiceImpl interface.
func (s *RecommendServiceImpl) GetHotTags(ctx context.Context, req *recommend.GetHotTagsReq) (resp *recommend.GetHotTagsResp, err error) {
	// TODO: Your code here...
	return
}

// GetTagVideos implements the RecommendServiceImpl interface.
func (s *RecommendServiceImpl) GetTagVideos(ctx context.Context, req *recommend.GetTagVideosReq) (resp *recommend.GetTagVideosResp, err error) {
	// TODO: Your code here...
	return
}

// GetPersonalizedFeed implements the RecommendServiceImpl interface.
func (s *RecommendServiceImpl) GetPersonalizedFeed(ctx context.Context, req *recommend.GetPersonalizedFeedReq) (resp *recommend.GetPersonalizedFeedResp, err error) {
	// TODO: Your code here...
	return
}
