package db

import (
	"errors"
	"github.com/Fesaa/Media-Provider/db/models"
	"gorm.io/gorm"
	"testing"
)

func TestUsers_All(t *testing.T) {
	db := databaseHelper(t)
	u := Users(db)

	_, err := u.Create("User1")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	_, err = u.Create("User2")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	users, err := u.All()
	if err != nil {
		t.Fatalf("failed to get all users: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
}

func TestUsers_ExistsAny(t *testing.T) {
	db := databaseHelper(t)
	u := Users(db)

	exists, err := u.ExistsAny()
	if err != nil {
		t.Fatalf("failed to check if users exist: %v", err)
	}

	if exists {
		t.Fatalf("expected no users to exist, but found some")
	}

	_, err = u.Create("User1")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	exists, err = u.ExistsAny()
	if err != nil {
		t.Fatalf("failed to check if users exist: %v", err)
	}

	if !exists {
		t.Fatalf("expected users to exist, but found none")
	}
}

func TestUsers_GetById(t *testing.T) {
	db := databaseHelper(t)
	u := Users(db)

	createdUser, err := u.Create("TestUser")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	retrievedUser, err := u.GetById(createdUser.ID)
	if err != nil {
		t.Fatalf("failed to get user by ID: %v", err)
	}

	if retrievedUser.Name != "TestUser" {
		t.Fatalf("expected name %s, got %s", "TestUser", retrievedUser.Name)
	}
}

func TestUsers_GetByName(t *testing.T) {
	db := databaseHelper(t)
	u := Users(db)

	_, err := u.Create("TestUser")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	retrievedUser, err := u.GetByName("TestUser")
	if err != nil {
		t.Fatalf("failed to get user by name: %v", err)
	}

	if retrievedUser.Name != "TestUser" {
		t.Fatalf("expected name %s, got %s", "TestUser", retrievedUser.Name)
	}
}

func TestUsers_GetByApiKey(t *testing.T) {
	db := databaseHelper(t)
	u := Users(db)

	apiKey := "testApiKey"
	_, err := u.Create("TestUser", func(user models.User) models.User {
		user.ApiKey = apiKey
		return user
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	retrievedUser, err := u.GetByApiKey(apiKey)
	if err != nil {
		t.Fatalf("failed to get user by API key: %v", err)
	}

	if retrievedUser.ApiKey != apiKey {
		t.Fatalf("expected API key %s, got %s", apiKey, retrievedUser.ApiKey)
	}
}

func TestUsers_Create(t *testing.T) {
	db := databaseHelper(t)
	u := Users(db)

	createdUser, err := u.Create("NewUser")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if createdUser.Name != "NewUser" {
		t.Fatalf("expected name %s, got %s", "NewUser", createdUser.Name)
	}
}

func TestUsers_Update(t *testing.T) {
	db := databaseHelper(t)
	u := Users(db)

	createdUser, err := u.Create("InitialUser")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	updatedUser, err := u.Update(*createdUser, func(user models.User) models.User {
		user.Name = "UpdatedUser"
		return user
	})
	if err != nil {
		t.Fatalf("failed to update user: %v", err)
	}

	if updatedUser.Name != "UpdatedUser" {
		t.Fatalf("expected name %s, got %s", "UpdatedUser", updatedUser.Name)
	}
}

func TestUsers_UpdateById(t *testing.T) {
	db := databaseHelper(t)
	u := Users(db)

	createdUser, err := u.Create("InitialUser")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	updatedUser, err := u.UpdateById(createdUser.ID, func(user models.User) models.User {
		user.Name = "UpdatedUser"
		return user
	})
	if err != nil {
		t.Fatalf("failed to update user by ID: %v", err)
	}

	if updatedUser.Name != "UpdatedUser" {
		t.Fatalf("expected name %s, got %s", "UpdatedUser", updatedUser.Name)
	}
}

func TestUsers_GenerateReset(t *testing.T) {
	db := databaseHelper(t)
	u := Users(db)

	createdUser, err := u.Create("TestUser")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	reset, err := u.GenerateReset(createdUser.ID)
	if err != nil {
		t.Fatalf("failed to generate password reset: %v", err)
	}

	if reset.UserId != createdUser.ID {
		t.Fatalf("expected user ID %d, got %d", createdUser.ID, reset.UserId)
	}

	if reset.Key == "" {
		t.Fatalf("expected reset key to be generated")
	}
}

func TestUsers_GetReset(t *testing.T) {
	db := databaseHelper(t)
	u := Users(db)

	createdUser, err := u.Create("TestUser")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	reset, err := u.GenerateReset(createdUser.ID)
	if err != nil {
		t.Fatalf("failed to generate password reset: %v", err)
	}

	retrievedReset, err := u.GetReset(reset.Key)
	if err != nil {
		t.Fatalf("failed to get password reset: %v", err)
	}

	if retrievedReset.Key != reset.Key {
		t.Fatalf("expected key %s, got %s", reset.Key, retrievedReset.Key)
	}
}

func TestUsers_GetResetByUserId(t *testing.T) {
	db := databaseHelper(t)
	u := Users(db)

	createdUser, err := u.Create("TestUser")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	reset, err := u.GenerateReset(createdUser.ID)
	if err != nil {
		t.Fatalf("failed to generate password reset: %v", err)
	}

	retrievedReset, err := u.GetResetByUserId(createdUser.ID)
	if err != nil {
		t.Fatalf("failed to get password reset by user ID: %v", err)
	}

	if retrievedReset.Key != reset.Key {
		t.Fatalf("expected key %s, got %s", reset.Key, retrievedReset.Key)
	}

	noReset, err := u.GetResetByUserId(createdUser.ID + 1)
	if err != nil {
		t.Fatalf("error getting reset by non-existing user id: %v", err)
	}

	if noReset != nil {
		t.Fatalf("expected nil reset when user ID doesn't exist")
	}
}

func TestUsers_DeleteReset(t *testing.T) {
	db := databaseHelper(t)
	u := Users(db)

	createdUser, err := u.Create("TestUser")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	reset, err := u.GenerateReset(createdUser.ID)
	if err != nil {
		t.Fatalf("failed to generate password reset: %v", err)
	}

	err = u.DeleteReset(reset.Key)
	if err != nil {
		t.Fatalf("failed to delete password reset: %v", err)
	}

	_, err = u.GetReset(reset.Key)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected record not found error, got %v", err)
	}
}

func TestUsers_Delete(t *testing.T) {
	db := databaseHelper(t)
	u := Users(db)

	createdUser, err := u.Create("TestUser")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	err = u.Delete(createdUser.ID)
	if err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}

	_, err = u.GetById(createdUser.ID)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected record not found error, got %v", err)
	}
}
