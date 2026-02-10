package handler

import (
	"context"
	"shortvideo/internal/social/service"
	userService "shortvideo/internal/user/service"
	"shortvideo/kitex_gen/common"
	social "shortvideo/kitex_gen/social"
)

// SocialServiceImpl implements the last service interface defined in the IDL.
type SocialServiceImpl struct {
	socialService service.SocialService
	userService   userService.UserService
}

func NewSocialService(socialService service.SocialService, userService userService.UserService) *SocialServiceImpl {
	return &SocialServiceImpl{
		socialService: socialService,
		userService:   userService,
	}
}

// FollowAction implements the SocialServiceImpl interface.
func (s *SocialServiceImpl) FollowAction(ctx context.Context, req *social.FollowActionReq) (resp *social.FollowActionResp, err error) {
	successMsg := "成功"
	resp = &social.FollowActionResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
	}

	err = s.socialService.FollowAction(ctx, req.UserId, req.TargetUserId, req.Action)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	return resp, nil
}

// GetFollowList implements the SocialServiceImpl interface.
func (s *SocialServiceImpl) GetFollowList(ctx context.Context, req *social.FollowListReq) (resp *social.FollowListResp, err error) {
	successMsg := "成功"
	resp = &social.FollowListResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Users:      []*common.User{},
		TotalCount: 0,
	}

	targetUserIDs, total, err := s.socialService.GetFollowList(ctx, req.UserId, req.CurrentUserId, int(req.Page), int(req.PageSize))
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	users := make([]*common.User, len(targetUserIDs))

	followStatus := make(map[int64]bool)
	if req.CurrentUserId > 0 && len(targetUserIDs) > 0 {
		var err error
		followStatus, err = s.socialService.BatchCheckFollow(ctx, req.CurrentUserId, targetUserIDs)
		if err != nil {
			followStatus = make(map[int64]bool)
		}
	}

	for i, userID := range targetUserIDs {
		user, err := s.userService.GetUserByID(ctx, userID)
		if err != nil {
			users[i] = &common.User{
				Id:       userID,
				IsFollow: followStatus[userID],
			}
		} else {
			var avatar *string
			if user.Avatar != "" {
				avatar = &user.Avatar
			}
			var about *string
			if user.About != "" {
				about = &user.About
			}

			users[i] = &common.User{
				Id:            user.ID,
				Username:      user.Username,
				FollowCount:   user.FollowCount,
				FollowerCount: user.FollowerCount,
				Avatar:        avatar,
				About:         about,
				IsFollow:      followStatus[userID],
			}
		}
	}

	resp.Users = users
	resp.TotalCount = int32(total)
	return resp, nil
}

// GetFollowerList implements the SocialServiceImpl interface.
func (s *SocialServiceImpl) GetFollowerList(ctx context.Context, req *social.FollowerListReq) (resp *social.FollowerListResp, err error) {
	successMsg := "成功"
	resp = &social.FollowerListResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Users:      []*common.User{},
		TotalCount: 0,
	}

	followerIDs, total, err := s.socialService.GetFollowerList(ctx, req.UserId, req.CurrentUserId, int(req.Page), int(req.PageSize))
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	users := make([]*common.User, len(followerIDs))

	followStatus := make(map[int64]bool)
	if req.CurrentUserId > 0 && len(followerIDs) > 0 {
		var err error
		followStatus, err = s.socialService.BatchCheckFollow(ctx, req.CurrentUserId, followerIDs)
		if err != nil {
			followStatus = make(map[int64]bool)
		}
	}

	for i, userID := range followerIDs {
		user, err := s.userService.GetUserByID(ctx, userID)
		if err != nil {
			users[i] = &common.User{
				Id:       userID,
				IsFollow: followStatus[userID],
			}
		} else {
			var avatar *string
			if user.Avatar != "" {
				avatar = &user.Avatar
			}
			var about *string
			if user.About != "" {
				about = &user.About
			}

			users[i] = &common.User{
				Id:            user.ID,
				Username:      user.Username,
				FollowCount:   user.FollowCount,
				FollowerCount: user.FollowerCount,
				Avatar:        avatar,
				About:         about,
				IsFollow:      followStatus[userID],
			}
		}
	}

	resp.Users = users
	resp.TotalCount = int32(total)
	return resp, nil
}

// GetFriendList implements the SocialServiceImpl interface.
func (s *SocialServiceImpl) GetFriendList(ctx context.Context, req *social.FriendListReq) (resp *social.FriendListResp, err error) {
	successMsg := "成功"
	resp = &social.FriendListResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Users:      []*common.User{},
		TotalCount: 0,
	}

	friendIDs, err := s.socialService.GetFriendList(ctx, req.UserId, int(req.Page), int(req.PageSize))
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	users := make([]*common.User, len(friendIDs))
	for i, userID := range friendIDs {
		user, err := s.userService.GetUserByID(ctx, userID)
		if err != nil {
			users[i] = &common.User{
				Id: userID,
			}
		} else {
			var avatar *string
			if user.Avatar != "" {
				avatar = &user.Avatar
			}
			var about *string
			if user.About != "" {
				about = &user.About
			}

			users[i] = &common.User{
				Id:            user.ID,
				Username:      user.Username,
				FollowCount:   user.FollowCount,
				FollowerCount: user.FollowerCount,
				Avatar:        avatar,
				About:         about,
				IsFollow:      true,
			}
		}
	}

	resp.Users = users
	resp.TotalCount = int32(len(users))
	return resp, nil
}

// CheckFollow implements the SocialServiceImpl interface.
func (s *SocialServiceImpl) CheckFollow(ctx context.Context, req *social.CheckFollowReq) (resp *social.CheckFollowResp, err error) {
	successMsg := "成功"
	resp = &social.CheckFollowResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		IsFollowing: false,
	}

	isFollowing, err := s.socialService.CheckFollow(ctx, req.UserId, req.TargetUserId)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.IsFollowing = isFollowing
	return resp, nil
}

// CheckMutualFollow implements the SocialServiceImpl interface.
func (s *SocialServiceImpl) CheckMutualFollow(ctx context.Context, req *social.CheckMutualFollowReq) (resp *social.CheckMutualFollowResp, err error) {
	successMsg := "成功"
	resp = &social.CheckMutualFollowResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		IsMutualFollow: false,
	}

	isMutual, err := s.socialService.CheckMutualFollow(ctx, req.UserId1, req.UserId2)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.IsMutualFollow = isMutual
	return resp, nil
}

// GetFollowStats implements the SocialServiceImpl interface.
func (s *SocialServiceImpl) GetFollowStats(ctx context.Context, req *social.FollowStatsReq) (resp *social.FollowStatsResp, err error) {
	successMsg := "成功"
	resp = &social.FollowStatsResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Stats: &social.FollowStats{
			UserId:        req.UserId,
			FollowCount:   0,
			FollowerCount: 0,
			FriendCount:   0,
		},
	}

	followCount, followerCount, friendCount, err := s.socialService.GetFollowStats(ctx, req.UserId)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.Stats.FollowCount = followCount
	resp.Stats.FollowerCount = followerCount
	resp.Stats.FriendCount = friendCount
	return resp, nil
}

// BatchCheckFollow implements the SocialServiceImpl interface.
func (s *SocialServiceImpl) BatchCheckFollow(ctx context.Context, req *social.BatchCheckFollowReq) (resp *social.BatchCheckFollowResp, err error) {
	successMsg := "成功"
	resp = &social.BatchCheckFollowResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		FollowStatus: make(map[int64]bool),
	}

	followStatus, err := s.socialService.BatchCheckFollow(ctx, req.UserId, req.TargetUserIds)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.FollowStatus = followStatus
	return resp, nil
}
