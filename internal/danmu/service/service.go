package service

import (
	"context"
	"encoding/json"
	"errors"
	"shortvideo/internal/danmu/dao"
	"shortvideo/internal/danmu/model"
	"shortvideo/kitex_gen/common"
	"shortvideo/kitex_gen/danmu"
	"shortvideo/pkg/logger"
	"time"
)

var (
	ErrDanmuNotFound    = errors.New("弹幕不存在")
	ErrFilterNotFound   = errors.New("过滤设置不存在")
	ErrInternalServer   = errors.New("服务器内部错误")
	ErrInvalidParameter = errors.New("参数错误")
	ErrLiveNotFound     = errors.New("直播间不存在")
	ErrNotRoomAdmin     = errors.New("不是直播间管理员")
)

type DanmuService interface {
	//弹幕相关
	SendDanmu(ctx context.Context, userID, liveID int64, content, color string) (int64, error)
	GetDanmuHistory(ctx context.Context, liveID int64, startTime, endTime string, limit int) ([]*model.Danmu, error)
	ManageDanmu(ctx context.Context, managerID, liveID, danmuID int64, action int32) error
	LikeDanmu(ctx context.Context, userID, danmuID int64) (int64, error)

	//过滤相关
	SetDanmuFilter(ctx context.Context, userID, liveID int64, keywords []string, hideAnonymous, hideLowLevel bool) error
	GetDanmuFilter(ctx context.Context, userID, liveID int64) (*danmu.DanmuFilter, error)

	//统计相关
	GetDanmuStats(ctx context.Context, liveID int64) (*danmu.DanmuStats, error)

	//事务相关
	WithTransaction(ctx context.Context, fn func(txService DanmuService) error) error
}

type danmuServiceImpl struct {
	danmuRepo  dao.DanmuRepository
	filterRepo dao.DanmuFilterRepository
}

func NewDanmuService(
	danmuRepo dao.DanmuRepository,
	filterRepo dao.DanmuFilterRepository,
) DanmuService {
	return &danmuServiceImpl{
		danmuRepo:  danmuRepo,
		filterRepo: filterRepo,
	}
}

func NewDanmuServiceWithRepo(
	danmuRepo dao.DanmuRepository,
	filterRepo dao.DanmuFilterRepository,
) DanmuService {
	return &danmuServiceImpl{
		danmuRepo:  danmuRepo,
		filterRepo: filterRepo,
	}
}

// 发送弹幕
func (s *danmuServiceImpl) SendDanmu(ctx context.Context, userID, liveID int64, content, color string) (int64, error) {
	logger.Info("SendDanmu request",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("live_id", liveID),
		logger.StringField("content", content))

	if content == "" {
		return 0, ErrInvalidParameter
	}

	if color == "" {
		color = "#FFFFFF"
	}

	danmu := &model.Danmu{
		UserID:     userID,
		LiveID:     liveID,
		Content:    content,
		Color:      color,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := s.danmuRepo.Create(ctx, danmu); err != nil {
		logger.Error("SendDanmu failed", logger.ErrorField(err))
		return 0, ErrInternalServer
	}

	logger.Info("SendDanmu success", logger.Int64Field("danmu_id", danmu.ID))
	return danmu.ID, nil
}

// 获取弹幕历史
func (s *danmuServiceImpl) GetDanmuHistory(ctx context.Context, liveID int64, startTime, endTime string, limit int) ([]*model.Danmu, error) {
	logger.Info("GetDanmuHistory request",
		logger.Int64Field("live_id", liveID),
		logger.StringField("start_time", startTime),
		logger.StringField("end_time", endTime),
		logger.IntField("limit", limit))

	if liveID <= 0 {
		return nil, ErrInvalidParameter
	}

	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	danmus, err := s.danmuRepo.ListByLiveIDAndTime(ctx, liveID, startTime, endTime, limit)
	if err != nil {
		logger.Error("GetDanmuHistory failed", logger.ErrorField(err))
		return nil, ErrInternalServer
	}

	logger.Info("GetDanmuHistory success", logger.IntField("danmu_count", len(danmus)))
	return danmus, nil
}

// 管理弹幕
func (s *danmuServiceImpl) ManageDanmu(ctx context.Context, managerID, liveID, danmuID int64, action int32) error {
	logger.Info("ManageDanmu request",
		logger.Int64Field("manager_id", managerID),
		logger.Int64Field("live_id", liveID),
		logger.Int64Field("danmu_id", danmuID),
		logger.AnyField("action", action))

	if managerID <= 0 || liveID <= 0 || danmuID <= 0 {
		return ErrInvalidParameter
	}

	danmu, err := s.danmuRepo.FindByID(ctx, danmuID)
	if err != nil {
		logger.Error("FindDanmu failed", logger.ErrorField(err))
		return ErrInternalServer
	}

	if danmu == nil {
		logger.Warn("Danmu not found", logger.Int64Field("danmu_id", danmuID))
		return ErrDanmuNotFound
	}

	if danmu.LiveID != liveID {
		logger.Warn("Danmu not belong to live",
			logger.Int64Field("danmu_id", danmuID),
			logger.Int64Field("live_id", liveID),
			logger.Int64Field("actual_live_id", danmu.LiveID))
		return ErrInvalidParameter
	}

	switch action {
	case 1:
		if err := s.danmuRepo.Delete(ctx, danmuID); err != nil {
			logger.Error("DeleteDanmu failed", logger.ErrorField(err))
			return ErrInternalServer
		}
		//禁言//置顶横幅
	case 2:
	case 3:
	default:
		return ErrInvalidParameter
	}

	logger.Info("ManageDanmu success",
		logger.Int64Field("danmu_id", danmuID),
		logger.AnyField("action", action))
	return nil
}

// 点赞弹幕
func (s *danmuServiceImpl) LikeDanmu(ctx context.Context, userID, danmuID int64) (int64, error) {
	logger.Info("LikeDanmu request",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("danmu_id", danmuID))

	if userID <= 0 || danmuID <= 0 {
		return 0, ErrInvalidParameter
	}

	danmu, err := s.danmuRepo.FindByID(ctx, danmuID)
	if err != nil {
		logger.Error("FindDanmu failed", logger.ErrorField(err))
		return 0, ErrInternalServer
	}

	if danmu == nil {
		logger.Warn("Danmu not found", logger.Int64Field("danmu_id", danmuID))
		return 0, ErrDanmuNotFound
	}
	//不是实际业务
	likeCount := int64(1)

	logger.Info("LikeDanmu success",
		logger.Int64Field("danmu_id", danmuID),
		logger.Int64Field("like_count", likeCount))
	return likeCount, nil
}

// 设置弹幕过滤
func (s *danmuServiceImpl) SetDanmuFilter(ctx context.Context, userID, liveID int64, keywords []string, hideAnonymous, hideLowLevel bool) error {
	logger.Info("SetDanmuFilter request",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("live_id", liveID),
		logger.IntField("keyword_count", len(keywords)))

	if userID <= 0 || liveID <= 0 {
		return ErrInvalidParameter
	}

	keywordsJSON, err := json.Marshal(keywords)
	if err != nil {
		logger.Error("MarshalKeywords failed", logger.ErrorField(err))
		return ErrInternalServer
	}

	filter := &model.DanmuFilter{
		UserID:        userID,
		LiveID:        liveID,
		Keywords:      string(keywordsJSON),
		HideAnonymous: hideAnonymous,
		HideLowLevel:  hideLowLevel,
	}

	if err := s.filterRepo.CreateOrUpdate(ctx, filter); err != nil {
		logger.Error("CreateOrUpdateFilter failed", logger.ErrorField(err))
		return ErrInternalServer
	}

	logger.Info("SetDanmuFilter success",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("live_id", liveID))
	return nil
}

// 获取弹幕过滤设置
func (s *danmuServiceImpl) GetDanmuFilter(ctx context.Context, userID, liveID int64) (*danmu.DanmuFilter, error) {
	logger.Info("GetDanmuFilter request",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("live_id", liveID))

	if userID <= 0 || liveID <= 0 {
		return nil, ErrInvalidParameter
	}

	filter, err := s.filterRepo.FindByUserAndLive(ctx, userID, liveID)
	if err != nil {
		logger.Error("FindFilter failed", logger.ErrorField(err))
		return nil, ErrInternalServer
	}

	if filter == nil {
		logger.Warn("Filter not found",
			logger.Int64Field("user_id", userID),
			logger.Int64Field("live_id", liveID))
		return nil, ErrFilterNotFound
	}

	var keywords []string
	if filter.Keywords != "" {
		if err := json.Unmarshal([]byte(filter.Keywords), &keywords); err != nil {
			logger.Error("UnmarshalKeywords failed", logger.ErrorField(err))
			keywords = []string{}
		}
	}

	danmuFilter := &danmu.DanmuFilter{
		UserId:        filter.UserID,
		LiveId:        filter.LiveID,
		Keywords:      keywords,
		HideAnonymous: filter.HideAnonymous,
		HideLowLevel:  filter.HideLowLevel,
	}

	logger.Info("GetDanmuFilter success",
		logger.Int64Field("user_id", userID),
		logger.Int64Field("live_id", liveID))
	return danmuFilter, nil
}

// 获取弹幕统计
func (s *danmuServiceImpl) GetDanmuStats(ctx context.Context, liveID int64) (*danmu.DanmuStats, error) {
	logger.Info("GetDanmuStats request", logger.Int64Field("live_id", liveID))

	if liveID <= 0 {
		return nil, ErrInvalidParameter
	}

	stats, err := s.danmuRepo.GetDanmuStats(ctx, liveID)
	if err != nil {
		logger.Error("GetDanmuStats failed", logger.ErrorField(err))
		return nil, ErrInternalServer
	}

	danmuStats := &danmu.DanmuStats{
		LiveId:             stats.LiveID,
		TotalDanmuCount:    stats.TotalDanmuCount,
		ActiveUserCount:    stats.ActiveUserCount,
		PeakDanmuPerMinute: 0,
		WordCloud:          make(map[string]int64),
	}

	logger.Info("GetDanmuStats success",
		logger.Int64Field("live_id", liveID),
		logger.Int64Field("total_danmu_count", stats.TotalDanmuCount),
		logger.Int64Field("active_user_count", stats.ActiveUserCount))
	return danmuStats, nil
}

// 事务相关
func (s *danmuServiceImpl) WithTransaction(ctx context.Context, fn func(txService DanmuService) error) error {
	return s.danmuRepo.WithTransaction(ctx, func(txDanmuRepo dao.DanmuRepository) error {
		txService := &danmuServiceImpl{
			danmuRepo:  txDanmuRepo,
			filterRepo: s.filterRepo,
		}
		return fn(txService)
	})
}

// 转换为common.Danmu
func ConvertToCommonDanmu(danmu *model.Danmu) *common.Danmu {
	if danmu == nil {
		return nil
	}

	return &common.Danmu{
		Id:         danmu.ID,
		UserId:     danmu.UserID,
		LiveId:     danmu.LiveID,
		Content:    danmu.Content,
		Color:      danmu.Color,
		CreateTime: danmu.CreateTime,
	}
}

// 转换为danmu.DanmuFilter
func ConvertToDanmuFilter(filter *model.DanmuFilter) (*danmu.DanmuFilter, error) {
	if filter == nil {
		return nil, nil
	}

	var keywords []string
	if filter.Keywords != "" {
		if err := json.Unmarshal([]byte(filter.Keywords), &keywords); err != nil {
			logger.Error("UnmarshalKeywords failed", logger.ErrorField(err))
			keywords = []string{}
		}
	}

	return &danmu.DanmuFilter{
		UserId:        filter.UserID,
		LiveId:        filter.LiveID,
		Keywords:      keywords,
		HideAnonymous: filter.HideAnonymous,
		HideLowLevel:  filter.HideLowLevel,
	}, nil
}

// 转换为danmu.DanmuStats
func ConvertToDanmuStats(stats *model.DanmuStats) *danmu.DanmuStats {
	if stats == nil {
		return nil
	}

	return &danmu.DanmuStats{
		LiveId:             stats.LiveID,
		TotalDanmuCount:    stats.TotalDanmuCount,
		ActiveUserCount:    stats.ActiveUserCount,
		PeakDanmuPerMinute: 0,
		WordCloud:          make(map[string]int64),
	}
}
