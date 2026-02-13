package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"shortvideo/internal/user/dao"
	"shortvideo/internal/user/model"
	"shortvideo/pkg/cache"
	"shortvideo/pkg/es"
	"shortvideo/pkg/jwt"
	"shortvideo/pkg/logger"
	"shortvideo/pkg/mq"
	"shortvideo/pkg/storage"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound     = errors.New("用户不存在")
	ErrUsernameExists   = errors.New("用户名已存在")
	ErrInvalidPassword  = errors.New("密码错误")
	ErrOldPasswordWrong = errors.New("旧密码错误")
	ErrTokenInvalid     = errors.New("令牌无效")
	ErrInternalServer   = errors.New("服务器内部错误")
	ErrFileUploadFailed = errors.New("文件上传失败")
	ErrInvalidFile      = errors.New("无效的文件")
)

type UserService interface {
	//注册相关
	Register(ctx context.Context, username, password, avatar, about string) (*model.User, string, error)
	Login(ctx context.Context, username, password string) (*model.User, string, error)

	//用户信息相关
	GetUserByID(ctx context.Context, id int64) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	BatchGetUsersByIDs(ctx context.Context, ids []int64) (map[int64]*model.User, error)
	UpdateUser(ctx context.Context, userID int64, avatar, about, oldPassword, newPassword string) error
	UpdateAvatar(ctx context.Context, userID int64, avatarData []byte) (string, error)

	//用户名检查
	CheckUsernameAvailable(ctx context.Context, username string) (bool, error)
	BatchCheckUsernames(ctx context.Context, usernames []string) (map[string]bool, error)

	//统计相关
	GetUserCount(ctx context.Context) (int64, error)
	SearchUsers(ctx context.Context, keyword string, page, pageSize int) ([]*model.User, int64, error)

	//关注数相关
	UpdateFollowCount(ctx context.Context, userID int64, delta int64) error
	UpdateFollowerCount(ctx context.Context, userID int64, delta int64) error

	//Token相关
	VerifyToken(ctx context.Context, token string) (int64, error)
	GenerateToken(userID int64) (string, error)

	//事务相关
	WithTransaction(ctx context.Context, fn func(txService UserService) error) error
}

type userServiceImpl struct {
	repo          dao.UserRepository
	jwtManager    *jwt.JWTManager
	storage       storage.Storage
	kafkaProducer *mq.Producer
	cache         cache.Cache
	es            *es.ESManager
}

func NewUserService(repo dao.UserRepository, jwtManager *jwt.JWTManager, storage storage.Storage,
	kafkaProducer *mq.Producer, cache cache.Cache, es *es.ESManager) UserService {
	return &userServiceImpl{
		repo:          repo,
		jwtManager:    jwtManager,
		storage:       storage,
		kafkaProducer: kafkaProducer,
		cache:         cache,
		es:            es,
	}
}

func NewUserServiceWithRepo(repo dao.UserRepository, jwtManager *jwt.JWTManager) UserService {
	return &userServiceImpl{
		repo:          repo,
		jwtManager:    jwtManager,
		storage:       nil,
		kafkaProducer: nil,
		cache:         nil,
		es:            nil,
	}
}

// 注册新用户
func (s *userServiceImpl) Register(ctx context.Context, username, password, avatar, about string) (*model.User, string, error) {
	logger.Info("用户注册请求",
		logger.StringField("username", username),
		logger.StringField("about", about))

	existingUser, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		logger.Error("查询用户失败",
			logger.ErrorField(err),
			logger.StringField("username", username))
		return nil, "", ErrInternalServer
	}
	if existingUser != nil {
		logger.Warn("用户名已存在",
			logger.StringField("username", username))
		return nil, "", ErrUsernameExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("密码加密失败",
			logger.ErrorField(err),
			logger.StringField("username", username))
		return nil, "", ErrInternalServer
	}
	user := &model.User{
		Username: username,
		Password: string(hashedPassword),
		Avatar:   avatar,
		About:    about,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		logger.Error("创建用户失败",
			logger.ErrorField(err),
			logger.StringField("username", username))
		return nil, "", ErrInternalServer
	}

	token, err := s.jwtManager.GenerateToken(user.ID)
	if err != nil {
		logger.Error("生成令牌失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", user.ID))
		return nil, "", ErrInternalServer
	}

	if s.kafkaProducer != nil {
		eventData, _ := json.Marshal(map[string]interface{}{
			"user_id":       user.ID,
			"username":      user.Username,
			"registered_at": time.Now(),
			"avatar":        user.Avatar,
			"about":         user.About,
		})
		s.kafkaProducer.SendUserEvent(ctx, fmt.Sprintf("%d", user.ID), eventData)
		logger.Info("发送用户注册事件",
			logger.Int64Field("user_id", user.ID))
	}

	//将用户信息同步到Elasticsearch
	if s.es != nil {
		esUser := map[string]interface{}{
			"id":             user.ID,
			"username":       user.Username,
			"avatar":         user.Avatar,
			"about":          user.About,
			"follow_count":   user.FollowCount,
			"follower_count": user.FollowerCount,
			"created_at":     user.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		err = s.es.AddDocument("users", fmt.Sprintf("%d", user.ID), esUser)
		if err != nil {
			logger.Error("同步用户到ES失败",
				logger.ErrorField(err),
				logger.Int64Field("user_id", user.ID))
		}
	}

	logger.Info("用户注册成功",
		logger.Int64Field("user_id", user.ID),
		logger.StringField("username", username))

	return user, token, nil
}

// 用户登录
func (s *userServiceImpl) Login(ctx context.Context, username, password string) (*model.User, string, error) {
	logger.Info("用户登录请求",
		logger.StringField("username", username))

	user, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		logger.Error("查询用户失败",
			logger.ErrorField(err),
			logger.StringField("username", username))
		return nil, "", ErrInternalServer
	}
	if user == nil {
		logger.Warn("用户不存在",
			logger.StringField("username", username))
		return nil, "", ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		logger.Warn("密码错误",
			logger.StringField("username", username))
		return nil, "", ErrInvalidPassword
	}

	token, err := s.jwtManager.GenerateToken(user.ID)
	if err != nil {
		logger.Error("生成令牌失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", user.ID))
		return nil, "", ErrInternalServer
	}

	if s.kafkaProducer != nil {
		eventData, _ := json.Marshal(map[string]interface{}{
			"user_id":    user.ID,
			"username":   user.Username,
			"logged_at":  time.Now(),
			"ip_address": "",
		})
		s.kafkaProducer.SendUserEvent(ctx, fmt.Sprintf("%d", user.ID), eventData)
		logger.Info("发送用户登录事件",
			logger.Int64Field("user_id", user.ID))
	}

	logger.Info("用户登录成功",
		logger.Int64Field("user_id", user.ID),
		logger.StringField("username", username))

	return user, token, nil
}

// 根据ID获取用户
func (s *userServiceImpl) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	logger.Info("获取用户信息请求",
		logger.Int64Field("user_id", id))

	if s.cache != nil {
		userKey := cache.GenerateUserKey(id)
		cachedUser, err := s.cache.Get(ctx, userKey)
		if err == nil && cachedUser != "" {
			var user model.User
			if err := json.Unmarshal([]byte(cachedUser), &user); err == nil {
				logger.Info("从缓存获取用户信息成功",
					logger.Int64Field("user_id", id))
				return &user, nil
			}
		}
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		logger.Error("查询用户失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", id))
		return nil, ErrInternalServer
	}
	if user == nil {
		logger.Warn("用户不存在",
			logger.Int64Field("user_id", id))
		return nil, ErrUserNotFound
	}

	if s.cache != nil {
		userKey := cache.GenerateUserKey(id)
		userData, err := json.Marshal(user)
		if err == nil {
			s.cache.Set(ctx, userKey, string(userData), 10*time.Minute)
			logger.Info("用户信息存入缓存成功",
				logger.Int64Field("user_id", id))
		}
	}

	logger.Info("获取用户信息成功",
		logger.Int64Field("user_id", id),
		logger.StringField("username", user.Username))

	return user, nil
}

// 根据用户名获取用户
func (s *userServiceImpl) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	user, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, ErrInternalServer
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// 批量根据ID获取用户
func (s *userServiceImpl) BatchGetUsersByIDs(ctx context.Context, ids []int64) (map[int64]*model.User, error) {
	users, err := s.repo.BatchGetByIDs(ctx, ids)
	if err != nil {
		return nil, ErrInternalServer
	}
	return users, nil
}

// 更新用户信息
func (s *userServiceImpl) UpdateUser(ctx context.Context, userID int64, avatar, about, oldPassword, newPassword string) error {
	logger.Info("更新用户信息请求",
		logger.Int64Field("user_id", userID))

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		logger.Error("查询用户失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID))
		return ErrInternalServer
	}
	if user == nil {
		logger.Warn("用户不存在",
			logger.Int64Field("user_id", userID))
		return ErrUserNotFound
	}

	if avatar != "" {
		user.Avatar = avatar
	}
	if about != "" {
		user.About = about
	}

	if oldPassword != "" && newPassword != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
			logger.Warn("旧密码错误",
				logger.Int64Field("user_id", userID))
			return ErrOldPasswordWrong
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			logger.Error("密码加密失败",
				logger.ErrorField(err),
				logger.Int64Field("user_id", userID))
			return ErrInternalServer
		}
		user.Password = string(hashedPassword)
	}

	if err := s.repo.Update(ctx, user); err != nil {
		logger.Error("更新用户失败",
			logger.ErrorField(err),
			logger.Int64Field("user_id", userID))
		return ErrInternalServer
	}

	if s.cache != nil {
		userKey := cache.GenerateUserKey(userID)
		s.cache.Delete(ctx, userKey)
		logger.Info("删除用户缓存成功",
			logger.Int64Field("user_id", userID))
	}

	//将更新后的用户信息同步到Elasticsearch
	if s.es != nil {
		updatedUser, err := s.repo.FindByID(ctx, userID)
		if err == nil && updatedUser != nil {
			esUser := map[string]interface{}{
				"id":             updatedUser.ID,
				"username":       updatedUser.Username,
				"avatar":         updatedUser.Avatar,
				"about":          updatedUser.About,
				"follow_count":   updatedUser.FollowCount,
				"follower_count": updatedUser.FollowerCount,
				"created_at":     updatedUser.CreatedAt.Format("2006-01-02 15:04:05"),
			}
			err = s.es.UpdateDocument("users", fmt.Sprintf("%d", userID), esUser)
			if err != nil {
				logger.Error("更新用户到ES失败",
					logger.ErrorField(err),
					logger.Int64Field("user_id", userID))
			}
		}
	}

	logger.Info("更新用户信息成功",
		logger.Int64Field("user_id", userID))

	return nil
}

// 检查用户名是否可用
func (s *userServiceImpl) CheckUsernameAvailable(ctx context.Context, username string) (bool, error) {
	user, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return false, ErrInternalServer
	}
	return user == nil, nil
}

// 批量检查用户名
func (s *userServiceImpl) BatchCheckUsernames(ctx context.Context, usernames []string) (map[string]bool, error) {
	result, err := s.repo.BatchCheckUsername(ctx, usernames)
	if err != nil {
		return nil, ErrInternalServer
	}

	availableMap := make(map[string]bool)
	for username, exists := range result {
		availableMap[username] = !exists
	}

	return availableMap, nil
}

// 获取用户总数
func (s *userServiceImpl) GetUserCount(ctx context.Context) (int64, error) {
	count, err := s.repo.Count(ctx)
	if err != nil {
		return 0, ErrInternalServer
	}
	return count, nil
}

// 搜索用户
func (s *userServiceImpl) SearchUsers(ctx context.Context, keyword string, page, pageSize int) ([]*model.User, int64, error) {
	//优先使用Elasticsearch进行搜索
	if s.es != nil {
		//构建搜索查询
		query := es.SearchQuery{
			Query: map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":    keyword,
					"fields":   []string{"username", "about"},
					"type":     "best_fields",
					"operator": "and",
				},
			},
			From: (page - 1) * pageSize,
			Size: pageSize,
			Sort: []map[string]interface{}{
				{
					"follower_count": map[string]interface{}{
						"order": "desc",
					},
				},
				{
					"created_at": map[string]interface{}{
						"order": "desc",
					},
				},
			},
		}

		var searchResult es.SearchResult
		err := s.es.Search("users", query, &searchResult)
		if err == nil {
			//解析搜索结果
			var users []*model.User
			for _, hit := range searchResult.Hits.Hits {
				var esUser map[string]interface{}
				if json.Unmarshal(hit, &esUser) == nil {
					source, ok := esUser["_source"].(map[string]interface{})
					if ok {
						user := &model.User{}
						//从ES结果中提取用户信息
						if id, ok := source["id"].(float64); ok {
							user.ID = int64(id)
						}
						if username, ok := source["username"].(string); ok {
							user.Username = username
						}
						if avatar, ok := source["avatar"].(string); ok {
							user.Avatar = avatar
						}
						if about, ok := source["about"].(string); ok {
							user.About = about
						}
						if followCount, ok := source["follow_count"].(float64); ok {
							user.FollowCount = int64(followCount)
						}
						if followerCount, ok := source["follower_count"].(float64); ok {
							user.FollowerCount = int64(followerCount)
						}
						if createdAt, ok := source["created_at"].(string); ok {
							if t, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
								user.CreatedAt = t
							}
						}
						users = append(users, user)
					}
				}
			}
			return users, searchResult.Hits.Total.Value, nil
		}
	}

	//如果ES搜索失败，回退到数据库搜索
	users, total, err := s.repo.Search(ctx, keyword, page, pageSize)
	if err != nil {
		return nil, 0, ErrInternalServer
	}
	return users, total, nil
}

// 更新关注数
func (s *userServiceImpl) UpdateFollowCount(ctx context.Context, userID int64, delta int64) error {
	if err := s.repo.UpdateFollowCount(ctx, userID, delta); err != nil {
		return ErrInternalServer
	}
	return nil
}

// 更新粉丝数
func (s *userServiceImpl) UpdateFollowerCount(ctx context.Context, userID int64, delta int64) error {
	if err := s.repo.UpdateFollowerCount(ctx, userID, delta); err != nil {
		return ErrInternalServer
	}
	return nil
}

// 验证token
func (s *userServiceImpl) VerifyToken(ctx context.Context, token string) (int64, error) {
	userID, err := s.jwtManager.GetUserIDFromToken(token)
	if err != nil {
		return 0, ErrTokenInvalid
	}
	return userID, nil
}

// 生成token
func (s *userServiceImpl) GenerateToken(userID int64) (string, error) {
	token, err := s.jwtManager.GenerateToken(userID)
	if err != nil {
		return "", ErrInternalServer
	}
	return token, nil
}

// 事务支持
func (s *userServiceImpl) WithTransaction(ctx context.Context, fn func(txService UserService) error) error {
	return s.repo.WithTransaction(ctx, func(txRepo dao.UserRepository) error {
		txService := NewUserServiceWithRepo(txRepo, s.jwtManager)
		return fn(txService)
	})
}

// 更新用户头像
func (s *userServiceImpl) UpdateAvatar(ctx context.Context, userID int64, avatarData []byte) (string, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return "", ErrInternalServer
	}
	if user == nil {
		return "", ErrUserNotFound
	}

	if len(avatarData) == 0 {
		return "", ErrInvalidFile
	}

	objectName := fmt.Sprintf("avatars/%d_%d.jpg", userID, time.Now().Unix())

	if s.storage == nil {
		return "", ErrFileUploadFailed
	}

	reader := bytes.NewReader(avatarData)
	avatarURL, err := s.storage.Upload(ctx, "", objectName, reader, int64(len(avatarData)), "image/jpeg")
	if err != nil {
		return "", ErrFileUploadFailed
	}

	user.Avatar = avatarURL
	err = s.repo.Update(ctx, user)
	if err != nil {
		return "", ErrInternalServer
	}

	if s.kafkaProducer != nil {
		eventData, _ := json.Marshal(map[string]interface{}{
			"user_id":    userID,
			"avatar_url": avatarURL,
			"updated_at": time.Now(),
		})
		s.kafkaProducer.SendUserEvent(ctx, fmt.Sprintf("%d", userID), eventData)
	}

	return avatarURL, nil
}
