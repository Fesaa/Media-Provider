package repository

import (
	"context"
	"time"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"gorm.io/gorm"
)

type NotificationsRepository interface {
	// Get retrieves a notification by ID
	Get(context.Context, int) (models.Notification, error)
	// GetMany retrieves multiple notifications by their IDs
	GetMany(context.Context, []int) ([]models.Notification, error)
	// All returns all notifications
	All(context.Context) ([]models.Notification, error)
	// AllAfter returns all notifications created after a given time
	AllAfter(context.Context, time.Time) ([]models.Notification, error)
	// Recent returns the most recent notifications for a group
	Recent(context.Context, int, models.NotificationGroup) ([]models.Notification, error)
	// New creates a new notification
	New(context.Context, models.Notification) error
	// Delete removes a notification by ID
	Delete(context.Context, int) error
	// DeleteMany removes multiple notifications by their IDs
	DeleteMany(context.Context, []int) error
	// MarkRead marks a notification as read
	MarkRead(context.Context, int) error
	// MarkReadMany marks multiple notifications as read
	MarkReadMany(context.Context, []int) error
	// MarkUnread marks a notification as unread
	MarkUnread(context.Context, int) error
	// Unread returns the count of unread notifications
	Unread(context.Context) (int64, error)
}

type notificationsRepository struct {
	db *gorm.DB
}

func (r notificationsRepository) Get(ctx context.Context, id int) (models.Notification, error) {
	var notification models.Notification
	err := r.db.WithContext(ctx).First(&notification, id).Error
	if err != nil {
		return models.Notification{}, err
	}
	return notification, nil
}

func (r notificationsRepository) GetMany(ctx context.Context, ids []int) ([]models.Notification, error) {
	var many []models.Notification
	err := r.db.WithContext(ctx).Where("id IN (?)", ids).Find(&many).Error
	if err != nil {
		return nil, err
	}
	return many, nil
}

func (r notificationsRepository) All(ctx context.Context) ([]models.Notification, error) {
	var all []models.Notification
	err := r.db.WithContext(ctx).Find(&all).Error
	if err != nil {
		return nil, err
	}
	return all, nil
}

func (r notificationsRepository) AllAfter(ctx context.Context, t time.Time) ([]models.Notification, error) {
	var after []models.Notification
	err := r.db.WithContext(ctx).Where("created_at > ?", t).Find(&after).Error
	if err != nil {
		return nil, err
	}
	return after, nil
}

func (r notificationsRepository) Recent(ctx context.Context, limit int, group models.NotificationGroup) ([]models.Notification, error) {
	var many []models.Notification
	err := r.db.WithContext(ctx).
		Where(&models.Notification{Group: group}).
		Where(map[string]any{"read": false}).
		Order("created_at desc").
		Limit(limit).
		Find(&many).Error
	if err != nil {
		return nil, err
	}
	return many, nil
}

func (r notificationsRepository) New(ctx context.Context, notification models.Notification) error {
	notification.ID = 0
	return r.db.WithContext(ctx).Create(&notification).Error
}

func (r notificationsRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.Notification{}, id).Error
}

func (r notificationsRepository) DeleteMany(ctx context.Context, ids []int) error {
	notifs := utils.Map(ids, func(id int) models.Notification {
		return models.Notification{Model: models.Model{ID: id}}
	})
	return r.db.WithContext(ctx).Delete(notifs).Error
}

func (r notificationsRepository) MarkRead(ctx context.Context, id int) error {
	model := models.Notification{Model: models.Model{ID: id}}
	return r.db.WithContext(ctx).Model(&model).Updates(&models.Notification{Read: true}).Error
}

func (r notificationsRepository) MarkReadMany(ctx context.Context, ids []int) error {
	return r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id IN (?)", ids).
		Updates(&models.Notification{Read: true}).Error
}

func (r notificationsRepository) MarkUnread(ctx context.Context, id int) error {
	model := models.Notification{Model: models.Model{ID: id}}
	return r.db.WithContext(ctx).Model(&model).Updates(map[string]any{"read": false}).Error
}

func (r notificationsRepository) Unread(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where(map[string]any{"read": false}).
		Not(&models.Notification{Group: models.GroupContent}).
		Count(&count).Error
	return count, err
}

func NewNotificationsRepository(db *gorm.DB) NotificationsRepository {
	return &notificationsRepository{db: db}
}
