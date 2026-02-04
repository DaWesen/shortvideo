namespace go interaction

include "common.thrift"

struct LikeActionReq{
    1:i64 userId
    2:i64 videoId
    3:bool action
}

struct LikeActionResp{
    1:common.BaseResp BaseResp
}

struct LikeVideoListReq{
    1:i64 userId
    2:i64 currentUserId
    3:i32 page
    4:i32 pageSize
}

struct LikeVideoListResp{
    1:common.BaseResp BaseResp
    2:list<common.Video> videos
    3:i32 totalCount
}

struct StarActionReq{
    1:i64 userId
    2:i64 videoId
    3:bool action
}

struct StarActionResp{
    1:common.BaseResp BaseResp
}

struct StarVideoListReq{
    1:i64 userId
    2:i64 currentUserId
    3:i32 page
    4:i32 pageSize
}

struct StarVideoListResp{
    1:common.BaseResp BaseResp
    2:list<common.Video> videos
    3:i32 totalCount
}

struct CommentActionReq{
    1:i64 userId
    2:i64 videoId
    3:string content
    4:optional i64 replyToId
}

struct CommentActionResp{
    1:common.BaseResp BaseResp
    2:common.Comment comment
}

struct CommentListReq{
    1:i64 videoId
    2:i64 currentUserId
    3:i32 page
    4:i32 pageSize
}

struct CommentListResp{
    1:common.BaseResp BaseResp
    2:list<common.Comment> comments
    3:i32 totalCount
}

struct DeleteCommentReq{
    1:i64 userId
    2:i64 videoId
    3:i64 commentId
}

struct DeleteCommentResp{
    1:common.BaseResp BaseResp
}

struct ShareActionReq{
    1:i64 userId
    2:i64 videoId
}

struct ShareActionResp{
    1:common.BaseResp BaseResp
}

struct CountReq{
    1:i64 videoId
}

struct CountResp{
    1:common.BaseResp BaseResp
    2:i64 likeCount
    3:i64 commentCount
    4:i64 starCount
    5:i64 shareCount
}

struct CheckLikeStatusReq{
    1:i64 userId
    2:i64 videoId
}

struct CheckLikeStatusResp{
    1:common.BaseResp BaseResp
    2:bool isLiked
}

struct CheckStarStatusReq{
    1:i64 userId
    2:i64 videoId
}

struct CheckStarStatusResp{
    1:common.BaseResp BaseResp
    2:bool isStarred
}

service InteractionService{
    LikeActionResp LikeAction(1:LikeActionReq req)
    LikeVideoListResp GetLikeVideoList(1:LikeVideoListReq req)
    StarActionResp StarAction(1:StarActionReq req)
    StarVideoListResp GetStarVideoList(1:StarVideoListReq req)
    CommentActionResp CommentAction(1:CommentActionReq req)
    CommentListResp GetCommentList(1:CommentListReq req)
    DeleteCommentResp DeleteComment(1:DeleteCommentReq req)
    ShareActionResp ShareAction(1:ShareActionReq req)
    CountResp GetCount(1:CountReq req)
    CheckLikeStatusResp CheckLikeStatus(1:CheckLikeStatusReq req)
    CheckStarStatusResp CheckStarStatus(1:CheckStarStatusReq req)
}