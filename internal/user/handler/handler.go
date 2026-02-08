package handler

import (
	"context"
	"shortvideo/internal/user/service"
	"shortvideo/kitex_gen/common"
	user "shortvideo/kitex_gen/user"
)

// UserServiceImpl implements the last service interface defined in the IDL.
type UserServiceImpl struct {
	userService service.UserService
}

func NewUserService(userService service.UserService) *UserServiceImpl {
	return &UserServiceImpl{userService: userService}
}

// Register implements the UserServiceImpl interface.
func (s *UserServiceImpl) Register(ctx context.Context, req *user.RegisterReq) (resp *user.LoginRegisterResp, err error) {
	resp = &user.LoginRegisterResp{}

	avatar := ""
	if req.Avatar != nil {
		avatar = *req.Avatar
	}
	about := ""
	if req.About != nil {
		about = *req.About
	}

	user, token, err := s.userService.Register(ctx, req.Username, req.Password, avatar, about)
	if err != nil {
		errMsg := err.Error()
		resp.BaseResp = &common.BaseResp{
			StatusCode: -1,
			Msg:        &errMsg,
		}
		return resp, nil
	}

	var avatarPtr *string
	if user.Avatar != "" {
		avatarPtr = &user.Avatar
	}
	var aboutPtr *string
	if user.About != "" {
		aboutPtr = &user.About
	}

	resp.User = &common.User{
		Id:            user.ID,
		Username:      user.Username,
		Avatar:        avatarPtr,
		About:         aboutPtr,
		FollowCount:   user.FollowCount,
		FollowerCount: user.FollowerCount,
	}
	resp.Token = token
	successMsg := "注册成功"
	resp.BaseResp = &common.BaseResp{
		StatusCode: 0,
		Msg:        &successMsg,
	}
	return resp, nil
}

// Login implements the UserServiceImpl interface.
func (s *UserServiceImpl) Login(ctx context.Context, req *user.LoginReq) (resp *user.LoginRegisterResp, err error) {
	resp = &user.LoginRegisterResp{}

	user, token, err := s.userService.Login(ctx, req.Username, req.Password)
	if err != nil {
		errMsg := err.Error()
		resp.BaseResp = &common.BaseResp{
			StatusCode: -1,
			Msg:        &errMsg,
		}
		return resp, nil
	}

	var avatarPtr *string
	if user.Avatar != "" {
		avatarPtr = &user.Avatar
	}
	var aboutPtr *string
	if user.About != "" {
		aboutPtr = &user.About
	}

	resp.User = &common.User{
		Id:            user.ID,
		Username:      user.Username,
		Avatar:        avatarPtr,
		About:         aboutPtr,
		FollowCount:   user.FollowCount,
		FollowerCount: user.FollowerCount,
	}
	resp.Token = token
	successMsg := "登录成功"
	resp.BaseResp = &common.BaseResp{
		StatusCode: 0,
		Msg:        &successMsg,
	}
	return resp, nil
}

// GetUserInfo implements the UserServiceImpl interface.
func (s *UserServiceImpl) GetUserInfo(ctx context.Context, req *user.UserInfoReq) (resp *user.UserInfoResp, err error) {
	resp = &user.UserInfoResp{}

	user, err := s.userService.GetUserByID(ctx, req.UserId)
	if err != nil {
		errMsg := err.Error()
		resp.BaseResp = &common.BaseResp{
			StatusCode: -1,
			Msg:        &errMsg,
		}
		return resp, nil
	}

	var avatarPtr *string
	if user.Avatar != "" {
		avatarPtr = &user.Avatar
	}
	var aboutPtr *string
	if user.About != "" {
		aboutPtr = &user.About
	}

	resp.User = &common.User{
		Id:            user.ID,
		Username:      user.Username,
		Avatar:        avatarPtr,
		About:         aboutPtr,
		FollowCount:   user.FollowCount,
		FollowerCount: user.FollowerCount,
	}
	successMsg := "获取用户信息成功"
	resp.BaseResp = &common.BaseResp{
		StatusCode: 0,
		Msg:        &successMsg,
	}
	return resp, nil
}

// BatchGetUserInfo implements the UserServiceImpl interface.
func (s *UserServiceImpl) BatchGetUserInfo(ctx context.Context, req *user.BatchUserInfoReq) (resp *user.BatchUserInfoResp, err error) {
	resp = &user.BatchUserInfoResp{}

	users, err := s.userService.BatchGetUsersByIDs(ctx, req.UserIds)
	if err != nil {
		errMsg := err.Error()
		resp.BaseResp = &common.BaseResp{
			StatusCode: -1,
			Msg:        &errMsg,
		}
		return resp, nil
	}

	userMap := make(map[int64]*common.User)
	for id, user := range users {
		var avatarPtr *string
		if user.Avatar != "" {
			avatarPtr = &user.Avatar
		}
		var aboutPtr *string
		if user.About != "" {
			aboutPtr = &user.About
		}

		userMap[id] = &common.User{
			Id:            user.ID,
			Username:      user.Username,
			Avatar:        avatarPtr,
			About:         aboutPtr,
			FollowCount:   user.FollowCount,
			FollowerCount: user.FollowerCount,
		}
	}
	resp.Users = userMap
	successMsg := "批量获取用户信息成功"
	resp.BaseResp = &common.BaseResp{
		StatusCode: 0,
		Msg:        &successMsg,
	}
	return resp, nil
}

// UpdateUser implements the UserServiceImpl interface.
func (s *UserServiceImpl) UpdateUser(ctx context.Context, req *user.UpdateUserReq) (resp *common.BaseResp, err error) {
	resp = &common.BaseResp{}

	avatar := ""
	if req.Avatar != nil {
		avatar = *req.Avatar
	}
	about := ""
	if req.About != nil {
		about = *req.About
	}
	oldPassword := ""
	if req.OldPassword != nil {
		oldPassword = *req.OldPassword
	}
	newPassword := ""
	if req.NewPassword_ != nil {
		newPassword = *req.NewPassword_
	}

	err = s.userService.UpdateUser(ctx, req.UserId, avatar, about, oldPassword, newPassword)
	if err != nil {
		errMsg := err.Error()
		resp.StatusCode = -1
		resp.Msg = &errMsg
		return resp, nil
	}

	successMsg := "更新用户信息成功"
	resp.StatusCode = 0
	resp.Msg = &successMsg
	return resp, nil
}

// CheckUsername implements the UserServiceImpl interface.
func (s *UserServiceImpl) CheckUsername(ctx context.Context, req *user.CheckUsernameReq) (resp *user.CheckUsernameResp, err error) {
	resp = &user.CheckUsernameResp{}

	available, err := s.userService.CheckUsernameAvailable(ctx, req.Username)
	if err != nil {
		errMsg := err.Error()
		resp.BaseResp = &common.BaseResp{
			StatusCode: -1,
			Msg:        &errMsg,
		}
		return resp, nil
	}

	resp.Available = available
	successMsg := "检查用户名成功"
	resp.BaseResp = &common.BaseResp{
		StatusCode: 0,
		Msg:        &successMsg,
	}
	return resp, nil
}

// GetUserStats implements the UserServiceImpl interface.
func (s *UserServiceImpl) GetUserStats(ctx context.Context, req *user.UserStatsReq) (resp *user.UserStatsResp, err error) {
	resp = &user.UserStatsResp{}

	count, err := s.userService.GetUserCount(ctx)
	if err != nil {
		errMsg := err.Error()
		resp.BaseResp = &common.BaseResp{
			StatusCode: -1,
			Msg:        &errMsg,
		}
		return resp, nil
	}

	resp.Stats = &user.UserStats{
		UserId:             req.UserId,
		VideoCount:         0,
		TotalLikesReceived: 0,
		TotalComments:      0,
	}
	resp.TotalUserCount = count
	successMsg := "获取用户统计信息成功"
	resp.BaseResp = &common.BaseResp{
		StatusCode: 0,
		Msg:        &successMsg,
	}
	return resp, nil
}

// VerifyToken implements the UserServiceImpl interface.
func (s *UserServiceImpl) VerifyToken(ctx context.Context, token string) (resp bool, err error) {
	_, err = s.userService.VerifyToken(ctx, token)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetUserInfoByUsername implements the UserServiceImpl interface.
func (s *UserServiceImpl) GetUserInfoByUsername(ctx context.Context, req *user.UserInfoByUsernameReq) (resp *user.UserInfoResp, err error) {
	resp = &user.UserInfoResp{}

	user, err := s.userService.GetUserByUsername(ctx, req.Username)
	if err != nil {
		errMsg := err.Error()
		resp.BaseResp = &common.BaseResp{
			StatusCode: -1,
			Msg:        &errMsg,
		}
		return resp, nil
	}

	var avatarPtr *string
	if user.Avatar != "" {
		avatarPtr = &user.Avatar
	}
	var aboutPtr *string
	if user.About != "" {
		aboutPtr = &user.About
	}

	resp.User = &common.User{
		Id:            user.ID,
		Username:      user.Username,
		Avatar:        avatarPtr,
		About:         aboutPtr,
		FollowCount:   user.FollowCount,
		FollowerCount: user.FollowerCount,
	}
	successMsg := "获取用户信息成功"
	resp.BaseResp = &common.BaseResp{
		StatusCode: 0,
		Msg:        &successMsg,
	}
	return resp, nil
}

// UpdateAvatar implements the UserServiceImpl interface.
func (s *UserServiceImpl) UpdateAvatar(ctx context.Context, req *user.UpdateAvatarReq) (resp *common.BaseResp, err error) {
	resp = &common.BaseResp{}

	_, err = s.userService.UpdateAvatar(ctx, req.UserId, req.AvatarData)
	if err != nil {
		errMsg := err.Error()
		resp.StatusCode = -1
		resp.Msg = &errMsg
		return resp, nil
	}

	successMsg := "更新头像成功"
	resp.StatusCode = 0
	resp.Msg = &successMsg
	return resp, nil
}

// BatchCheckUsernames implements the UserServiceImpl interface.
func (s *UserServiceImpl) BatchCheckUsernames(ctx context.Context, req *user.BatchCheckUsernamesReq) (resp *user.BatchCheckUsernamesResp, err error) {
	resp = &user.BatchCheckUsernamesResp{}

	availableMap, err := s.userService.BatchCheckUsernames(ctx, req.Usernames)
	if err != nil {
		errMsg := err.Error()
		resp.BaseResp = &common.BaseResp{
			StatusCode: -1,
			Msg:        &errMsg,
		}
		return resp, nil
	}

	resp.AvailableMap = availableMap
	successMsg := "批量检查用户名成功"
	resp.BaseResp = &common.BaseResp{
		StatusCode: 0,
		Msg:        &successMsg,
	}
	return resp, nil
}

// SearchUsers implements the UserServiceImpl interface.
func (s *UserServiceImpl) SearchUsers(ctx context.Context, req *user.SearchUsersReq) (resp *user.SearchUsersResp, err error) {
	resp = &user.SearchUsersResp{}

	users, total, err := s.userService.SearchUsers(ctx, req.Keyword, int(req.Page), int(req.PageSize))
	if err != nil {
		errMsg := err.Error()
		resp.BaseResp = &common.BaseResp{
			StatusCode: -1,
			Msg:        &errMsg,
		}
		return resp, nil
	}

	userList := make([]*common.User, 0, len(users))
	for _, user := range users {
		var avatarPtr *string
		if user.Avatar != "" {
			avatarPtr = &user.Avatar
		}
		var aboutPtr *string
		if user.About != "" {
			aboutPtr = &user.About
		}

		userList = append(userList, &common.User{
			Id:            user.ID,
			Username:      user.Username,
			Avatar:        avatarPtr,
			About:         aboutPtr,
			FollowCount:   user.FollowCount,
			FollowerCount: user.FollowerCount,
		})
	}

	resp.Users = userList
	resp.Total = total
	successMsg := "搜索用户成功"
	resp.BaseResp = &common.BaseResp{
		StatusCode: 0,
		Msg:        &successMsg,
	}
	return resp, nil
}

// UpdateFollowCount implements the UserServiceImpl interface.
func (s *UserServiceImpl) UpdateFollowCount(ctx context.Context, req *user.UpdateFollowCountReq) (resp *common.BaseResp, err error) {
	resp = &common.BaseResp{}

	err = s.userService.UpdateFollowCount(ctx, req.UserId, req.Delta)
	if err != nil {
		errMsg := err.Error()
		resp.StatusCode = -1
		resp.Msg = &errMsg
		return resp, nil
	}

	successMsg := "更新关注数成功"
	resp.StatusCode = 0
	resp.Msg = &successMsg
	return resp, nil
}

// UpdateFollowerCount implements the UserServiceImpl interface.
func (s *UserServiceImpl) UpdateFollowerCount(ctx context.Context, req *user.UpdateFollowerCountReq) (resp *common.BaseResp, err error) {
	resp = &common.BaseResp{}

	err = s.userService.UpdateFollowerCount(ctx, req.UserId, req.Delta)
	if err != nil {
		errMsg := err.Error()
		resp.StatusCode = -1
		resp.Msg = &errMsg
		return resp, nil
	}

	successMsg := "更新粉丝数成功"
	resp.StatusCode = 0
	resp.Msg = &successMsg
	return resp, nil
}
