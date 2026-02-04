package handler

import (
	"context"
	danmu "shortvideo/kitex_gen/danmu"
)

// DanmuServiceImpl implements the last service interface defined in the IDL.
type DanmuServiceImpl struct{}

func NewDanmuService() *DanmuServiceImpl {
	return &DanmuServiceImpl{}
}

// SendDanmu implements the DanmuServiceImpl interface.
func (s *DanmuServiceImpl) SendDanmu(ctx context.Context, req *danmu.SendDanmuReq) (resp *danmu.SendDanmuResp, err error) {
	// TODO: Your code here...
	return
}

// GetDanmuHistory implements the DanmuServiceImpl interface.
func (s *DanmuServiceImpl) GetDanmuHistory(ctx context.Context, req *danmu.GetDanmuHistoryReq) (resp *danmu.GetDanmuHistoryResp, err error) {
	// TODO: Your code here...
	return
}

// SetDanmuFilter implements the DanmuServiceImpl interface.
func (s *DanmuServiceImpl) SetDanmuFilter(ctx context.Context, req *danmu.SetDanmuFilterReq) (resp *danmu.SetDanmuFilterResp, err error) {
	// TODO: Your code here...
	return
}

// GetDanmuFilter implements the DanmuServiceImpl interface.
func (s *DanmuServiceImpl) GetDanmuFilter(ctx context.Context, req *danmu.GetDanmuFilterReq) (resp *danmu.GetDanmuFilterResp, err error) {
	// TODO: Your code here...
	return
}

// GetDanmuStats implements the DanmuServiceImpl interface.
func (s *DanmuServiceImpl) GetDanmuStats(ctx context.Context, req *danmu.GetDanmuStatsReq) (resp *danmu.GetDanmuStatsResp, err error) {
	// TODO: Your code here...
	return
}

// ManageDanmu implements the DanmuServiceImpl interface.
func (s *DanmuServiceImpl) ManageDanmu(ctx context.Context, req *danmu.ManageDanmuReq) (resp *danmu.ManageDanmuResp, err error) {
	// TODO: Your code here...
	return
}

// LikeDanmu implements the DanmuServiceImpl interface.
func (s *DanmuServiceImpl) LikeDanmu(ctx context.Context, req *danmu.LikeDanmuReq) (resp *danmu.LikeDanmuResp, err error) {
	// TODO: Your code here...
	return
}
