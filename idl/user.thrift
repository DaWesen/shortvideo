namespace go user

include "common.thrift"

struct RegisterReq{
    1:string username
    2:string password
    3:optional string avatar
    4:optional string about
}

struct LoginReq{
    1:string username
    2:string password
}

struct LoginRegisterResp{
    1:common.BaseResp BaseResp
    2:common.User user
    3:string token
}

struct UserInfoReq{
    1:i64 userId
    2:i64 currentUserId
}

struct UserInfoResp{
    1:common.BaseResp BaseResp
    2:common.User user
}

struct BatchUserInfoReq{
    1:i64 currentUserId
    2:list<i64> userIds
}

struct BatchUserInfoResp{
    1:common.BaseResp BaseResp
    2:map<i64,common.User> users
}

struct UpdateUserReq{
    1:i64 userId
    2:optional string avatar
    3:optional string about
    4:optional string oldPassword
    5:optional string newPassword
}

struct CheckUsernameReq{
    1:string username
}

struct CheckUsernameResp{
    1:bool available
    2:common.BaseResp BaseResp
}

struct UserStatsReq{
    1:i64 userId
}

struct UserStats{
    1:i64 userId
    2:i64 videoCount
    3:i64 totalLikesReceived
    4:i64 totalComments
}

struct UserStatsResp{
    1:common.BaseResp BaseResp
    2:UserStats stats
    3:i64 totalUserCount
}

struct UserInfoByUsernameReq{
    1:string username
}

struct UpdateAvatarReq{
    1:i64 userId
    2:binary avatarData
}

struct BatchCheckUsernamesReq{
    1:list<string> usernames
}

struct BatchCheckUsernamesResp{
    1:common.BaseResp BaseResp
    2:map<string,bool> availableMap
}

struct SearchUsersReq{
    1:string keyword
    2:i32 page
    3:i32 pageSize
}

struct SearchUsersResp{
    1:common.BaseResp BaseResp
    2:list<common.User> users
    3:i64 total
}

struct UpdateFollowCountReq{
    1:i64 userId
    2:i64 delta
}

struct UpdateFollowerCountReq{
    1:i64 userId
    2:i64 delta
}

service UserService{
    LoginRegisterResp Register(1:RegisterReq req)
    LoginRegisterResp Login(1:LoginReq req)
    UserInfoResp GetUserInfo(1:UserInfoReq req)
    UserInfoResp GetUserInfoByUsername(1:UserInfoByUsernameReq req)
    BatchUserInfoResp BatchGetUserInfo(1:BatchUserInfoReq req)
    common.BaseResp UpdateUser(1:UpdateUserReq req)
    common.BaseResp UpdateAvatar(1:UpdateAvatarReq req)
    CheckUsernameResp CheckUsername(1:CheckUsernameReq req)
    BatchCheckUsernamesResp BatchCheckUsernames(1:BatchCheckUsernamesReq req)
    UserStatsResp GetUserStats(1:UserStatsReq req)
    SearchUsersResp SearchUsers(1:SearchUsersReq req)
    common.BaseResp UpdateFollowCount(1:UpdateFollowCountReq req)
    common.BaseResp UpdateFollowerCount(1:UpdateFollowerCountReq req)
    bool VerifyToken(1:string token)
}