package handler

import (
	"context"
	"shortvideo/internal/danmu/service"
	"shortvideo/kitex_gen/common"
	danmu "shortvideo/kitex_gen/danmu"
	"shortvideo/pkg/logger"
	"time"
)

// DanmuServiceImpl implements the last service interface defined in the IDL.
type DanmuServiceImpl struct {
	danmuService service.DanmuService
}

func NewDanmuService(danmuService service.DanmuService) *DanmuServiceImpl {
	return &DanmuServiceImpl{
		danmuService: danmuService,
	}
}

// SendDanmu implements the DanmuServiceImpl interface.
func (s *DanmuServiceImpl) SendDanmu(ctx context.Context, req *danmu.SendDanmuReq) (resp *danmu.SendDanmuResp, err error) {
	logger.Info("SendDanmu request",
		logger.Int64Field("user_id", req.UserId),
		logger.Int64Field("live_id", req.LiveId),
		logger.StringField("content", req.Content))

	successMsg := "成功"
	resp = &danmu.SendDanmuResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		DanmuId: 0,
	}

	danmuID, err := s.danmuService.SendDanmu(ctx, req.UserId, req.LiveId, req.Content, req.GetColor())
	if err != nil {
		logger.Error("SendDanmu failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.DanmuId = danmuID
	logger.Info("SendDanmu success", logger.Int64Field("danmu_id", danmuID))
	return resp, nil
}

// GetDanmuHistory implements the DanmuServiceImpl interface.
func (s *DanmuServiceImpl) GetDanmuHistory(ctx context.Context, req *danmu.GetDanmuHistoryReq) (resp *danmu.GetDanmuHistoryResp, err error) {
	logger.Info("GetDanmuHistory request",
		logger.Int64Field("live_id", req.LiveId),
		logger.Int64Field("start_time", req.StartTime),
		logger.Int64Field("end_time", req.EndTime),
		logger.AnyField("limit", req.Limit))

	successMsg := "成功"
	resp = &danmu.GetDanmuHistoryResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Danmus: []*common.Danmu{},
	}

	startTime := time.Unix(req.StartTime/1000, 0).Format("2006-01-02 15:04:05")
	endTime := time.Unix(req.EndTime/1000, 0).Format("2006-01-02 15:04:05")

	danmus, err := s.danmuService.GetDanmuHistory(ctx, req.LiveId, startTime, endTime, int(req.Limit))
	if err != nil {
		logger.Error("GetDanmuHistory failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	commonDanmus := make([]*common.Danmu, len(danmus))
	for i, danmu := range danmus {
		commonDanmus[i] = service.ConvertToCommonDanmu(danmu)
	}

	resp.Danmus = commonDanmus
	logger.Info("GetDanmuHistory success", logger.IntField("danmu_count", len(danmus)))
	return resp, nil
}

// SetDanmuFilter implements the DanmuServiceImpl interface.
func (s *DanmuServiceImpl) SetDanmuFilter(ctx context.Context, req *danmu.SetDanmuFilterReq) (resp *danmu.SetDanmuFilterResp, err error) {
	logger.Info("SetDanmuFilter request",
		logger.Int64Field("user_id", req.UserId),
		logger.Int64Field("live_id", req.LiveId))

	successMsg := "成功"
	resp = &danmu.SetDanmuFilterResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
	}

	err = s.danmuService.SetDanmuFilter(ctx, req.UserId, req.LiveId, req.Filter.Keywords, req.Filter.HideAnonymous, req.Filter.HideLowLevel)
	if err != nil {
		logger.Error("SetDanmuFilter failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	logger.Info("SetDanmuFilter success",
		logger.Int64Field("user_id", req.UserId),
		logger.Int64Field("live_id", req.LiveId))
	return resp, nil
}

// GetDanmuFilter implements the DanmuServiceImpl interface.
func (s *DanmuServiceImpl) GetDanmuFilter(ctx context.Context, req *danmu.GetDanmuFilterReq) (resp *danmu.GetDanmuFilterResp, err error) {
	logger.Info("GetDanmuFilter request",
		logger.Int64Field("user_id", req.UserId),
		logger.Int64Field("live_id", req.LiveId))

	successMsg := "成功"
	resp = &danmu.GetDanmuFilterResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Filter: nil,
	}

	filter, err := s.danmuService.GetDanmuFilter(ctx, req.UserId, req.LiveId)
	if err != nil {
		logger.Error("GetDanmuFilter failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.Filter = filter
	logger.Info("GetDanmuFilter success",
		logger.Int64Field("user_id", req.UserId),
		logger.Int64Field("live_id", req.LiveId))
	return resp, nil
}

// GetDanmuStats implements the DanmuServiceImpl interface.
func (s *DanmuServiceImpl) GetDanmuStats(ctx context.Context, req *danmu.GetDanmuStatsReq) (resp *danmu.GetDanmuStatsResp, err error) {
	logger.Info("GetDanmuStats request", logger.Int64Field("live_id", req.LiveId))

	successMsg := "成功"
	resp = &danmu.GetDanmuStatsResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Stats: nil,
	}

	stats, err := s.danmuService.GetDanmuStats(ctx, req.LiveId)
	if err != nil {
		logger.Error("GetDanmuStats failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.Stats = stats
	logger.Info("GetDanmuStats success", logger.Int64Field("live_id", req.LiveId))
	return resp, nil
}

// ManageDanmu implements the DanmuServiceImpl interface.
func (s *DanmuServiceImpl) ManageDanmu(ctx context.Context, req *danmu.ManageDanmuReq) (resp *danmu.ManageDanmuResp, err error) {
	logger.Info("ManageDanmu request",
		logger.Int64Field("manager_id", req.ManagerId),
		logger.Int64Field("live_id", req.LiveId),
		logger.Int64Field("danmu_id", req.DanmuId),
		logger.AnyField("action", req.Action))

	successMsg := "成功"
	resp = &danmu.ManageDanmuResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
	}

	err = s.danmuService.ManageDanmu(ctx, req.ManagerId, req.LiveId, req.DanmuId, req.Action)
	if err != nil {
		logger.Error("ManageDanmu failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	logger.Info("ManageDanmu success",
		logger.Int64Field("danmu_id", req.DanmuId),
		logger.AnyField("action", req.Action))
	return resp, nil
}

// LikeDanmu implements the DanmuServiceImpl interface.
func (s *DanmuServiceImpl) LikeDanmu(ctx context.Context, req *danmu.LikeDanmuReq) (resp *danmu.LikeDanmuResp, err error) {
	logger.Info("LikeDanmu request",
		logger.Int64Field("user_id", req.UserId),
		logger.Int64Field("danmu_id", req.DanmuId))

	successMsg := "成功"
	resp = &danmu.LikeDanmuResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		LikeCount: 0,
	}

	likeCount, err := s.danmuService.LikeDanmu(ctx, req.UserId, req.DanmuId)
	if err != nil {
		logger.Error("LikeDanmu failed", logger.ErrorField(err))
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.LikeCount = likeCount
	logger.Info("LikeDanmu success",
		logger.Int64Field("danmu_id", req.DanmuId),
		logger.Int64Field("like_count", likeCount))
	return resp, nil
}
