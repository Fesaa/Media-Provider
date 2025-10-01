package repository

import (
	"context"
	"errors"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"gorm.io/gorm"
)

type SubscriptionsRepository interface {
	// All returns all subscriptions
	All(context.Context) ([]models.Subscription, error)
	// AllPaginated returns all subscriptions, with a give offset and max amount
	AllPaginated(context.Context, utils.UserParams) (utils.PagedList[models.Subscription], error)
	// AllForUser returns all subscriptions for a given user
	AllForUser(context.Context, int) ([]models.Subscription, error)
	// AllForUserPaginated returns all subscriptions for a given user, with a give offset and max amount
	AllForUserPaginated(context.Context, int, utils.UserParams) (utils.PagedList[models.Subscription], error)
	// Get retrieves a subscription by ID
	Get(context.Context, int) (*models.Subscription, error)
	// GetForUser retrieves a subscription by ID and user
	GetForUser(context.Context, int, int) (models.Subscription, error)
	// GetByContentID retrieves a subscription by content ID
	GetByContentID(context.Context, string) (*models.Subscription, error)
	// GetByContentIDForUser retrieves a subscription by content ID and user
	GetByContentIDForUser(context.Context, string, int) (*models.Subscription, error)
	// New creates a new subscription
	New(context.Context, models.Subscription) (*models.Subscription, error)
	// Update updates an existing subscription
	Update(context.Context, models.Subscription) error
	// Delete deletes a subscription by ID
	Delete(context.Context, int) error
}

type subscriptionsRepository struct {
	db *gorm.DB
}

func (r subscriptionsRepository) All(ctx context.Context) ([]models.Subscription, error) {
	var subscriptions []models.Subscription
	result := r.db.WithContext(ctx).Find(&subscriptions)
	if result.Error != nil {
		return nil, result.Error
	}
	return subscriptions, nil
}

func (r subscriptionsRepository) AllPaginated(ctx context.Context, params utils.UserParams) (utils.PagedList[models.Subscription], error) {
	query := r.db.WithContext(ctx).Model(&models.Subscription{}).Order("title asc")
	return utils.NewPageListFromUserParams[models.Subscription](ctx, query, params)
}

func (r subscriptionsRepository) AllForUser(ctx context.Context, userID int) ([]models.Subscription, error) {
	var subscriptions []models.Subscription
	result := r.db.WithContext(ctx).
		Where(&models.Subscription{Owner: userID}).
		Find(&subscriptions)
	if result.Error != nil {
		return nil, result.Error
	}
	return subscriptions, nil
}

func (r subscriptionsRepository) AllForUserPaginated(ctx context.Context, userID int, params utils.UserParams) (utils.PagedList[models.Subscription], error) {
	query := r.db.WithContext(ctx).Model(&models.Subscription{}).Where(&models.Subscription{Owner: userID}).Order("title asc")
	return utils.NewPageListFromUserParams[models.Subscription](ctx, query, params)
}

func (r subscriptionsRepository) Get(ctx context.Context, id int) (*models.Subscription, error) {
	var subscription models.Subscription
	result := r.db.WithContext(ctx).First(&subscription, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &subscription, nil
}

func (r subscriptionsRepository) GetForUser(ctx context.Context, subID int, userID int) (models.Subscription, error) {
	var subscription models.Subscription
	result := r.db.WithContext(ctx).
		Where(&models.Subscription{Owner: userID}).
		First(&subscription, subID)
	return subscription, result.Error
}

func (r subscriptionsRepository) GetByContentID(ctx context.Context, contentID string) (*models.Subscription, error) {
	var subscription models.Subscription
	result := r.db.WithContext(ctx).
		Where(&models.Subscription{ContentId: contentID}).
		First(&subscription)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &subscription, nil
}

func (r subscriptionsRepository) GetByContentIDForUser(ctx context.Context, contentID string, userID int) (*models.Subscription, error) {
	var subscription models.Subscription
	result := r.db.WithContext(ctx).
		Where(&models.Subscription{ContentId: contentID, Owner: userID}).
		First(&subscription)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &subscription, nil
}

func (r subscriptionsRepository) New(ctx context.Context, subscription models.Subscription) (*models.Subscription, error) {
	subscription.ID = 0
	result := r.db.WithContext(ctx).Create(&subscription)
	if result.Error != nil {
		return nil, result.Error
	}
	return &subscription, nil
}

func (r subscriptionsRepository) Update(ctx context.Context, subscription models.Subscription) error {
	return r.db.WithContext(ctx).Save(&subscription).Error
}

func (r subscriptionsRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.Subscription{Model: models.Model{ID: id}}).Error
}

func NewSubscriptionsRepository(db *gorm.DB) SubscriptionsRepository {
	return &subscriptionsRepository{db: db}
}
