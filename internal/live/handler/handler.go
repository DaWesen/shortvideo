package handler

import (
	"context"
	live "shortvideo/kitex_gen/live"
)

// LiveServiceImpl implements the last service interface defined in the IDL.
type LiveServiceImpl struct{}

func NewLiveService() *LiveServiceImpl {
	return &LiveServiceImpl{}
}

// CreateLiveRoom implements the LiveServiceImpl interface.
func (s *LiveServiceImpl) CreateLiveRoom(ctx context.Context, req *live.CreateLiveRoomReq) (resp *live.CreateLiveRoomResp, err error) {
	// TODO: Your code here...
	return
}

// StartLive implements the LiveServiceImpl interface.
func (s *LiveServiceImpl) StartLive(ctx context.Context, req *live.StartLiveReq) (resp *live.StartLiveResp, err error) {
	// TODO: Your code here...
	return
}

// StopLive implements the LiveServiceImpl interface.
func (s *LiveServiceImpl) StopLive(ctx context.Context, req *live.StopLiveReq) (resp *live.StopLiveResp, err error) {
	// TODO: Your code here...
	return
}

// GetLiveRooms implements the LiveServiceImpl interface.
func (s *LiveServiceImpl) GetLiveRooms(ctx context.Context, req *live.GetLiveRoomsReq) (resp *live.GetLiveRoomsResp, err error) {
	// TODO: Your code here...
	return
}

// GetLiveRoomDetail implements the LiveServiceImpl interface.
func (s *LiveServiceImpl) GetLiveRoomDetail(ctx context.Context, req *live.GetLiveRoomDetailReq) (resp *live.GetLiveRoomDetailResp, err error) {
	// TODO: Your code here...
	return
}

// JoinLiveRoom implements the LiveServiceImpl interface.
func (s *LiveServiceImpl) JoinLiveRoom(ctx context.Context, req *live.JoinLiveRoomReq) (resp *live.JoinLiveRoomResp, err error) {
	// TODO: Your code here...
	return
}

// LeaveLiveRoom implements the LiveServiceImpl interface.
func (s *LiveServiceImpl) LeaveLiveRoom(ctx context.Context, req *live.LeaveLiveRoomReq) (resp *live.LeaveLiveRoomResp, err error) {
	// TODO: Your code here...
	return
}

// SendGift implements the LiveServiceImpl interface.
func (s *LiveServiceImpl) SendGift(ctx context.Context, req *live.SendGiftReq) (resp *live.SendGiftResp, err error) {
	// TODO: Your code here...
	return
}

// GetGiftList implements the LiveServiceImpl interface.
func (s *LiveServiceImpl) GetGiftList(ctx context.Context) (resp *live.GetGiftListResp, err error) {
	// TODO: Your code here...
	return
}

// SetRoomAdmin implements the LiveServiceImpl interface.
func (s *LiveServiceImpl) SetRoomAdmin(ctx context.Context, req *live.SetRoomAdminReq) (resp *live.SetRoomAdminResp, err error) {
	// TODO: Your code here...
	return
}

// RecordLive implements the LiveServiceImpl interface.
func (s *LiveServiceImpl) RecordLive(ctx context.Context, req *live.RecordLiveReq) (resp *live.RecordLiveResp, err error) {
	// TODO: Your code here...
	return
}
