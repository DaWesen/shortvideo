namespace go danmu

include "common.thrift"

struct SendDanmuReq{
    1:i64 userId
    2:i64 liveId
    3:string content
    4:optional string color
    5:optional i32 fontSize
    6:optional i32 position
}

struct SendDanmuResp{
    1:common.BaseResp BaseResp
    2:i64 danmuId
}

struct GetDanmuHistoryReq{
    1:i64 liveId
    2:i64 startTime
    3:i64 endTime
    4:i32 limit
}

struct GetDanmuHistoryResp{
    1:common.BaseResp BaseResp
    2:list<common.Danmu> danmus
}

struct DanmuFilter{
    1:i64 userId
    2:i64 liveId
    3:list<string> keywords
    4:bool hideAnonymous
    5:bool hideLowLevel
}

struct SetDanmuFilterReq{
    1:i64 userId
    2:i64 liveId
    3:DanmuFilter filter
}

struct SetDanmuFilterResp{
    1:common.BaseResp BaseResp
}

struct GetDanmuFilterReq{
    1:i64 userId
    2:i64 liveId
}

struct GetDanmuFilterResp{
    1:common.BaseResp BaseResp
    2:DanmuFilter filter
}

struct DanmuStats{
    1:i64 liveId
    2:i64 totalDanmuCount
    3:i64 activeUserCount
    4:i64 peakDanmuPerMinute
    5:map<string,i64> wordCloud
}

struct GetDanmuStatsReq{
    1:i64 liveId
}

struct GetDanmuStatsResp{
    1:common.BaseResp BaseResp
    2:DanmuStats stats
}

struct ManageDanmuReq{
    1:i64 managerId
    2:i64 liveId
    3:i64 danmuId
    4:i32 action
}

struct ManageDanmuResp{
    1:common.BaseResp BaseResp
}

struct LikeDanmuReq{
    1:i64 userId
    2:i64 danmuId
}

struct LikeDanmuResp{
    1:common.BaseResp BaseResp
    2:i64 likeCount
}

service DanmuService{
    SendDanmuResp SendDanmu(1:SendDanmuReq req)
    GetDanmuHistoryResp GetDanmuHistory(1:GetDanmuHistoryReq req)
    SetDanmuFilterResp SetDanmuFilter(1:SetDanmuFilterReq req)
    GetDanmuFilterResp GetDanmuFilter(1:GetDanmuFilterReq req)
    GetDanmuStatsResp GetDanmuStats(1:GetDanmuStatsReq req)
    ManageDanmuResp ManageDanmu(1:ManageDanmuReq req)
    LikeDanmuResp LikeDanmu(1:LikeDanmuReq req)
}