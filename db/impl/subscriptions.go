package impl

import (
	"fmt"
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

func (s subscriptionImpl) Get(i uint) (*models.Subscription, error) {
	var subscription models.Subscription
	res := s.db.Preload("Info").First(&subscription, i)
	if res.Error != nil {
		return nil, res.Error
	}

	return &subscription, nil
}

func (s subscriptionImpl) New(subscription models.Subscription) (*models.Subscription, error) {
	fmt.Printf("%+v\n", subscription)
	res := s.db.Create(&subscription)
	if res.Error != nil {
		return nil, res.Error
	}
	return &subscription, nil
}

func (s subscriptionImpl) Update(subscription models.Subscription) error {
	res := s.db.Save(&subscription)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (s subscriptionImpl) Delete(i uint) error {
	return s.db.Delete(&models.Subscription{}, i).Error
}
