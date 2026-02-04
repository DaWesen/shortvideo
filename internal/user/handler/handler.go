package handler

import (
	"context"
	common "shortvideo/kitex_gen/common"
	user "shortvideo/kitex_gen/user"
)

// UserServiceImpl implements the last service interface defined in the IDL.
type UserServiceImpl struct{}

func NewUserService() *UserServiceImpl {
	return &UserServiceImpl{}
}

// Register implements the UserServiceImpl interface.
func (s *UserServiceImpl) Register(ctx context.Context, req *user.RegisterReq) (resp *user.LoginRegisterResp, err error) {
	// TODO: Your code here...
	return
}

// Login implements the UserServiceImpl interface.
func (s *UserServiceImpl) Login(ctx context.Context, req *user.LoginReq) (resp *user.LoginRegisterResp, err error) {
	// TODO: Your code here...
	return
}

// GetUserInfo implements the UserServiceImpl interface.
func (s *UserServiceImpl) GetUserInfo(ctx context.Context, req *user.UserInfoReq) (resp *user.UserInfoResp, err error) {
	// TODO: Your code here...
	return
}

// BatchGetUserInfo implements the UserServiceImpl interface.
func (s *UserServiceImpl) BatchGetUserInfo(ctx context.Context, req *user.BatchUserInfoReq) (resp *user.BatchUserInfoResp, err error) {
	// TODO: Your code here...
	return
}

// UpdateUser implements the UserServiceImpl interface.
func (s *UserServiceImpl) UpdateUser(ctx context.Context, req *user.UpdateUserReq) (resp *common.BaseResp, err error) {
	// TODO: Your code here...
	return
}

// CheckUsername implements the UserServiceImpl interface.
func (s *UserServiceImpl) CheckUsername(ctx context.Context, req *user.CheckUsernameReq) (resp *user.CheckUsernameResp, err error) {
	// TODO: Your code here...
	return
}

// GetUserStats implements the UserServiceImpl interface.
func (s *UserServiceImpl) GetUserStats(ctx context.Context, req *user.UserStatsReq) (resp *user.UserStatsResp, err error) {
	// TODO: Your code here...
	return
}

// VerifyToken implements the UserServiceImpl interface.
func (s *UserServiceImpl) VerifyToken(ctx context.Context, token string) (resp bool, err error) {
	// TODO: Your code here...
	return
}
