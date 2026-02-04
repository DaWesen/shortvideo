namespace go recommend

include "common.thrift"

struct GetRecommendVideosReq{
    1:i64 userId
    2:i32 pageSize
    3:optional i64 offset
}

struct GetRecommendVideosResp{
    1:common.BaseResp BaseResp
    2:list<common.Video> videos
    3:i64 nextOffset
}

struct GetRecommendUsersReq{
    1:i64 userId
    2:i32 count
}

struct GetRecommendUsersResp{
    1:common.BaseResp BaseResp
    2:list<common.User> users
}

struct UserActionReq{
    1:i64 userId
    2:i64 itemId
    3:string itemType
    4:string actionType
    5:optional i32 duration
    6:optional double score
    7:string timestamp
}

struct UserActionResp{
    1:common.BaseResp BaseResp
}

struct GetHotTagsReq{
    1:i32 count
}

struct TagInfo{
    1:string tagName
    2:i64 videoCount
    3:i64 viewCount
}

struct GetHotTagsResp{
    1:common.BaseResp BaseResp
    2:list<TagInfo> tags
}

struct GetTagVideosReq{
    1:string tag
    2:i64 userId
    3:i32 pageSize
}

struct GetTagVideosResp{
    1:common.BaseResp BaseResp
    2:list<common.Video> videos
}

struct GetPersonalizedFeedReq{
    1:i64 userId
    2:i32 pageSize
    3:optional i64 lastVideoId
}

struct GetPersonalizedFeedResp{
    1:common.BaseResp BaseResp
    2:list<common.Video> videos
    3:i64 nextLastVideoId
}

service RecommendService{
    GetRecommendVideosResp GetRecommendVideos(1:GetRecommendVideosReq req)
    GetRecommendUsersResp GetRecommendUsers(1:GetRecommendUsersReq req)
    UserActionResp RecordUserAction(1:UserActionReq req)
    GetHotTagsResp GetHotTags(1:GetHotTagsReq req)
    GetTagVideosResp GetTagVideos(1:GetTagVideosReq req)
    GetPersonalizedFeedResp GetPersonalizedFeed(1:GetPersonalizedFeedReq req)
}