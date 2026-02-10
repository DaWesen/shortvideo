package handler

import (
	"context"
	"shortvideo/internal/live/service"
	"shortvideo/kitex_gen/common"
	live "shortvideo/kitex_gen/live"
	"shortvideo/pkg/logger"
)

// LiveServiceImpl implements the last service interface defined in the IDL.
type LiveServiceImpl struct {
	liveService service.LiveService
}

func NewLiveService(liveService service.LiveService) *LiveServiceImpl {
	return &LiveServiceImpl{
		liveService: liveService,
	}
}

// CreateLiveRoom implements the LiveServiceImpl interface.
func (s *LiveServiceImpl) CreateLiveRoom(ctx context.Context, req *live.CreateLiveRoomReq) (resp *live.CreateLiveRoomResp, err error) {
	logger.Info("CreateLiveRoom request",
		logger.Int64Field("host_id", req.HostId),
		logger.StringField("title", req.Title))

	successMsg := "成功"
	resp = &live.CreateLiveRoomResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Room: nil,
	}

	room, err := s.liveService.CreateLiveRoom(ctx, req.HostId, req.Title, req.CoverUrl)
	if err != nil {
		logger.Error("CreateLiveRoom failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	commonRoom := service.ConvertToCommonLiveRoom(room)
	resp.Room = commonRoom

	logger.Info("CreateLiveRoom success", logger.Int64Field("room_id", room.ID))
	return resp, nil
}

// StartLive implements the LiveServiceImpl interface.
func (s *LiveServiceImpl) StartLive(ctx context.Context, req *live.StartLiveReq) (resp *live.StartLiveResp, err error) {
	logger.Info("StartLive request",
		logger.Int64Field("host_id", req.HostId),
		logger.Int64Field("room_id", req.RoomId))

	successMsg := "成功"
	resp = &live.StartLiveResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
	}

	err = s.liveService.StartLive(ctx, req.HostId, req.RoomId, req.RtmpUrl)
	if err != nil {
		logger.Error("StartLive failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	logger.Info("StartLive success", logger.Int64Field("room_id", req.RoomId))
	return resp, nil
}

// StopLive implements the LiveServiceImpl interface.
func (s *LiveServiceImpl) StopLive(ctx context.Context, req *live.StopLiveReq) (resp *live.StopLiveResp, err error) {
	logger.Info("StopLive request",
		logger.Int64Field("host_id", req.HostId),
		logger.Int64Field("room_id", req.RoomId))

	successMsg := "成功"
	resp = &live.StopLiveResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
	}

	err = s.liveService.StopLive(ctx, req.HostId, req.RoomId)
	if err != nil {
		logger.Error("StopLive failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	logger.Info("StopLive success", logger.Int64Field("room_id", req.RoomId))
	return resp, nil
}

// GetLiveRooms implements the LiveServiceImpl interface.
func (s *LiveServiceImpl) GetLiveRooms(ctx context.Context, req *live.GetLiveRoomsReq) (resp *live.GetLiveRoomsResp, err error) {
	logger.Info("GetLiveRooms request",
		logger.Int64Field("user_id", req.UserId),
		logger.IntField("page", int(req.Page)),
		logger.IntField("page_size", int(req.PageSize)))

	successMsg := "成功"
	resp = &live.GetLiveRoomsResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Rooms:      []*common.LiveRoom{},
		TotalCount: 0,
	}

	rooms, total, err := s.liveService.GetLiveRooms(ctx, req.UserId, int(req.Page), int(req.PageSize), req.GetFollowingOnly())
	if err != nil {
		logger.Error("GetLiveRooms failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	commonRooms := make([]*common.LiveRoom, len(rooms))
	for i, room := range rooms {
		commonRooms[i] = service.ConvertToCommonLiveRoom(room)
	}

	resp.Rooms = commonRooms
	resp.TotalCount = int32(total)

	logger.Info("GetLiveRooms success", logger.IntField("room_count", len(rooms)), logger.Int64Field("total", total))
	return resp, nil
}

// GetLiveRoomDetail implements the LiveServiceImpl interface.
func (s *LiveServiceImpl) GetLiveRoomDetail(ctx context.Context, req *live.GetLiveRoomDetailReq) (resp *live.GetLiveRoomDetailResp, err error) {
	logger.Info("GetLiveRoomDetail request",
		logger.Int64Field("room_id", req.RoomId),
		logger.Int64Field("user_id", req.UserId))

	successMsg := "成功"
	resp = &live.GetLiveRoomDetailResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Room:        nil,
		OnlineCount: 0,
	}

	room, onlineCount, err := s.liveService.GetLiveRoomDetail(ctx, req.RoomId, req.UserId)
	if err != nil {
		logger.Error("GetLiveRoomDetail failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	commonRoom := service.ConvertToCommonLiveRoom(room)
	resp.Room = commonRoom
	resp.OnlineCount = onlineCount

	logger.Info("GetLiveRoomDetail success", logger.Int64Field("room_id", room.ID))
	return resp, nil
}

// JoinLiveRoom implements the LiveServiceImpl interface.
func (s *LiveServiceImpl) JoinLiveRoom(ctx context.Context, req *live.JoinLiveRoomReq) (resp *live.JoinLiveRoomResp, err error) {
	logger.Info("JoinLiveRoom request",
		logger.Int64Field("room_id", req.RoomId),
		logger.Int64Field("user_id", req.UserId))

	successMsg := "成功"
	resp = &live.JoinLiveRoomResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		HlsUrl:      "",
		ChatHistory: []string{},
	}

	hlsURL, chatHistory, err := s.liveService.JoinLiveRoom(ctx, req.RoomId, req.UserId)
	if err != nil {
		logger.Error("JoinLiveRoom failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.HlsUrl = hlsURL
	resp.ChatHistory = chatHistory

	logger.Info("JoinLiveRoom success", logger.Int64Field("room_id", req.RoomId), logger.Int64Field("user_id", req.UserId))
	return resp, nil
}

// LeaveLiveRoom implements the LiveServiceImpl interface.
func (s *LiveServiceImpl) LeaveLiveRoom(ctx context.Context, req *live.LeaveLiveRoomReq) (resp *live.LeaveLiveRoomResp, err error) {
	logger.Info("LeaveLiveRoom request",
		logger.Int64Field("room_id", req.RoomId),
		logger.Int64Field("user_id", req.UserId))

	successMsg := "成功"
	resp = &live.LeaveLiveRoomResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
	}

	err = s.liveService.LeaveLiveRoom(ctx, req.RoomId, req.UserId)
	if err != nil {
		logger.Error("LeaveLiveRoom failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	logger.Info("LeaveLiveRoom success", logger.Int64Field("room_id", req.RoomId), logger.Int64Field("user_id", req.UserId))
	return resp, nil
}

// SendGift implements the LiveServiceImpl interface.
func (s *LiveServiceImpl) SendGift(ctx context.Context, req *live.SendGiftReq) (resp *live.SendGiftResp, err error) {
	logger.Info("SendGift request",
		logger.Int64Field("sender_id", req.SenderId),
		logger.Int64Field("room_id", req.RoomId),
		logger.Int64Field("gift_id", req.GiftId),
		logger.AnyField("count", req.Count))

	successMsg := "成功"
	resp = &live.SendGiftResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		TotalPrice: 0,
	}

	totalPrice, err := s.liveService.SendGift(ctx, req.SenderId, req.RoomId, req.GiftId, req.Count)
	if err != nil {
		logger.Error("SendGift failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.TotalPrice = totalPrice

	logger.Info("SendGift success", logger.Int64Field("total_price", totalPrice))
	return resp, nil
}

// GetGiftList implements the LiveServiceImpl interface.
func (s *LiveServiceImpl) GetGiftList(ctx context.Context) (resp *live.GetGiftListResp, err error) {
	logger.Info("GetGiftList request")

	successMsg := "成功"
	resp = &live.GetGiftListResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Gifts: []*live.Gift{},
	}

	gifts, err := s.liveService.GetGiftList(ctx)
	if err != nil {
		logger.Error("GetGiftList failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	liveGifts := make([]*live.Gift, len(gifts))
	for i, gift := range gifts {
		liveGifts[i] = service.ConvertToLiveGift(gift)
	}

	resp.Gifts = liveGifts

	logger.Info("GetGiftList success", logger.IntField("gift_count", len(gifts)))
	return resp, nil
}

// SetRoomAdmin implements the LiveServiceImpl interface.
func (s *LiveServiceImpl) SetRoomAdmin(ctx context.Context, req *live.SetRoomAdminReq) (resp *live.SetRoomAdminResp, err error) {
	logger.Info("SetRoomAdmin request",
		logger.Int64Field("host_id", req.HostId),
		logger.Int64Field("room_id", req.RoomId),
		logger.Int64Field("target_user_id", req.TargetUserId),
		logger.BoolField("action", req.Action))

	successMsg := "成功"
	resp = &live.SetRoomAdminResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
	}

	err = s.liveService.SetRoomAdmin(ctx, req.HostId, req.RoomId, req.TargetUserId, req.Action)
	if err != nil {
		logger.Error("SetRoomAdmin failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	logger.Info("SetRoomAdmin success", logger.Int64Field("room_id", req.RoomId), logger.Int64Field("target_user_id", req.TargetUserId))
	return resp, nil
}

// RecordLive implements the LiveServiceImpl interface.
func (s *LiveServiceImpl) RecordLive(ctx context.Context, req *live.RecordLiveReq) (resp *live.RecordLiveResp, err error) {
	logger.Info("RecordLive request",
		logger.Int64Field("host_id", req.HostId),
		logger.Int64Field("room_id", req.RoomId),
		logger.BoolField("action", req.Action))

	successMsg := "成功"
	resp = &live.RecordLiveResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		VideoUrl: nil,
	}

	videoURL, err := s.liveService.RecordLive(ctx, req.HostId, req.RoomId, req.Action)
	if err != nil {
		logger.Error("RecordLive failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.VideoUrl = &videoURL

	logger.Info("RecordLive success", logger.Int64Field("room_id", req.RoomId), logger.StringField("video_url", videoURL))
	return resp, nil
}
