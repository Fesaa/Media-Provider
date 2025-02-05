package impl

import (
	"database/sql"
	"github.com/Fesaa/Media-Provider/db/models"
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
		Model: gorm.Model{
			ID: u,
		},
	}
	return n.db.Model(&model).Updates(&models.Notification{Read: true, ReadAt: sql.NullTime{Time: time.Now()}}).Error
}

func (n notifications) MarkUnread(u uint) error {
	model := models.Notification{
		Model: gorm.Model{
			ID: u,
		},
	}
	return n.db.Model(&model).Updates(&models.Notification{Read: false, ReadAt: sql.NullTime{}}).Error
}

func (n notifications) Unread() (int64, error) {
	var count int64
	err := n.db.Model(&models.Notification{Read: false}).
		Where(&models.Notification{Group: models.GroupSecurity}).
		Or(&models.Notification{Group: models.GroupGeneral}).
		Count(&count).Error
	return count, err
}
