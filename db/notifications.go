package db

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"gorm.io/gorm"
	"time"
)

func Notifications(db *gorm.DB) models.Notifications {
	return &notifications{
		db: db,
	}
}

type notifications struct {
	db *gorm.DB
}

func (n notifications) Recent(i int, group models.NotificationGroup) ([]models.Notification, error) {
	var many []models.Notification
	err := n.db.
		Where(&models.Notification{Group: group}).
		Where(map[string]any{"read": false}).
		Order("created_at desc").
		Limit(i).
		Find(&many).
		Error
	if err != nil {
		return nil, err
	}
	return many, nil
}

func (n notifications) GetMany(ids []uint) ([]models.Notification, error) {
	var many []models.Notification
	err := n.db.Where("id IN (?)", ids).Find(&many).Error
	if err != nil {
		return nil, err
	}
	return many, nil
}

func (n notifications) DeleteMany(ids []uint) error {
	notifs := utils.Map(ids, func(id uint) models.Notification {
		return models.Notification{
			Model: models.Model{
				ID: id,
			},
		}
	})
	return n.db.Delete(notifs).Error
}

func (n notifications) MarkReadMany(ids []uint) error {
	return n.db.Model(&models.Notification{}).
		Where("id IN (?)", ids).
		Updates(&models.Notification{Read: true}).
		Error
}

func (n notifications) Get(id uint) (models.Notification, error) {
	var notification models.Notification
	err := n.db.First(&notification, id).Error
	if err != nil {
		return models.Notification{}, err
	}
	return notification, nil
}

func (n notifications) All() ([]models.Notification, error) {
	var all []models.Notification
	err := n.db.Find(&all).Error
	if err != nil {
		return nil, err
	}
	return all, nil
}

func (n notifications) AllAfter(time time.Time) ([]models.Notification, error) {
	var after []models.Notification
	err := n.db.Where("created_at > ?", time).Find(&after).Error
	if err != nil {
		return nil, err
	}
	return after, nil
}

func (n notifications) New(notification models.Notification) error {
	return n.db.Create(&notification).Error
}

func (n notifications) Delete(u uint) error {
	return n.db.Delete(&models.Notification{}, u).Error
}

func (n notifications) MarkRead(u uint) error {
	model := models.Notification{
		Model: models.Model{
			ID: u,
		},
	}
	return n.db.Model(&model).Updates(&models.Notification{Read: true}).Error
}

func (n notifications) MarkUnread(u uint) error {
	model := models.Notification{
		Model: models.Model{
			ID: u,
		},
	}
	return n.db.Model(&model).Updates(map[string]any{"read": false}).Error
}

func (n notifications) Unread() (int64, error) {
	var count int64
	err := n.db.Model(&models.Notification{}).
		Where(map[string]any{"read": false}).
		Not(&models.Notification{Group: models.GroupContent}).
		Count(&count).Error
	return count, err
}
