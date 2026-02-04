namespace go social

include "common.thrift"

struct FollowActionReq{
    1:i64 userId
    2:i64 targetUserId
    3:bool action
}

struct FollowActionResp{
    1:common.BaseResp BaseResp
}

struct FollowListReq{
    1:i64 userId
    2:i64 currentUserId
    3:i32 page
    4:i32 pageSize
}

struct FollowListResp{
    1:common.BaseResp BaseResp
    2:list<common.User> users
    3:i32 totalCount
}

struct FollowerListReq{
    1:i64 userId
    2:i64 currentUserId
    3:i32 page
    4:i32 pageSize
}

struct FollowerListResp{
    1:common.BaseResp BaseResp
    2:list<common.User> users
    3:i32 totalCount
}

struct FriendListReq{
    1:i64 userId
    2:i32 page
    3:i32 pageSize
}

struct FriendListResp{
    1:common.BaseResp BaseResp
    2:list<common.User> users
    3:i32 totalCount
}

struct CheckFollowReq{
    1:i64 userId
    2:i64 targetUserId
}

struct CheckFollowResp{
    1:common.BaseResp BaseResp
    2:bool isFollowing
}

struct CheckMutualFollowReq{
    1:i64 userId1
    2:i64 userId2
}

struct CheckMutualFollowResp{
    1:common.BaseResp BaseResp
    2:bool isMutualFollow
}

struct FollowStatsReq{
    1:i64 userId
}

struct FollowStats{
    1:i64 userId
    2:i64 followCount
    3:i64 followerCount
    4:i64 friendCount
}

struct FollowStatsResp{
    1:common.BaseResp BaseResp
    2:FollowStats stats
}

struct BatchCheckFollowReq{
    1:i64 userId
    2:list<i64> targetUserIds
}

struct BatchCheckFollowResp{
    1:common.BaseResp BaseResp
    2:map<i64,bool> followStatus
}

service SocialService{
    FollowActionResp FollowAction(1:FollowActionReq req)
    FollowListResp GetFollowList(1:FollowListReq req)
    FollowerListResp GetFollowerList(1:FollowerListReq req)
    FriendListResp GetFriendList(1:FriendListReq req)
    CheckFollowResp CheckFollow(1:CheckFollowReq req)
    CheckMutualFollowResp CheckMutualFollow(1:CheckMutualFollowReq req)
    FollowStatsResp GetFollowStats(1:FollowStatsReq req)
    BatchCheckFollowResp BatchCheckFollow(1:BatchCheckFollowReq req)
}