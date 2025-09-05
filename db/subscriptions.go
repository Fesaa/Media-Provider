package db

import (
	"errors"

	"github.com/Fesaa/Media-Provider/db/models"
	"gorm.io/gorm"
)

type subscriptionImpl struct {
	db *gorm.DB
}

func Subscriptions(db *gorm.DB) models.Subscriptions {
	return &subscriptionImpl{db}
}

func (s subscriptionImpl) All() ([]models.Subscription, error) {
	var subscriptions []models.Subscription
	res := s.db.Preload("Info").Find(&subscriptions)
	if res.Error != nil {
		return nil, res.Error
	}

	return subscriptions, nil
}

func (s subscriptionImpl) AllForUser(userID uint) ([]models.Subscription, error) {
	var subscriptions []models.Subscription
	res := s.db.
		Preload("Info").
		Where(&models.Subscription{Owner: userID}).
		Find(&subscriptions)
	if res.Error != nil {
		return nil, res.Error
	}

	return subscriptions, nil
}

func (s subscriptionImpl) Get(i uint) (*models.Subscription, error) {
	var subscription models.Subscription
	res := s.db.Preload("Info").First(&subscription, i)
	if res.Error != nil {
		return nil, res.Error
	}

	return &subscription, nil
}

func (s subscriptionImpl) GetForUser(subId, userID uint) (models.Subscription, error) {
	var subscription models.Subscription
	res := s.db.
		Preload("Info").
		Where(&models.Subscription{Owner: userID}).
		First(&subscription, subId)

	return subscription, res.Error
}

func (s subscriptionImpl) GetByContentId(contentID string) (*models.Subscription, error) {
	var subscription models.Subscription
	res := s.db.Preload("Info").
		Where(&models.Subscription{ContentId: contentID}).
		First(&subscription)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if res.Error != nil {
		return nil, res.Error
	}

	return &subscription, nil
}

func (s subscriptionImpl) GetByContentIdForUser(contentID string, userId uint) (*models.Subscription, error) {
	var subscription models.Subscription
	res := s.db.Preload("Info").
		Where(&models.Subscription{ContentId: contentID, Owner: userId}).
		First(&subscription)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if res.Error != nil {
		return nil, res.Error
	}

	return &subscription, nil
}

func (s subscriptionImpl) New(subscription models.Subscription) (*models.Subscription, error) {
	res := s.db.Create(&subscription)
	if res.Error != nil {
		return nil, res.Error
	}
	return &subscription, nil
}

func (s subscriptionImpl) Update(subscription models.Subscription) error {
	return s.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&subscription).Error
}

func (s subscriptionImpl) Delete(i uint) error {
	return s.db.Select("Info").Delete(&models.Subscription{
		Model: models.Model{ID: i},
	}).Error
}
