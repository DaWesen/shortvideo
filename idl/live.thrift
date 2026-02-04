namespace go live

include "common.thrift"

struct CreateLiveRoomReq{
    1:i64 hostId
    2:string title
    3:string coverUrl
    4:optional string description
}

struct CreateLiveRoomResp{
    1:common.BaseResp BaseResp
    2:common.LiveRoom room
}

struct StartLiveReq{
    1:i64 hostId
    2:i64 roomId
    3:string rtmpUrl
}

struct StartLiveResp{
    1:common.BaseResp BaseResp
}

struct StopLiveReq{
    1:i64 hostId
    2:i64 roomId
}

struct StopLiveResp{
    1:common.BaseResp BaseResp
}

struct GetLiveRoomsReq{
    1:i64 userId
    2:i32 page
    3:i32 pageSize
    4:optional bool followingOnly
}

struct GetLiveRoomsResp{
    1:common.BaseResp BaseResp
    2:list<common.LiveRoom> rooms
    3:i32 totalCount
}

struct GetLiveRoomDetailReq{
    1:i64 roomId
    2:i64 userId
}

struct GetLiveRoomDetailResp{
    1:common.BaseResp BaseResp
    2:common.LiveRoom room
    3:i64 onlineCount
}

struct JoinLiveRoomReq{
    1:i64 roomId
    2:i64 userId
}

struct JoinLiveRoomResp{
    1:common.BaseResp BaseResp
    2:string hlsUrl
    3:list<string> chatHistory
}

struct LeaveLiveRoomReq{
    1:i64 roomId
    2:i64 userId
}

struct LeaveLiveRoomResp{
    1:common.BaseResp BaseResp
}

struct Gift{
    1:i64 id
    2:string name
    3:i64 price
    4:string iconUrl
    5:string animationUrl
}

struct SendGiftReq{
    1:i64 senderId
    2:i64 roomId
    3:i64 giftId
    4:i32 count
}

struct SendGiftResp{
    1:common.BaseResp BaseResp
    2:i64 totalPrice
}

struct GetGiftListResp{
    1:common.BaseResp BaseResp
    2:list<Gift> gifts
}

struct SetRoomAdminReq{
    1:i64 hostId
    2:i64 roomId
    3:i64 targetUserId
    4:bool action
}

struct SetRoomAdminResp{
    1:common.BaseResp BaseResp
}

struct RecordLiveReq{
    1:i64 hostId
    2:i64 roomId
    3:bool action
}

struct RecordLiveResp{
    1:common.BaseResp BaseResp
    2:optional string videoUrl
}

service LiveService{
    CreateLiveRoomResp CreateLiveRoom(1:CreateLiveRoomReq req)
    StartLiveResp StartLive(1:StartLiveReq req)
    StopLiveResp StopLive(1:StopLiveReq req)
    GetLiveRoomsResp GetLiveRooms(1:GetLiveRoomsReq req)
    GetLiveRoomDetailResp GetLiveRoomDetail(1:GetLiveRoomDetailReq req)
    JoinLiveRoomResp JoinLiveRoom(1:JoinLiveRoomReq req)
    LeaveLiveRoomResp LeaveLiveRoom(1:LeaveLiveRoomReq req)
    SendGiftResp SendGift(1:SendGiftReq req)
    GetGiftListResp GetGiftList()
    SetRoomAdminResp SetRoomAdmin(1:SetRoomAdminReq req)
    RecordLiveResp RecordLive(1:RecordLiveReq req)
}