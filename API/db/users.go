package db

import (
	"errors"
	"time"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"gorm.io/gorm"
)

type userImpl struct {
	db *gorm.DB
}

func Users(db *gorm.DB) models.Users {
	return &userImpl{db}
}

func (u *userImpl) All() ([]models.User, error) {
	var users []models.User
	res := u.db.Find(&users)
	if res.Error != nil {
		return nil, res.Error
	}

	return users, nil
}

func (u *userImpl) ExistsAny() (bool, error) {
	var size int64
	err := u.db.Model(&models.User{}).Count(&size).Error
	if err != nil {
		return false, err
	}
	return size > 0, nil
}

func (u *userImpl) GetById(id int) (*models.User, error) {
	var user models.User
	res := u.db.First(&user, id)
	if res.Error != nil {
		return nil, res.Error
	}

	return &user, nil
}

func (u *userImpl) GetByExternalId(externalId string) (*models.User, error) {
	var user models.User
	res := u.db.Where("external_id = ?", externalId).First(&user)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if res.Error != nil {
		return nil, res.Error
	}
	return &user, nil
}

func (u *userImpl) GetByEmail(email string) (*models.User, error) {
	var user models.User
	res := u.db.Where("email = ?", email).First(&user)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if res.Error != nil {
		return nil, res.Error
	}
	return &user, nil
}

func (u *userImpl) GetByName(name string) (*models.User, error) {
	var user models.User
	res := u.db.Where(&models.User{Name: name}).First(&user)
	if res.Error != nil {
		return nil, res.Error
	}
	return &user, nil
}

func (u *userImpl) GetByApiKey(key string) (*models.User, error) {
	var user models.User
	res := u.db.Where(&models.User{ApiKey: key}).First(&user)
	if res.Error != nil {
		return nil, res.Error
	}

	return &user, nil
}

func (u *userImpl) Create(name string, opts ...models.Option[models.User]) (*models.User, error) {
	user := models.User{
		Name: name,
	}

	for _, opt := range opts {
		user = opt(user)
	}

	res := u.db.Create(&user)
	if res.Error != nil {
		return nil, res.Error
	}

	return &user, nil
}

func (u *userImpl) Update(user models.User, opts ...models.Option[models.User]) (*models.User, error) {
	for _, opt := range opts {
		user = opt(user)
	}

	res := u.db.Save(&user)
	if res.Error != nil {
		return nil, res.Error
	}

	return &user, nil
}

func (u *userImpl) UpdateById(id int, opts ...models.Option[models.User]) (*models.User, error) {
	var user models.User
	res := u.db.Where("id = ?", id).First(&user)
	if res.Error != nil {
		return nil, res.Error
	}
	return u.Update(user, opts...)
}

func (u *userImpl) GenerateReset(userId int) (*models.PasswordReset, error) {
	key, err := utils.GenerateUrlSecret(32)
	if err != nil {
		return nil, err
	}

	reset := models.PasswordReset{
		UserId: userId,
		Key:    key,
		Expiry: time.Now().Add(time.Hour * 24),
	}

	res := u.db.Create(&reset)

	if res.Error != nil {
		return nil, res.Error
	}

	return &reset, nil
}

func (u *userImpl) GetReset(key string) (*models.PasswordReset, error) {
	var reset models.PasswordReset
	res := u.db.Where(&models.PasswordReset{Key: key}).First(&reset)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, res.Error
	}
	return &reset, nil
}

func (u *userImpl) GetResetByUserId(userId int) (*models.PasswordReset, error) {
	var reset models.PasswordReset
	res := u.db.Where(&models.PasswordReset{UserId: userId}).First(&reset)
	if res.Error != nil {
		if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, res.Error
		}
		return nil, nil
	}
	return &reset, nil
}

func (u *userImpl) DeleteReset(key string) error {
	return u.db.Delete(&models.PasswordReset{}, "key = ?", key).Error
}

func (u *userImpl) Delete(id int) error {
	return u.db.Unscoped().Delete(&models.User{}, "id = ?", id).Error
}
