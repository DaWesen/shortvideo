package handler

import (
	"context"
	interaction "shortvideo/kitex_gen/interaction"
)

// InteractionServiceImpl implements the last service interface defined in the IDL.
type InteractionServiceImpl struct{}

func NewInteractionService() *InteractionServiceImpl {
	return &InteractionServiceImpl{}
}

// LikeAction implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) LikeAction(ctx context.Context, req *interaction.LikeActionReq) (resp *interaction.LikeActionResp, err error) {
	// TODO: Your code here...
	return
}

// GetLikeVideoList implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) GetLikeVideoList(ctx context.Context, req *interaction.LikeVideoListReq) (resp *interaction.LikeVideoListResp, err error) {
	// TODO: Your code here...
	return
}

// StarAction implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) StarAction(ctx context.Context, req *interaction.StarActionReq) (resp *interaction.StarActionResp, err error) {
	// TODO: Your code here...
	return
}

// GetStarVideoList implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) GetStarVideoList(ctx context.Context, req *interaction.StarVideoListReq) (resp *interaction.StarVideoListResp, err error) {
	// TODO: Your code here...
	return
}

// CommentAction implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) CommentAction(ctx context.Context, req *interaction.CommentActionReq) (resp *interaction.CommentActionResp, err error) {
	// TODO: Your code here...
	return
}

// GetCommentList implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) GetCommentList(ctx context.Context, req *interaction.CommentListReq) (resp *interaction.CommentListResp, err error) {
	// TODO: Your code here...
	return
}

// DeleteComment implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) DeleteComment(ctx context.Context, req *interaction.DeleteCommentReq) (resp *interaction.DeleteCommentResp, err error) {
	// TODO: Your code here...
	return
}

// ShareAction implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) ShareAction(ctx context.Context, req *interaction.ShareActionReq) (resp *interaction.ShareActionResp, err error) {
	// TODO: Your code here...
	return
}

// GetCount implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) GetCount(ctx context.Context, req *interaction.CountReq) (resp *interaction.CountResp, err error) {
	// TODO: Your code here...
	return
}

// CheckLikeStatus implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) CheckLikeStatus(ctx context.Context, req *interaction.CheckLikeStatusReq) (resp *interaction.CheckLikeStatusResp, err error) {
	// TODO: Your code here...
	return
}

// CheckStarStatus implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) CheckStarStatus(ctx context.Context, req *interaction.CheckStarStatusReq) (resp *interaction.CheckStarStatusResp, err error) {
	// TODO: Your code here...
	return
}
