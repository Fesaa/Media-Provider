package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"gorm.io/gorm"
)

type UserRepository interface {
	// GetAllUsers returns all users in the system
	GetAllUsers(context.Context) ([]models.User, error)
	// ExistsAny checks if at least one user exists
	ExistsAny(context.Context) (bool, error)

	// GetByID retrieves a user by internal ID
	GetByID(context.Context, int) (*models.User, error)
	// GetByExternalID retrieves a user by external identifier
	GetByExternalID(context.Context, string) (*models.User, error)
	// GetByEmail retrieves a user by email address
	GetByEmail(context.Context, string) (*models.User, error)
	// GetByName retrieves a user by name
	GetByName(context.Context, string) (*models.User, error)
	// GetByAPIKey retrieves a user by API key
	GetByAPIKey(context.Context, string) (*models.User, error)

	// Create creates a new user
	Create(context.Context, models.User) error
	// Update updates an existing user
	Update(context.Context, models.User) error

	// GenerateReset generates a password reset token for a user
	GenerateReset(context.Context, int) (*models.PasswordReset, error)
	// GetResetByUserID retrieves a password reset token by user ID
	GetResetByUserID(context.Context, int) (*models.PasswordReset, error)
	// GetReset retrieves a password reset token by its key
	GetReset(context.Context, string) (*models.PasswordReset, error)
	// DeleteReset deletes a password reset token by its key
	DeleteReset(context.Context, string) error
	// Delete deletes a user by ID
	Delete(context.Context, int) error
}

type userRepository struct {
	db *gorm.DB
}

func (u userRepository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	res := u.db.WithContext(ctx).Find(&users)
	if res.Error != nil {
		return nil, res.Error
	}
	return users, nil
}

func (u userRepository) ExistsAny(ctx context.Context) (bool, error) {
	var count int64
	err := u.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (u userRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	var user models.User
	result := u.db.WithContext(ctx).First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (u userRepository) GetByExternalID(ctx context.Context, externalID string) (*models.User, error) {
	var user models.User
	result := u.db.WithContext(ctx).Where("external_id = ?", externalID).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (u userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	result := u.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (u userRepository) GetByName(ctx context.Context, name string) (*models.User, error) {
	var user models.User
	result := u.db.WithContext(ctx).Where(&models.User{Name: name}).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (u userRepository) GetByAPIKey(ctx context.Context, key string) (*models.User, error) {
	var user models.User
	result := u.db.WithContext(ctx).Where(&models.User{ApiKey: key}).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (u userRepository) Create(ctx context.Context, user models.User) error {
	user.ID = 0
	return u.db.WithContext(ctx).Create(&user).Error
}

func (u userRepository) Update(ctx context.Context, user models.User) error {
	return u.db.WithContext(ctx).Save(&user).Error
}

func (u userRepository) GenerateReset(ctx context.Context, userId int) (*models.PasswordReset, error) {
	key, err := utils.GenerateUrlSecret(32)
	if err != nil {
		return nil, err
	}

	reset := models.PasswordReset{
		UserId: userId,
		Key:    key,
		Expiry: time.Now().Add(time.Hour * 24),
	}

	res := u.db.WithContext(ctx).Create(&reset)

	if res.Error != nil {
		return nil, res.Error
	}

	return &reset, nil
}

func (u userRepository) GetResetByUserID(ctx context.Context, userId int) (*models.PasswordReset, error) {
	var reset models.PasswordReset
	res := u.db.WithContext(ctx).Where(&models.PasswordReset{UserId: userId}).First(&reset)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, res.Error
	}

	return &reset, nil
}

func (u userRepository) GetReset(ctx context.Context, key string) (*models.PasswordReset, error) {
	var reset models.PasswordReset
	res := u.db.WithContext(ctx).Where(&models.PasswordReset{Key: key}).First(&reset)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, res.Error
	}

	return &reset, nil
}

func (u userRepository) DeleteReset(ctx context.Context, key string) error {
	return u.db.WithContext(ctx).Delete(&models.PasswordReset{}, "key = ?", key).Error
}

func (u userRepository) Delete(ctx context.Context, userId int) error {
	return u.db.WithContext(ctx).Unscoped().Delete(&models.User{}, "id = ?", userId).Error
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}
