namespace go common

struct BaseResp{
    1:i32 statusCode
    2:optional string msg
}

struct User{
    1:i64 id
    2:string username
    3:string password
    4:optional string avatar
    5:optional string about
    6:i64 followCount
    7:i64 followerCount
    8:bool isFollow
}

struct Video{
    1:i64 id
    2:i64 authorId
    3:string url
    4:string coverUrl
    5:i64 likeCount
    6:i64 commentCount
    7:bool isLike
    8:string title
    9:i64 publishTime
    10:string description
}

struct Comment{
    1:i64 id
    2:i64 userId
    3:i64 videoId
    4:string content
    5:string createTime
    6:i64 replyToId
}

struct Message{
    1:i64 id
    2:i64 receiveId
    3:i64 sendId
    4:string content
    5:string createTime
    6:bool isRead
}

struct LiveRoom{
    1:i64 id
    2:i64 hostId
    3:string title
    4:string coverUrl
    5:string rtmpUrl
    6:string hlsUrl
    7:i64 viewerCount
    8:bool isLive
    9:string createTime
}

struct Danmu{
    1:i64 id
    2:i64 userId
    3:i64 liveId
    4:string content
    5:string color
    6:string createTime
}