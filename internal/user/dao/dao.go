package dao

import (
	"context"
	"errors"
	"shortvideo/internal/user/model"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id int64) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id int64) error
	ListByIDs(ctx context.Context, ids []int64) ([]*model.User, error)
	Search(ctx context.Context, keyword string, page, pageSize int) ([]*model.User, int64, error)
	Count(ctx context.Context) (int64, error)
	BatchCheckUsername(ctx context.Context, usernames []string) (map[string]bool, error)
	BatchGetByIDs(ctx context.Context, ids []int64) (map[int64]*model.User, error)
	UpdateFollowCount(ctx context.Context, userID int64, delta int64) error
	UpdateFollowerCount(ctx context.Context, userID int64, delta int64) error
	WithTransaction(ctx context.Context, fn func(txRepo UserRepository) error) error
}

type userRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{db: db}
}

func (r *userRepositoryImpl) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepositoryImpl) FindByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *userRepositoryImpl) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *userRepositoryImpl) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepositoryImpl) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

func (r *userRepositoryImpl) ListByIDs(ctx context.Context, ids []int64) ([]*model.User, error) {
	var users []*model.User
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error
	return users, err
}

func (r *userRepositoryImpl) Search(ctx context.Context, keyword string, page, pageSize int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64
	offset := (page - 1) * pageSize

	db := r.db.WithContext(ctx)
	if keyword != "" {
		db = db.Where("username LIKE ?", "%"+keyword+"%")
	}

	if err := db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error
	return users, total, err
}

func (r *userRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Count(&count).Error
	return count, err
}

func (r *userRepositoryImpl) BatchCheckUsername(ctx context.Context, usernames []string) (map[string]bool, error) {
	var users []model.User
	result := make(map[string]bool)

	for _, username := range usernames {
		result[username] = false
	}

	err := r.db.WithContext(ctx).Select("username").Where("username IN ?", usernames).Find(&users).Error
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		result[user.Username] = true
	}

	return result, nil
}

func (r *userRepositoryImpl) BatchGetByIDs(ctx context.Context, ids []int64) (map[int64]*model.User, error) {
	var users []*model.User
	result := make(map[int64]*model.User)

	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		result[user.ID] = user
	}

	return result, err
}

func (r *userRepositoryImpl) UpdateFollowCount(ctx context.Context, userID int64, delta int64) error {
	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", userID).
		UpdateColumn("follow_count", gorm.Expr("follow_count + ?", delta)).Error
}

func (r *userRepositoryImpl) UpdateFollowerCount(ctx context.Context, userID int64, delta int64) error {
	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", userID).
		UpdateColumn("follower_count", gorm.Expr("follower_count + ?", delta)).Error
}

func (r *userRepositoryImpl) WithTransaction(ctx context.Context, fn func(txRepo UserRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &userRepositoryImpl{db: tx}
		return fn(txRepo)
	})
}
