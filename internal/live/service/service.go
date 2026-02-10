package service

import (
	"context"
	"errors"
	"fmt"
	"shortvideo/internal/live/dao"
	"shortvideo/internal/live/model"
	"shortvideo/kitex_gen/common"
	"shortvideo/kitex_gen/live"
	"shortvideo/pkg/logger"
	"time"
)

var (
	ErrRoomNotFound     = errors.New("直播间不存在")
	ErrNotRoomOwner     = errors.New("不是直播间房主")
	ErrRoomAlreadyLive  = errors.New("直播间已经在直播")
	ErrRoomNotLive      = errors.New("直播间不在直播")
	ErrGiftNotFound     = errors.New("礼物不存在")
	ErrInternalServer   = errors.New("服务器内部错误")
	ErrInvalidParameter = errors.New("参数错误")
	ErrUserNotFound     = errors.New("用户不存在")
)

type LiveService interface {
	// 直播间相关
	CreateLiveRoom(ctx context.Context, hostID int64, title, coverURL string) (*model.LiveRoom, error)
	StartLive(ctx context.Context, hostID, roomID int64, rtmpURL string) error
	StopLive(ctx context.Context, hostID, roomID int64) error
	GetLiveRooms(ctx context.Context, userID int64, page, pageSize int, followingOnly bool) ([]*model.LiveRoom, int64, error)
	GetLiveRoomDetail(ctx context.Context, roomID, userID int64) (*model.LiveRoom, int64, error)
	JoinLiveRoom(ctx context.Context, roomID, userID int64) (string, []string, error)
	LeaveLiveRoom(ctx context.Context, roomID, userID int64) error
	//礼物相关
	SendGift(ctx context.Context, senderID, roomID, giftID int64, count int32) (int64, error)
	GetGiftList(ctx context.Context) ([]*model.Gift, error)
	//管理员相关
	SetRoomAdmin(ctx context.Context, hostID, roomID, targetUserID int64, action bool) error
	//录制相关
	RecordLive(ctx context.Context, hostID, roomID int64, action bool) (string, error)
	//事务相关
	WithTransaction(ctx context.Context, fn func(txService LiveService) error) error
}

type liveServiceImpl struct {
	roomRepo       dao.LiveRoomRepository
	giftRepo       dao.GiftRepository
	giftRecordRepo dao.GiftRecordRepository
	roomAdminRepo  dao.RoomAdminRepository
	liveRecordRepo dao.LiveRecordRepository
	roomViewerRepo dao.RoomViewerRepository
}

func NewLiveService(
	roomRepo dao.LiveRoomRepository,
	giftRepo dao.GiftRepository,
	giftRecordRepo dao.GiftRecordRepository,
	roomAdminRepo dao.RoomAdminRepository,
	liveRecordRepo dao.LiveRecordRepository,
	roomViewerRepo dao.RoomViewerRepository,
) LiveService {
	return &liveServiceImpl{
		roomRepo:       roomRepo,
		giftRepo:       giftRepo,
		giftRecordRepo: giftRecordRepo,
		roomAdminRepo:  roomAdminRepo,
		liveRecordRepo: liveRecordRepo,
		roomViewerRepo: roomViewerRepo,
	}
}

func NewLiveServiceWithRepo(
	roomRepo dao.LiveRoomRepository,
	giftRepo dao.GiftRepository,
	giftRecordRepo dao.GiftRecordRepository,
	roomAdminRepo dao.RoomAdminRepository,
	liveRecordRepo dao.LiveRecordRepository,
	roomViewerRepo dao.RoomViewerRepository,
) LiveService {
	return &liveServiceImpl{
		roomRepo:       roomRepo,
		giftRepo:       giftRepo,
		giftRecordRepo: giftRecordRepo,
		roomAdminRepo:  roomAdminRepo,
		liveRecordRepo: liveRecordRepo,
		roomViewerRepo: roomViewerRepo,
	}
}

// 创建直播间
func (s *liveServiceImpl) CreateLiveRoom(ctx context.Context, hostID int64, title, coverURL string) (*model.LiveRoom, error) {
	logger.Info("创建直播间请求",
		logger.Int64Field("host_id", hostID),
		logger.StringField("title", title))

	if title == "" {
		return nil, ErrInvalidParameter
	}

	existingRoom, err := s.roomRepo.FindByHostID(ctx, hostID)
	if err != nil {
		logger.Error("查询直播间失败",
			logger.ErrorField(err),
			logger.Int64Field("host_id", hostID))
		return nil, ErrInternalServer
	}

	if existingRoom != nil {
		logger.Warn("用户已有直播间",
			logger.Int64Field("host_id", hostID),
			logger.Int64Field("room_id", existingRoom.ID))
		return existingRoom, nil
	}

	room := &model.LiveRoom{
		HostID:      hostID,
		Title:       title,
		CoverURL:    coverURL,
		ViewerCount: 0,
		IsLive:      false,
		CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := s.roomRepo.Create(ctx, room); err != nil {
		logger.Error("创建直播间失败",
			logger.ErrorField(err),
			logger.Int64Field("host_id", hostID),
			logger.StringField("title", title))
		return nil, ErrInternalServer
	}

	logger.Info("创建直播间成功",
		logger.Int64Field("room_id", room.ID),
		logger.Int64Field("host_id", hostID),
		logger.StringField("title", title))

	return room, nil
}

// 开始直播
func (s *liveServiceImpl) StartLive(ctx context.Context, hostID, roomID int64, rtmpURL string) error {
	logger.Info("开始直播请求",
		logger.Int64Field("host_id", hostID),
		logger.Int64Field("room_id", roomID))

	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		logger.Error("查询直播间失败",
			logger.ErrorField(err),
			logger.Int64Field("room_id", roomID))
		return ErrInternalServer
	}

	if room == nil {
		logger.Warn("直播间不存在",
			logger.Int64Field("room_id", roomID))
		return ErrRoomNotFound
	}

	if room.HostID != hostID {
		logger.Warn("不是直播间房主",
			logger.Int64Field("host_id", hostID),
			logger.Int64Field("room_id", roomID),
			logger.Int64Field("actual_host_id", room.HostID))
		return ErrNotRoomOwner
	}

	if room.IsLive {
		logger.Warn("直播间已经在直播",
			logger.Int64Field("room_id", roomID))
		return ErrRoomAlreadyLive
	}

	hlsURL := fmt.Sprintf("/live/%d/index.m3u8", roomID)

	if err := s.roomRepo.UpdateStreamURLs(ctx, roomID, rtmpURL, hlsURL); err != nil {
		logger.Error("更新流地址失败",
			logger.ErrorField(err),
			logger.Int64Field("room_id", roomID))
		return ErrInternalServer
	}

	if err := s.roomRepo.UpdateLiveStatus(ctx, roomID, true); err != nil {
		logger.Error("更新直播状态失败",
			logger.ErrorField(err),
			logger.Int64Field("room_id", roomID))
		return ErrInternalServer
	}

	logger.Info("开始直播成功",
		logger.Int64Field("room_id", roomID),
		logger.Int64Field("host_id", hostID))

	return nil
}

// 停止直播
func (s *liveServiceImpl) StopLive(ctx context.Context, hostID, roomID int64) error {
	logger.Info("停止直播请求",
		logger.Int64Field("host_id", hostID),
		logger.Int64Field("room_id", roomID))

	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		logger.Error("查询直播间失败",
			logger.ErrorField(err),
			logger.Int64Field("room_id", roomID))
		return ErrInternalServer
	}

	if room == nil {
		logger.Warn("直播间不存在",
			logger.Int64Field("room_id", roomID))
		return ErrRoomNotFound
	}

	if room.HostID != hostID {
		logger.Warn("不是直播间房主",
			logger.Int64Field("host_id", hostID),
			logger.Int64Field("room_id", roomID),
			logger.Int64Field("actual_host_id", room.HostID))
		return ErrNotRoomOwner
	}

	if !room.IsLive {
		logger.Warn("直播间不在直播",
			logger.Int64Field("room_id", roomID))
		return ErrRoomNotLive
	}

	if err := s.roomRepo.UpdateLiveStatus(ctx, roomID, false); err != nil {
		logger.Error("更新直播状态失败",
			logger.ErrorField(err),
			logger.Int64Field("room_id", roomID))
		return ErrInternalServer
	}

	if err := s.roomRepo.UpdateViewerCount(ctx, roomID, -room.ViewerCount); err != nil {
		logger.Error("重置观众数失败",
			logger.ErrorField(err),
			logger.Int64Field("room_id", roomID))
	}

	logger.Info("停止直播成功",
		logger.Int64Field("room_id", roomID),
		logger.Int64Field("host_id", hostID))

	return nil
}

// 获取直播列表
func (s *liveServiceImpl) GetLiveRooms(ctx context.Context, userID int64, page, pageSize int, followingOnly bool) ([]*model.LiveRoom, int64, error) {
	logger.Info("获取直播列表请求",
		logger.Int64Field("user_id", userID),
		logger.IntField("page", page),
		logger.IntField("page_size", pageSize),
		logger.BoolField("following_only", followingOnly))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	rooms, total, err := s.roomRepo.ListLiveRooms(ctx, page, pageSize, followingOnly, userID)
	if err != nil {
		logger.Error("获取直播列表失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID))
		return nil, 0, ErrInternalServer
	}

	logger.Info("获取直播列表成功",
		logger.Int64Field("user_id", userID),
		logger.IntField("page", page),
		logger.IntField("page_size", pageSize),
		logger.Int64Field("total", total))

	return rooms, total, nil
}

// 获取直播间详情
func (s *liveServiceImpl) GetLiveRoomDetail(ctx context.Context, roomID, userID int64) (*model.LiveRoom, int64, error) {
	logger.Info("获取直播间详情请求",
		logger.Int64Field("room_id", roomID),
		logger.Int64Field("user_id", userID))

	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		logger.Error("查询直播间失败",
			logger.ErrorField(err),
			logger.Int64Field("room_id", roomID))
		return nil, 0, ErrInternalServer
	}

	if room == nil {
		logger.Warn("直播间不存在",
			logger.Int64Field("room_id", roomID))
		return nil, 0, ErrRoomNotFound
	}

	onlineCount, err := s.roomViewerRepo.CountByRoomID(ctx, roomID)
	if err != nil {
		logger.Error("获取在线人数失败",
			logger.ErrorField(err),
			logger.Int64Field("room_id", roomID))
		onlineCount = room.ViewerCount
	}

	logger.Info("获取直播间详情成功",
		logger.Int64Field("room_id", roomID),
		logger.Int64Field("host_id", room.HostID),
		logger.Int64Field("online_count", onlineCount))

	return room, onlineCount, nil
}

// 加入直播间
func (s *liveServiceImpl) JoinLiveRoom(ctx context.Context, roomID, userID int64) (string, []string, error) {
	logger.Info("加入直播间请求",
		logger.Int64Field("room_id", roomID),
		logger.Int64Field("user_id", userID))

	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		logger.Error("查询直播间失败",
			logger.ErrorField(err),
			logger.Int64Field("room_id", roomID))
		return "", nil, ErrInternalServer
	}

	if room == nil {
		logger.Warn("直播间不存在",
			logger.Int64Field("room_id", roomID))
		return "", nil, ErrRoomNotFound
	}

	if !room.IsLive {
		logger.Warn("直播间不在直播",
			logger.Int64Field("room_id", roomID))
		return "", nil, ErrRoomNotLive
	}

	existingViewer, err := s.roomViewerRepo.Find(ctx, roomID, userID)
	if err != nil {
		logger.Error("查询观众记录失败",
			logger.ErrorField(err),
			logger.Int64Field("room_id", roomID),
			logger.Int64Field("user_id", userID))
	}

	if existingViewer == nil || existingViewer.LeaveTime != nil {
		viewer := &model.RoomViewer{
			RoomID:   roomID,
			UserID:   userID,
			JoinTime: time.Now(),
		}
		if err := s.roomViewerRepo.CreateOrUpdate(ctx, viewer); err != nil {
			logger.Error("创建观众记录失败",
				logger.ErrorField(err),
				logger.Int64Field("room_id", roomID),
				logger.Int64Field("user_id", userID))
		}

		if err := s.roomRepo.UpdateViewerCount(ctx, roomID, 1); err != nil {
			logger.Error("增加观众数失败",
				logger.ErrorField(err),
				logger.Int64Field("room_id", roomID))
		}
	}

	chatHistory := []string{
		fmt.Sprintf("欢迎来到直播间 %d", roomID),
		"主播正在直播中，欢迎互动！",
	}

	logger.Info("加入直播间成功",
		logger.Int64Field("room_id", roomID),
		logger.Int64Field("user_id", userID),
		logger.StringField("hls_url", room.HlsURL))

	return room.HlsURL, chatHistory, nil
}

// 离开直播间
func (s *liveServiceImpl) LeaveLiveRoom(ctx context.Context, roomID, userID int64) error {
	logger.Info("离开直播间请求",
		logger.Int64Field("room_id", roomID),
		logger.Int64Field("user_id", userID))

	viewer, err := s.roomViewerRepo.Find(ctx, roomID, userID)
	if err != nil {
		logger.Error("查询观众记录失败",
			logger.ErrorField(err),
			logger.Int64Field("room_id", roomID),
			logger.Int64Field("user_id", userID))
		return ErrInternalServer
	}

	if viewer == nil || viewer.LeaveTime != nil {
		logger.Warn("用户不在直播间",
			logger.Int64Field("room_id", roomID),
			logger.Int64Field("user_id", userID))
		return nil
	}

	leaveTime := time.Now().Format("2006-01-02 15:04:05")
	if err := s.roomViewerRepo.UpdateLeaveTime(ctx, roomID, userID, leaveTime); err != nil {
		logger.Error("更新离开时间失败",
			logger.ErrorField(err),
			logger.Int64Field("room_id", roomID),
			logger.Int64Field("user_id", userID))
		return ErrInternalServer
	}

	if err := s.roomRepo.UpdateViewerCount(ctx, roomID, -1); err != nil {
		logger.Error("减少观众数失败",
			logger.ErrorField(err),
			logger.Int64Field("room_id", roomID))
	}

	logger.Info("离开直播间成功",
		logger.Int64Field("room_id", roomID),
		logger.Int64Field("user_id", userID))

	return nil
}

// 发送礼物
func (s *liveServiceImpl) SendGift(ctx context.Context, senderID, roomID, giftID int64, count int32) (int64, error) {
	logger.Info("发送礼物请求",
		logger.Int64Field("sender_id", senderID),
		logger.Int64Field("room_id", roomID),
		logger.Int64Field("gift_id", giftID),
		logger.AnyField("count", count))

	if count <= 0 {
		return 0, ErrInvalidParameter
	}

	gift, err := s.giftRepo.FindByID(ctx, giftID)
	if err != nil {
		logger.Error("查询礼物失败",
			logger.ErrorField(err),
			logger.Int64Field("gift_id", giftID))
		return 0, ErrInternalServer
	}

	if gift == nil {
		logger.Warn("礼物不存在",
			logger.Int64Field("gift_id", giftID))
		return 0, ErrGiftNotFound
	}

	totalPrice := gift.Price * int64(count)

	giftRecord := &model.GiftRecord{
		SenderID:   senderID,
		RoomID:     roomID,
		GiftID:     giftID,
		Count:      count,
		TotalPrice: totalPrice,
	}

	if err := s.giftRecordRepo.Create(ctx, giftRecord); err != nil {
		logger.Error("创建礼物记录失败",
			logger.ErrorField(err),
			logger.Int64Field("sender_id", senderID),
			logger.Int64Field("room_id", roomID),
			logger.Int64Field("gift_id", giftID))
		return 0, ErrInternalServer
	}

	logger.Info("发送礼物成功",
		logger.Int64Field("sender_id", senderID),
		logger.Int64Field("room_id", roomID),
		logger.Int64Field("gift_id", giftID),
		logger.Int64Field("total_price", totalPrice))

	return totalPrice, nil
}

// 获取礼物列表
func (s *liveServiceImpl) GetGiftList(ctx context.Context) ([]*model.Gift, error) {
	logger.Info("获取礼物列表请求")

	gifts, err := s.giftRepo.ListAll(ctx)
	if err != nil {
		logger.Error("获取礼物列表失败",
			logger.ErrorField(err))
		return nil, ErrInternalServer
	}

	logger.Info("获取礼物列表成功",
		logger.IntField("gift_count", len(gifts)))

	return gifts, nil
}

// 设置管理员
func (s *liveServiceImpl) SetRoomAdmin(ctx context.Context, hostID, roomID, targetUserID int64, action bool) error {
	logger.Info("设置管理员请求",
		logger.Int64Field("host_id", hostID),
		logger.Int64Field("room_id", roomID),
		logger.Int64Field("target_user_id", targetUserID),
		logger.BoolField("action", action))

	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		logger.Error("查询直播间失败",
			logger.ErrorField(err),
			logger.Int64Field("room_id", roomID))
		return ErrInternalServer
	}

	if room == nil {
		logger.Warn("直播间不存在",
			logger.Int64Field("room_id", roomID))
		return ErrRoomNotFound
	}

	if room.HostID != hostID {
		logger.Warn("不是直播间房主",
			logger.Int64Field("host_id", hostID),
			logger.Int64Field("room_id", roomID),
			logger.Int64Field("actual_host_id", room.HostID))
		return ErrNotRoomOwner
	}

	if action {
		admin := &model.RoomAdmin{
			RoomID: roomID,
			UserID: targetUserID,
		}
		if err := s.roomAdminRepo.Create(ctx, admin); err != nil {
			logger.Error("添加管理员失败",
				logger.ErrorField(err),
				logger.Int64Field("room_id", roomID),
				logger.Int64Field("target_user_id", targetUserID))
			return ErrInternalServer
		}
	} else {
		if err := s.roomAdminRepo.Delete(ctx, roomID, targetUserID); err != nil {
			logger.Error("移除管理员失败",
				logger.ErrorField(err),
				logger.Int64Field("room_id", roomID),
				logger.Int64Field("target_user_id", targetUserID))
			return ErrInternalServer
		}
	}

	logger.Info("设置管理员成功",
		logger.Int64Field("host_id", hostID),
		logger.Int64Field("room_id", roomID),
		logger.Int64Field("target_user_id", targetUserID),
		logger.BoolField("action", action))

	return nil
}

// 录制直播
func (s *liveServiceImpl) RecordLive(ctx context.Context, hostID, roomID int64, action bool) (string, error) {
	logger.Info("录制直播请求",
		logger.Int64Field("host_id", hostID),
		logger.Int64Field("room_id", roomID),
		logger.BoolField("action", action))

	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		logger.Error("查询直播间失败",
			logger.ErrorField(err),
			logger.Int64Field("room_id", roomID))
		return "", ErrInternalServer
	}

	if room == nil {
		logger.Warn("直播间不存在",
			logger.Int64Field("room_id", roomID))
		return "", ErrRoomNotFound
	}

	if room.HostID != hostID {
		logger.Warn("不是直播间房主",
			logger.Int64Field("host_id", hostID),
			logger.Int64Field("room_id", roomID),
			logger.Int64Field("actual_host_id", room.HostID))
		return "", ErrNotRoomOwner
	}
	//不是实际业务
	if action {
		logger.Info("开始录制直播",
			logger.Int64Field("room_id", roomID))
		return "", nil
	} else {
		videoURL := fmt.Sprintf("/recordings/%d_%d.mp4", roomID, time.Now().Unix())

		liveRecord := &model.LiveRecord{
			RoomID:    roomID,
			VideoURL:  videoURL,
			StartTime: time.Now().Add(-time.Hour),
			EndTime:   time.Now(),
			Duration:  3600,
		}

		if err := s.liveRecordRepo.Create(ctx, liveRecord); err != nil {
			logger.Error("创建录制记录失败",
				logger.ErrorField(err),
				logger.Int64Field("room_id", roomID))
			return "", ErrInternalServer
		}

		logger.Info("停止录制直播成功",
			logger.Int64Field("room_id", roomID),
			logger.StringField("video_url", videoURL))

		return videoURL, nil
	}
}

// 事务相关
func (s *liveServiceImpl) WithTransaction(ctx context.Context, fn func(txService LiveService) error) error {
	return s.roomRepo.WithTransaction(ctx, func(txRoomRepo dao.LiveRoomRepository) error {
		txService := &liveServiceImpl{
			roomRepo:       txRoomRepo,
			giftRepo:       s.giftRepo,
			giftRecordRepo: s.giftRecordRepo,
			roomAdminRepo:  s.roomAdminRepo,
			liveRecordRepo: s.liveRecordRepo,
			roomViewerRepo: s.roomViewerRepo,
		}
		return fn(txService)
	})
}

func ConvertToCommonLiveRoom(room *model.LiveRoom) *common.LiveRoom {
	if room == nil {
		return nil
	}

	return &common.LiveRoom{
		Id:          room.ID,
		HostId:      room.HostID,
		Title:       room.Title,
		CoverUrl:    room.CoverURL,
		RtmpUrl:     room.RtmpURL,
		HlsUrl:      room.HlsURL,
		ViewerCount: room.ViewerCount,
		IsLive:      room.IsLive,
		CreateTime:  room.CreateTime,
	}
}

func ConvertToLiveGift(gift *model.Gift) *live.Gift {
	if gift == nil {
		return nil
	}

	return &live.Gift{
		Id:           gift.ID,
		Name:         gift.Name,
		Price:        gift.Price,
		IconUrl:      gift.IconURL,
		AnimationUrl: gift.AnimationURL,
	}
}
