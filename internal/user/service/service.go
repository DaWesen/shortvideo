package service

import (
	"context"
	"errors"
	"shortvideo/internal/user/dao"
	"shortvideo/internal/user/model"
	"shortvideo/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound     = errors.New("用户不存在")
	ErrUsernameExists   = errors.New("用户名已存在")
	ErrInvalidPassword  = errors.New("密码错误")
	ErrOldPasswordWrong = errors.New("旧密码错误")
	ErrTokenInvalid     = errors.New("令牌无效")
	ErrInternalServer   = errors.New("服务器内部错误")
)

type UserService interface {
	// 注册相关
	Register(ctx context.Context, username, password, avatar, about string) (*model.User, string, error)
	Login(ctx context.Context, username, password string) (*model.User, string, error)
	// 用户信息相关
	GetUserByID(ctx context.Context, id int64) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	BatchGetUsersByIDs(ctx context.Context, ids []int64) (map[int64]*model.User, error)
	UpdateUser(ctx context.Context, userID int64, avatar, about, oldPassword, newPassword string) error
	// 用户名检查
	CheckUsernameAvailable(ctx context.Context, username string) (bool, error)
	BatchCheckUsernames(ctx context.Context, usernames []string) (map[string]bool, error)
	// 统计相关
	GetUserCount(ctx context.Context) (int64, error)
	SearchUsers(ctx context.Context, keyword string, page, pageSize int) ([]*model.User, int64, error)
	// 关注数相关
	UpdateFollowCount(ctx context.Context, userID int64, delta int64) error
	UpdateFollowerCount(ctx context.Context, userID int64, delta int64) error
	// Token相关
	VerifyToken(ctx context.Context, token string) (int64, error)
	GenerateToken(userID int64) (string, error)
	// 事务相关
	WithTransaction(ctx context.Context, fn func(txService UserService) error) error
}

type userServiceImpl struct {
	repo       dao.UserRepository
	jwtManager *jwt.JWTManager
}

func NewUserService(repo dao.UserRepository, jwtManager *jwt.JWTManager) UserService {
	return &userServiceImpl{
		repo:       repo,
		jwtManager: jwtManager,
	}
}

// 注册新用户
func (s *userServiceImpl) Register(ctx context.Context, username, password, avatar, about string) (*model.User, string, error) {
	existingUser, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, "", ErrInternalServer
	}
	if existingUser != nil {
		return nil, "", ErrUsernameExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", ErrInternalServer
	}
	user := &model.User{
		Username: username,
		Password: string(hashedPassword),
		Avatar:   avatar,
		About:    about,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, "", ErrInternalServer
	}

	token, err := s.jwtManager.GenerateToken(user.ID)
	if err != nil {
		return nil, "", ErrInternalServer
	}

	return user, token, nil
}

// 用户登录
func (s *userServiceImpl) Login(ctx context.Context, username, password string) (*model.User, string, error) {
	user, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, "", ErrInternalServer
	}
	if user == nil {
		return nil, "", ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", ErrInvalidPassword
	}

	token, err := s.jwtManager.GenerateToken(user.ID)
	if err != nil {
		return nil, "", ErrInternalServer
	}

	return user, token, nil
}

// 根据ID获取用户
func (s *userServiceImpl) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrInternalServer
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
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
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return ErrInternalServer
	}
	if user == nil {
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
			return ErrOldPasswordWrong
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return ErrInternalServer
		}
		user.Password = string(hashedPassword)
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return ErrInternalServer
	}

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
		txService := NewUserService(txRepo, s.jwtManager)
		return fn(txService)
	})
}
