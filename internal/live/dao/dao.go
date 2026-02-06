package dao

import (
	"context"
	"errors"
	"shortvideo/internal/live/model"

	"gorm.io/gorm"
)

type LiveRoomRepository interface {
	Create(ctx context.Context, room *model.LiveRoom) error
	Update(ctx context.Context, room *model.LiveRoom) error
	FindByID(ctx context.Context, id int64) (*model.LiveRoom, error)
	FindByHostID(ctx context.Context, hostID int64) (*model.LiveRoom, error)
	Delete(ctx context.Context, id int64, hostID int64) error
	ListLiveRooms(ctx context.Context, page, pageSize int, followingOnly bool, userID int64) ([]*model.LiveRoom, int64, error)
	UpdateViewerCount(ctx context.Context, roomID int64, delta int64) error
	UpdateLiveStatus(ctx context.Context, roomID int64, isLive bool) error
	UpdateStreamURLs(ctx context.Context, roomID int64, rtmpURL, hlsURL string) error
	WithTransaction(ctx context.Context, fn func(txRepo LiveRoomRepository) error) error
}

type GiftRepository interface {
	Create(ctx context.Context, gift *model.Gift) error
	Update(ctx context.Context, gift *model.Gift) error
	FindByID(ctx context.Context, id int64) (*model.Gift, error)
	ListAll(ctx context.Context) ([]*model.Gift, error)
	Delete(ctx context.Context, id int64) error
	WithTransaction(ctx context.Context, fn func(txRepo GiftRepository) error) error
}

type GiftRecordRepository interface {
	Create(ctx context.Context, record *model.GiftRecord) error
	FindByID(ctx context.Context, id int64) (*model.GiftRecord, error)
	ListByRoomID(ctx context.Context, roomID int64, page, pageSize int) ([]*model.GiftRecord, int64, error)
	ListBySenderID(ctx context.Context, senderID int64, page, pageSize int) ([]*model.GiftRecord, int64, error)
	GetTotalGiftValueByRoom(ctx context.Context, roomID int64) (int64, error)
	GetTotalGiftValueBySender(ctx context.Context, senderID int64) (int64, error)
	WithTransaction(ctx context.Context, fn func(txRepo GiftRecordRepository) error) error
}

type RoomAdminRepository interface {
	Create(ctx context.Context, admin *model.RoomAdmin) error
	Delete(ctx context.Context, roomID, userID int64) error
	Find(ctx context.Context, roomID, userID int64) (*model.RoomAdmin, error)
	ListByRoomID(ctx context.Context, roomID int64) ([]*model.RoomAdmin, error)
	ListByUserID(ctx context.Context, userID int64) ([]*model.RoomAdmin, error)
	IsAdmin(ctx context.Context, roomID, userID int64) (bool, error)
	WithTransaction(ctx context.Context, fn func(txRepo RoomAdminRepository) error) error
}

type LiveRecordRepository interface {
	Create(ctx context.Context, record *model.LiveRecord) error
	Update(ctx context.Context, record *model.LiveRecord) error
	FindByID(ctx context.Context, id int64) (*model.LiveRecord, error)
	FindByRoomID(ctx context.Context, roomID int64) ([]*model.LiveRecord, error)
	Delete(ctx context.Context, id int64) error
	WithTransaction(ctx context.Context, fn func(txRepo LiveRecordRepository) error) error
}

type RoomViewerRepository interface {
	CreateOrUpdate(ctx context.Context, viewer *model.RoomViewer) error
	Find(ctx context.Context, roomID, userID int64) (*model.RoomViewer, error)
	ListByRoomID(ctx context.Context, roomID int64, page, pageSize int) ([]*model.RoomViewer, int64, error)
	CountByRoomID(ctx context.Context, roomID int64) (int64, error)
	UpdateLeaveTime(ctx context.Context, roomID, userID int64, leaveTime string) error
	Delete(ctx context.Context, roomID, userID int64) error
	WithTransaction(ctx context.Context, fn func(txRepo RoomViewerRepository) error) error
}

type liveRoomRepositoryImpl struct {
	db *gorm.DB
}

func NewLiveRoomRepository(db *gorm.DB) LiveRoomRepository {
	return &liveRoomRepositoryImpl{db: db}
}

func (r *liveRoomRepositoryImpl) Create(ctx context.Context, room *model.LiveRoom) error {
	return r.db.WithContext(ctx).Create(room).Error
}

func (r *liveRoomRepositoryImpl) Update(ctx context.Context, room *model.LiveRoom) error {
	return r.db.WithContext(ctx).Save(room).Error
}

func (r *liveRoomRepositoryImpl) FindByID(ctx context.Context, id int64) (*model.LiveRoom, error) {
	var room model.LiveRoom
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&room).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &room, err
}

func (r *liveRoomRepositoryImpl) FindByHostID(ctx context.Context, hostID int64) (*model.LiveRoom, error) {
	var room model.LiveRoom
	err := r.db.WithContext(ctx).Where("host_id = ?", hostID).First(&room).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &room, err
}

func (r *liveRoomRepositoryImpl) Delete(ctx context.Context, id int64, hostID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND host_id = ?", id, hostID).
		Delete(&model.LiveRoom{}).Error
}

func (r *liveRoomRepositoryImpl) ListLiveRooms(ctx context.Context, page, pageSize int, followingOnly bool, userID int64) ([]*model.LiveRoom, int64, error) {
	var rooms []*model.LiveRoom
	var total int64
	offset := (page - 1) * pageSize

	query := r.db.WithContext(ctx).Model(&model.LiveRoom{})

	if followingOnly && userID > 0 {
	}

	query = query.Where("is_live = ?", true)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(pageSize).
		Order("viewer_count DESC").
		Find(&rooms).Error

	return rooms, total, err
}

func (r *liveRoomRepositoryImpl) UpdateViewerCount(ctx context.Context, roomID int64, delta int64) error {
	return r.db.WithContext(ctx).Model(&model.LiveRoom{}).
		Where("id = ?", roomID).
		UpdateColumn("viewer_count", gorm.Expr("viewer_count + ?", delta)).Error
}

func (r *liveRoomRepositoryImpl) UpdateLiveStatus(ctx context.Context, roomID int64, isLive bool) error {
	return r.db.WithContext(ctx).Model(&model.LiveRoom{}).
		Where("id = ?", roomID).
		Update("is_live", isLive).Error
}

func (r *liveRoomRepositoryImpl) UpdateStreamURLs(ctx context.Context, roomID int64, rtmpURL, hlsURL string) error {
	updates := make(map[string]interface{})
	if rtmpURL != "" {
		updates["rtmp_url"] = rtmpURL
	}
	if hlsURL != "" {
		updates["hls_url"] = hlsURL
	}

	return r.db.WithContext(ctx).Model(&model.LiveRoom{}).
		Where("id = ?", roomID).
		Updates(updates).Error
}

func (r *liveRoomRepositoryImpl) WithTransaction(ctx context.Context, fn func(txRepo LiveRoomRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &liveRoomRepositoryImpl{db: tx}
		return fn(txRepo)
	})
}

type giftRepositoryImpl struct {
	db *gorm.DB
}

func NewGiftRepository(db *gorm.DB) GiftRepository {
	return &giftRepositoryImpl{db: db}
}

func (r *giftRepositoryImpl) Create(ctx context.Context, gift *model.Gift) error {
	return r.db.WithContext(ctx).Create(gift).Error
}

func (r *giftRepositoryImpl) Update(ctx context.Context, gift *model.Gift) error {
	return r.db.WithContext(ctx).Save(gift).Error
}

func (r *giftRepositoryImpl) FindByID(ctx context.Context, id int64) (*model.Gift, error) {
	var gift model.Gift
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&gift).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &gift, err
}

func (r *giftRepositoryImpl) ListAll(ctx context.Context) ([]*model.Gift, error) {
	var gifts []*model.Gift
	err := r.db.WithContext(ctx).Order("price ASC").Find(&gifts).Error
	return gifts, err
}

func (r *giftRepositoryImpl) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Gift{}, id).Error
}

func (r *giftRepositoryImpl) WithTransaction(ctx context.Context, fn func(txRepo GiftRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &giftRepositoryImpl{db: tx}
		return fn(txRepo)
	})
}

type giftRecordRepositoryImpl struct {
	db *gorm.DB
}

func NewGiftRecordRepository(db *gorm.DB) GiftRecordRepository {
	return &giftRecordRepositoryImpl{db: db}
}

func (r *giftRecordRepositoryImpl) Create(ctx context.Context, record *model.GiftRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *giftRecordRepositoryImpl) FindByID(ctx context.Context, id int64) (*model.GiftRecord, error) {
	var record model.GiftRecord
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &record, err
}

func (r *giftRecordRepositoryImpl) ListByRoomID(ctx context.Context, roomID int64, page, pageSize int) ([]*model.GiftRecord, int64, error) {
	var records []*model.GiftRecord
	var total int64
	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).Model(&model.GiftRecord{}).
		Where("room_id = ?", roomID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).Where("room_id = ?", roomID).
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&records).Error

	return records, total, err
}

func (r *giftRecordRepositoryImpl) ListBySenderID(ctx context.Context, senderID int64, page, pageSize int) ([]*model.GiftRecord, int64, error) {
	var records []*model.GiftRecord
	var total int64
	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).Model(&model.GiftRecord{}).
		Where("sender_id = ?", senderID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).Where("sender_id = ?", senderID).
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&records).Error

	return records, total, err
}

func (r *giftRecordRepositoryImpl) GetTotalGiftValueByRoom(ctx context.Context, roomID int64) (int64, error) {
	var totalValue int64
	err := r.db.WithContext(ctx).Model(&model.GiftRecord{}).
		Select("COALESCE(SUM(total_price), 0)").
		Where("room_id = ?", roomID).
		Scan(&totalValue).Error
	return totalValue, err
}

func (r *giftRecordRepositoryImpl) GetTotalGiftValueBySender(ctx context.Context, senderID int64) (int64, error) {
	var totalValue int64
	err := r.db.WithContext(ctx).Model(&model.GiftRecord{}).
		Select("COALESCE(SUM(total_price), 0)").
		Where("sender_id = ?", senderID).
		Scan(&totalValue).Error
	return totalValue, err
}

func (r *giftRecordRepositoryImpl) WithTransaction(ctx context.Context, fn func(txRepo GiftRecordRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &giftRecordRepositoryImpl{db: tx}
		return fn(txRepo)
	})
}

type roomAdminRepositoryImpl struct {
	db *gorm.DB
}

func NewRoomAdminRepository(db *gorm.DB) RoomAdminRepository {
	return &roomAdminRepositoryImpl{db: db}
}

func (r *roomAdminRepositoryImpl) Create(ctx context.Context, admin *model.RoomAdmin) error {
	return r.db.WithContext(ctx).Create(admin).Error
}

func (r *roomAdminRepositoryImpl) Delete(ctx context.Context, roomID, userID int64) error {
	return r.db.WithContext(ctx).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Delete(&model.RoomAdmin{}).Error
}

func (r *roomAdminRepositoryImpl) Find(ctx context.Context, roomID, userID int64) (*model.RoomAdmin, error) {
	var admin model.RoomAdmin
	err := r.db.WithContext(ctx).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		First(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &admin, err
}

func (r *roomAdminRepositoryImpl) ListByRoomID(ctx context.Context, roomID int64) ([]*model.RoomAdmin, error) {
	var admins []*model.RoomAdmin
	err := r.db.WithContext(ctx).Where("room_id = ?", roomID).Find(&admins).Error
	return admins, err
}

func (r *roomAdminRepositoryImpl) ListByUserID(ctx context.Context, userID int64) ([]*model.RoomAdmin, error) {
	var admins []*model.RoomAdmin
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&admins).Error
	return admins, err
}

func (r *roomAdminRepositoryImpl) IsAdmin(ctx context.Context, roomID, userID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.RoomAdmin{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *roomAdminRepositoryImpl) WithTransaction(ctx context.Context, fn func(txRepo RoomAdminRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &roomAdminRepositoryImpl{db: tx}
		return fn(txRepo)
	})
}

type liveRecordRepositoryImpl struct {
	db *gorm.DB
}

func NewLiveRecordRepository(db *gorm.DB) LiveRecordRepository {
	return &liveRecordRepositoryImpl{db: db}
}

func (r *liveRecordRepositoryImpl) Create(ctx context.Context, record *model.LiveRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *liveRecordRepositoryImpl) Update(ctx context.Context, record *model.LiveRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

func (r *liveRecordRepositoryImpl) FindByID(ctx context.Context, id int64) (*model.LiveRecord, error) {
	var record model.LiveRecord
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &record, err
}

func (r *liveRecordRepositoryImpl) FindByRoomID(ctx context.Context, roomID int64) ([]*model.LiveRecord, error) {
	var records []*model.LiveRecord
	err := r.db.WithContext(ctx).Where("room_id = ?", roomID).
		Order("start_time DESC").
		Find(&records).Error
	return records, err
}

func (r *liveRecordRepositoryImpl) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.LiveRecord{}, id).Error
}

func (r *liveRecordRepositoryImpl) WithTransaction(ctx context.Context, fn func(txRepo LiveRecordRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &liveRecordRepositoryImpl{db: tx}
		return fn(txRepo)
	})
}

type roomViewerRepositoryImpl struct {
	db *gorm.DB
}

func NewRoomViewerRepository(db *gorm.DB) RoomViewerRepository {
	return &roomViewerRepositoryImpl{db: db}
}

func (r *roomViewerRepositoryImpl) CreateOrUpdate(ctx context.Context, viewer *model.RoomViewer) error {
	return r.db.WithContext(ctx).Save(viewer).Error
}

func (r *roomViewerRepositoryImpl) Find(ctx context.Context, roomID, userID int64) (*model.RoomViewer, error) {
	var viewer model.RoomViewer
	err := r.db.WithContext(ctx).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		First(&viewer).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &viewer, err
}

func (r *roomViewerRepositoryImpl) ListByRoomID(ctx context.Context, roomID int64, page, pageSize int) ([]*model.RoomViewer, int64, error) {
	var viewers []*model.RoomViewer
	var total int64
	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).Model(&model.RoomViewer{}).
		Where("room_id = ? AND leave_time IS NULL", roomID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).Where("room_id = ? AND leave_time IS NULL", roomID).
		Offset(offset).Limit(pageSize).
		Order("join_time DESC").
		Find(&viewers).Error

	return viewers, total, err
}

func (r *roomViewerRepositoryImpl) CountByRoomID(ctx context.Context, roomID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.RoomViewer{}).
		Where("room_id = ? AND leave_time IS NULL", roomID).
		Count(&count).Error
	return count, err
}

func (r *roomViewerRepositoryImpl) UpdateLeaveTime(ctx context.Context, roomID, userID int64, leaveTime string) error {
	return r.db.WithContext(ctx).Model(&model.RoomViewer{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Update("leave_time", leaveTime).Error
}

func (r *roomViewerRepositoryImpl) Delete(ctx context.Context, roomID, userID int64) error {
	return r.db.WithContext(ctx).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Delete(&model.RoomViewer{}).Error
}

func (r *roomViewerRepositoryImpl) WithTransaction(ctx context.Context, fn func(txRepo RoomViewerRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &roomViewerRepositoryImpl{db: tx}
		return fn(txRepo)
	})
}
