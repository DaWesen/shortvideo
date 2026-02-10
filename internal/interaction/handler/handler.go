package handler

import (
	"context"
	"shortvideo/internal/interaction/service"
	"shortvideo/kitex_gen/common"
	interaction "shortvideo/kitex_gen/interaction"
)

// InteractionServiceImpl implements the last service interface defined in the IDL.
type InteractionServiceImpl struct {
	interactionService service.InteractionService
}

func NewInteractionService(interactionService service.InteractionService) *InteractionServiceImpl {
	return &InteractionServiceImpl{
		interactionService: interactionService,
	}
}

// LikeAction implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) LikeAction(ctx context.Context, req *interaction.LikeActionReq) (resp *interaction.LikeActionResp, err error) {
	// TODO: Your code here...
	successMsg := "成功"
	resp = &interaction.LikeActionResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
	}
	err = s.interactionService.LikeAction(ctx, req.UserId, req.VideoId, req.Action)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	return resp, nil
}

// GetLikeVideoList implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) GetLikeVideoList(ctx context.Context, req *interaction.LikeVideoListReq) (resp *interaction.LikeVideoListResp, err error) {
	successMsg := "成功"
	resp = &interaction.LikeVideoListResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Videos:     []*common.Video{},
		TotalCount: 0,
	}

	videos, total, err := s.interactionService.GetLikeVideoList(ctx, req.UserId, req.CurrentUserId, int(req.Page), int(req.PageSize))
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	commonVideos := make([]*common.Video, len(videos))
	for i, v := range videos {
		commonVideos[i] = &common.Video{
			Id:           v.ID,
			AuthorId:     v.AuthorID,
			Url:          v.URL,
			CoverUrl:     v.CoverURL,
			Title:        v.Title,
			Description:  v.Description,
			LikeCount:    v.LikeCount,
			CommentCount: v.CommentCount,
			PublishTime:  v.PublishTime,
		}
	}

	resp.Videos = commonVideos
	resp.TotalCount = int32(total)
	return resp, nil
}

// StarAction implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) StarAction(ctx context.Context, req *interaction.StarActionReq) (resp *interaction.StarActionResp, err error) {
	successMsg := "成功"
	resp = &interaction.StarActionResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
	}

	err = s.interactionService.StarAction(ctx, req.UserId, req.VideoId, req.Action)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	return resp, nil
}

// GetStarVideoList implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) GetStarVideoList(ctx context.Context, req *interaction.StarVideoListReq) (resp *interaction.StarVideoListResp, err error) {
	successMsg := "成功"
	resp = &interaction.StarVideoListResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Videos:     []*common.Video{},
		TotalCount: 0,
	}

	videos, total, err := s.interactionService.GetStarVideoList(ctx, req.UserId, req.CurrentUserId, int(req.Page), int(req.PageSize))
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	commonVideos := make([]*common.Video, len(videos))
	for i, v := range videos {
		commonVideos[i] = &common.Video{
			Id:           v.ID,
			AuthorId:     v.AuthorID,
			Url:          v.URL,
			CoverUrl:     v.CoverURL,
			Title:        v.Title,
			Description:  v.Description,
			LikeCount:    v.LikeCount,
			CommentCount: v.CommentCount,
			PublishTime:  v.PublishTime,
		}
	}

	resp.Videos = commonVideos
	resp.TotalCount = int32(total)
	return resp, nil
}

// CommentAction implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) CommentAction(ctx context.Context, req *interaction.CommentActionReq) (resp *interaction.CommentActionResp, err error) {
	successMsg := "成功"
	resp = &interaction.CommentActionResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Comment: nil,
	}

	var replyToID int64
	if req.ReplyToId != nil {
		replyToID = *req.ReplyToId
	}

	comment, err := s.interactionService.CommentAction(ctx, req.UserId, req.VideoId, req.Content, replyToID)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.Comment = &common.Comment{
		Id:         comment.ID,
		UserId:     comment.UserID,
		VideoId:    comment.VideoID,
		Content:    comment.Content,
		CreateTime: comment.CreateTime,
		ReplyToId:  comment.ReplyToID,
	}

	return resp, nil
}

// GetCommentList implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) GetCommentList(ctx context.Context, req *interaction.CommentListReq) (resp *interaction.CommentListResp, err error) {
	successMsg := "成功"
	resp = &interaction.CommentListResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		Comments:   []*common.Comment{},
		TotalCount: 0,
	}

	comments, total, err := s.interactionService.GetCommentList(ctx, req.VideoId, req.CurrentUserId, int(req.Page), int(req.PageSize))
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	commonComments := make([]*common.Comment, len(comments))
	for i, c := range comments {
		commonComments[i] = &common.Comment{
			Id:         c.ID,
			UserId:     c.UserID,
			VideoId:    c.VideoID,
			Content:    c.Content,
			CreateTime: c.CreateTime,
			ReplyToId:  c.ReplyToID,
		}
	}

	resp.Comments = commonComments
	resp.TotalCount = int32(total)
	return resp, nil
}

// DeleteComment implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) DeleteComment(ctx context.Context, req *interaction.DeleteCommentReq) (resp *interaction.DeleteCommentResp, err error) {
	successMsg := "成功"
	resp = &interaction.DeleteCommentResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
	}

	err = s.interactionService.DeleteComment(ctx, req.UserId, req.VideoId, req.CommentId)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	return resp, nil
}

// ShareAction implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) ShareAction(ctx context.Context, req *interaction.ShareActionReq) (resp *interaction.ShareActionResp, err error) {
	successMsg := "成功"
	resp = &interaction.ShareActionResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
	}

	err = s.interactionService.ShareAction(ctx, req.UserId, req.VideoId)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	return resp, nil
}

// GetCount implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) GetCount(ctx context.Context, req *interaction.CountReq) (resp *interaction.CountResp, err error) {
	successMsg := "成功"
	resp = &interaction.CountResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		LikeCount:    0,
		CommentCount: 0,
		StarCount:    0,
		ShareCount:   0,
	}

	likeCount, commentCount, starCount, shareCount, err := s.interactionService.GetCount(ctx, req.VideoId)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.LikeCount = likeCount
	resp.CommentCount = commentCount
	resp.StarCount = starCount
	resp.ShareCount = shareCount
	return resp, nil
}

// CheckLikeStatus implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) CheckLikeStatus(ctx context.Context, req *interaction.CheckLikeStatusReq) (resp *interaction.CheckLikeStatusResp, err error) {
	successMsg := "成功"
	resp = &interaction.CheckLikeStatusResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		IsLiked: false,
	}

	isLiked, err := s.interactionService.CheckLikeStatus(ctx, req.UserId, req.VideoId)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.IsLiked = isLiked
	return resp, nil
}

// CheckStarStatus implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) CheckStarStatus(ctx context.Context, req *interaction.CheckStarStatusReq) (resp *interaction.CheckStarStatusResp, err error) {
	successMsg := "成功"
	resp = &interaction.CheckStarStatusResp{
		BaseResp: &common.BaseResp{
			StatusCode: 0,
			Msg:        &successMsg,
		},
		IsStarred: false,
	}

	isStarred, err := s.interactionService.CheckStarStatus(ctx, req.UserId, req.VideoId)
	if err != nil {
		errorMsg := err.Error()
		resp.BaseResp.StatusCode = 1
		resp.BaseResp.Msg = &errorMsg
		return resp, nil
	}

	resp.IsStarred = isStarred
	return resp, nil
}
