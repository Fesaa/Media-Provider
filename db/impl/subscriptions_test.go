package impl

import (
	"errors"
	"github.com/Fesaa/Media-Provider/db/models"
	"gorm.io/gorm"
	"testing"
)

func TestSubscriptions_All(t *testing.T) {
	db := databaseHelper(t)
	s := Subscriptions(db)

	for i := 0; i < 5; i++ {
		sub := models.Subscription{
			ContentId: "contentID" + string(rune(i)),
			Info: models.SubscriptionInfo{
				Title: "Name" + string(rune(i)),
			},
		}
		if _, err := s.New(sub); err != nil {
			t.Fatalf("failed to create subscription: %v", err)
		}
	}

	subs, err := s.All()
	if err != nil {
		t.Fatalf("failed to get all subscriptions: %v", err)
	}

	if len(subs) != 5 {
		t.Fatalf("expected 5 subscriptions, got %d", len(subs))
	}

	for _, sub := range subs {
		if sub.Info.Title == "" {
			t.Fatalf("subscription info not loaded")
		}
	}
}

func TestSubscriptions_Get(t *testing.T) {
	db := databaseHelper(t)
	s := Subscriptions(db)

	sub := models.Subscription{
		ContentId: "testContentID",
		Info: models.SubscriptionInfo{
			Title: "TestName",
		},
	}
	createdSub, err := s.New(sub)
	if err != nil {
		t.Fatalf("failed to create subscription: %v", err)
	}

	retrievedSub, err := s.Get(createdSub.ID)
	if err != nil {
		t.Fatalf("failed to get subscription: %v", err)
	}

	if retrievedSub.ContentId != "testContentID" {
		t.Fatalf("expected content ID %s, got %s", "testContentID", retrievedSub.ContentId)
	}

	if retrievedSub.Info.Title != "TestName" {
		t.Fatalf("expected name %s, got %s", "TestName", retrievedSub.Info.Title)
	}
}

func TestSubscriptions_GetByContentId(t *testing.T) {
	db := databaseHelper(t)
	s := Subscriptions(db)

	sub := models.Subscription{
		ContentId: "uniqueContentID",
		Info: models.SubscriptionInfo{
			Title: "UniqueName",
		},
	}
	_, err := s.New(sub)
	if err != nil {
		t.Fatalf("failed to create subscription: %v", err)
	}

	retrievedSub, err := s.GetByContentId("uniqueContentID")
	if err != nil {
		t.Fatalf("failed to get subscription by content ID: %v", err)
	}

	if retrievedSub == nil {
		t.Fatalf("expected subscription, got nil")
	}

	if retrievedSub.Info.Title != "UniqueName" {
		t.Fatalf("expected name %s, got %s", "UniqueName", retrievedSub.Info.Title)
	}

	retrievedNilSub, err := s.GetByContentId("nonexistentContentID")
	if err != nil {
		t.Fatalf("error getting non existent content ID: %v", err)
	}

	if retrievedNilSub != nil {
		t.Fatalf("expected nil for non-existent content ID, got subscription")
	}
}

func TestSubscriptions_New(t *testing.T) {
	db := databaseHelper(t)
	s := Subscriptions(db)

	sub := models.Subscription{
		ContentId: "newContentID",
		Info: models.SubscriptionInfo{
			Title: "NewName",
		},
	}
	createdSub, err := s.New(sub)
	if err != nil {
		t.Fatalf("failed to create subscription: %v", err)
	}

	if createdSub.ContentId != "newContentID" {
		t.Fatalf("expected content ID %s, got %s", "newContentID", createdSub.ContentId)
	}
}

func TestSubscriptions_Update(t *testing.T) {
	db := databaseHelper(t)
	s := Subscriptions(db)

	sub := models.Subscription{
		ContentId: "updateContentID",
		Info: models.SubscriptionInfo{
			Title: "InitialName",
		},
	}
	_, err := s.New(sub)
	if err != nil {
		t.Fatalf("failed to create subscription: %v", err)
	}

	createdSub, err := s.GetByContentId("updateContentID")
	if err != nil {
		t.Fatalf("failed to get subscription by content ID: %v", err)
	}

	createdSub.Info.Title = "UpdatedName"
	err = s.Update(*createdSub)
	if err != nil {
		t.Fatalf("failed to update subscription: %v", err)
	}

	updatedSub, err := s.GetByContentId("updateContentID")
	if err != nil {
		t.Fatalf("failed to get updated subscription: %v", err)
	}

	if updatedSub.Info.Title != "UpdatedName" {
		t.Fatalf("expected name %s, got %s", "UpdatedName", updatedSub.Info.Title)
	}
}

func TestSubscriptions_Delete(t *testing.T) {
	db := databaseHelper(t)
	s := Subscriptions(db)

	sub := models.Subscription{
		ContentId: "deleteContentID",
		Info: models.SubscriptionInfo{
			Title: "DeleteName",
		},
	}
	createdSub, err := s.New(sub)
	if err != nil {
		t.Fatalf("failed to create subscription: %v", err)
	}

	err = s.Delete(createdSub.ID)
	if err != nil {
		t.Fatalf("failed to delete subscription: %v", err)
	}

	_, err = s.Get(createdSub.ID)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected record not found error, got %v", err)
	}
}
