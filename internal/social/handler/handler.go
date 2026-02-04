package handler

import (
	"context"
	social "shortvideo/kitex_gen/social"
)

// SocialServiceImpl implements the last service interface defined in the IDL.
type SocialServiceImpl struct{}

func NewSocialService() *SocialServiceImpl {
	return &SocialServiceImpl{}
}

// FollowAction implements the SocialServiceImpl interface.
func (s *SocialServiceImpl) FollowAction(ctx context.Context, req *social.FollowActionReq) (resp *social.FollowActionResp, err error) {
	// TODO: Your code here...
	return
}

// GetFollowList implements the SocialServiceImpl interface.
func (s *SocialServiceImpl) GetFollowList(ctx context.Context, req *social.FollowListReq) (resp *social.FollowListResp, err error) {
	// TODO: Your code here...
	return
}

// GetFollowerList implements the SocialServiceImpl interface.
func (s *SocialServiceImpl) GetFollowerList(ctx context.Context, req *social.FollowerListReq) (resp *social.FollowerListResp, err error) {
	// TODO: Your code here...
	return
}

// GetFriendList implements the SocialServiceImpl interface.
func (s *SocialServiceImpl) GetFriendList(ctx context.Context, req *social.FriendListReq) (resp *social.FriendListResp, err error) {
	// TODO: Your code here...
	return
}

// CheckFollow implements the SocialServiceImpl interface.
func (s *SocialServiceImpl) CheckFollow(ctx context.Context, req *social.CheckFollowReq) (resp *social.CheckFollowResp, err error) {
	// TODO: Your code here...
	return
}

// CheckMutualFollow implements the SocialServiceImpl interface.
func (s *SocialServiceImpl) CheckMutualFollow(ctx context.Context, req *social.CheckMutualFollowReq) (resp *social.CheckMutualFollowResp, err error) {
	// TODO: Your code here...
	return
}

// GetFollowStats implements the SocialServiceImpl interface.
func (s *SocialServiceImpl) GetFollowStats(ctx context.Context, req *social.FollowStatsReq) (resp *social.FollowStatsResp, err error) {
	// TODO: Your code here...
	return
}

// BatchCheckFollow implements the SocialServiceImpl interface.
func (s *SocialServiceImpl) BatchCheckFollow(ctx context.Context, req *social.BatchCheckFollowReq) (resp *social.BatchCheckFollowResp, err error) {
	// TODO: Your code here...
	return
}
