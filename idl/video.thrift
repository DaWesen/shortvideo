namespace go video

include "common.thrift"

struct PublishVideoReq{
    1:i64 userId
    2:string title
    3:string videoUrl
    4:string coverUrl
    5:string description
}

struct PublishVideoResp{
    1:common.BaseResp BaseResp
    2:i64 videoId
}

struct UserVideoListReq{
    1:i64 userId
    2:i64 currentUserId
    3:i32 page
    4:i32 pageSize
}

struct UserVideoListResp{
    1:common.BaseResp BaseResp
    2:list<common.Video> videos
    3:i32 totalCount
}

struct FeedReq{
    1:i64 userId
    2:i64 latestTime
    3:i32 pageSize
}

struct FeedResp{
    1:common.BaseResp BaseResp
    2:list<common.Video> videos
    3:i64 nextTime
}

struct SearchVideoReq{
    1:string keyword
    2:i64 currentUserId
    3:i32 page
    4:i32 pageSize
}

struct SearchVideoResp{
    1:common.BaseResp BaseResp
    2:list<common.Video> videos
    3:i32 totalCount
}

struct VideoDetailReq{
    1:i64 videoId
    2:i64 currentUserId
}

struct VideoDetailResp{
    1:common.BaseResp BaseResp
    2:common.Video video
}

struct BatchVideoInfoReq{
    1:i64 currentUserId
    2:list<i64> videoIds
}

struct BatchVideoInfoResp{
    1:common.BaseResp BaseResp
    2:map<i64,common.Video> videos
}

struct DeleteVideoReq{
    1:i64 videoId
    2:i64 userId
}

struct DeleteVideoResp{
    1:common.BaseResp BaseResp
}

struct UpdateVideoInfoReq{
    1:i64 videoId
    2:i64 userId
    3:optional string title
    4:optional string description
}

struct UpdateVideoInfoResp{
    1:common.BaseResp BaseResp
}

struct VideoStatsReq{
    1:i64 videoId
}

struct VideoStats{
    1:i64 videoId
    2:i64 viewCount
    3:i64 likeCount
    4:i64 commentCount
    5:i64 shareCount
}

struct VideoStatsResp{
    1:common.BaseResp BaseResp
    2:VideoStats stats
}

struct HotVideoReq{
    1:i64 userId
    2:i32 pageSize
}

struct HotVideoResp{
    1:common.BaseResp BaseResp
    2:list<common.Video> videos
}

service VideoService{
    PublishVideoResp PublishVideo(1:PublishVideoReq req)
    UserVideoListResp GetUserVideoList(1:UserVideoListReq req)
    FeedResp GetFeed(1:FeedReq req)
    SearchVideoResp SearchVideo(1:SearchVideoReq req)
    VideoDetailResp GetVideoDetail(1:VideoDetailReq req)
    BatchVideoInfoResp BatchGetVideoInfo(1:BatchVideoInfoReq req)
    DeleteVideoResp DeleteVideo(1:DeleteVideoReq req)
    UpdateVideoInfoResp UpdateVideoInfo(1:UpdateVideoInfoReq req)
    VideoStatsResp GetVideoStats(1:VideoStatsReq req)
    HotVideoResp GetHotVideos(1:HotVideoReq req)
}