package db

import (
	"errors"
	"testing"
	"time"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"gorm.io/gorm"
)

func TestNotifications_All(t *testing.T) {
	n := Notifications(databaseHelper(t))

	for range 10 {
		if err := n.New(models.Notification{}); err != nil {
			t.Fatal(err)
		}
	}

	if err := n.New(models.Notification{
		Title: "Test Notification",
	}); err != nil {
		t.Fatal(err)
	}

	want := 11
	got, err := n.All()
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != want {
		t.Fatalf("got %d, want %d", len(got), want)
	}

	testN := utils.Find(got, func(notification models.Notification) bool {
		return notification.Title == "Test Notification"
	})

	if testN == nil {
		t.Fatal("notification not found")
	}

}

func TestNotifications_AllAfter(t *testing.T) {
	n := Notifications(databaseHelper(t))

	for range 5 {
		if err := n.New(models.Notification{
			Model: models.Model{
				CreatedAt: time.Now().Add(time.Hour * -24),
			},
		}); err != nil {
			t.Fatal(err)
		}
	}

	for range 5 {
		if err := n.New(models.Notification{
			Model: models.Model{
				CreatedAt: time.Now().Add(time.Hour),
			},
		}); err != nil {
			t.Fatal(err)
		}
	}

	want := 5
	got, err := n.AllAfter(time.Now())
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != want {
		t.Fatalf("got %d, want %d", len(got), want)
	}
}

func TestNotifications_Delete(t *testing.T) {
	n := Notifications(databaseHelper(t))

	if err := n.New(models.Notification{}); err != nil {
		t.Fatal(err)
	}

	if err := n.Delete(1); err != nil {
		t.Fatal(err)
	}

	_, err := n.Get(1)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("got nil, want error")
	}

}

func TestNotifications_DeleteMany(t *testing.T) {
	n := Notifications(databaseHelper(t))

	for range 10 {
		if err := n.New(models.Notification{}); err != nil {
			t.Fatal(err)
		}
	}

	if err := n.DeleteMany([]uint{1, 2, 3, 4, 5}); err != nil {
		t.Fatal(err)
	}

	_, err := n.Get(1)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("got nil, want error")
	}

	got, err := n.All()
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 5 {
		t.Fatalf("got %d, want %d", len(got), 5)
	}
}

func TestNotifications_Get(t *testing.T) {
	n := Notifications(databaseHelper(t))

	if err := n.New(models.Notification{
		Title: "Test Notification",
	}); err != nil {
		t.Fatal(err)
	}

	got, err := n.Get(1)
	if err != nil {
		t.Fatal(err)
	}

	if got.Title != "Test Notification" {
		t.Fatalf("got %v, want %v", got, "Test Notification")
	}

}

func TestNotifications_MarkRead(t *testing.T) {
	n := Notifications(databaseHelper(t))

	if err := n.New(models.Notification{}); err != nil {
		t.Fatal(err)
	}

	if err := n.MarkRead(1); err != nil {
		t.Fatal(err)
	}

	got, err := n.Get(1)
	if err != nil {
		t.Fatal(err)
	}

	if !got.Read {
		t.Fatalf("got %v, want true", got)
	}
}

func TestNotifications_MarkReadMany(t *testing.T) {
	n := Notifications(databaseHelper(t))

	for range 10 {
		if err := n.New(models.Notification{}); err != nil {
			t.Fatal(err)
		}
	}

	if err := n.MarkReadMany([]uint{1, 2, 3, 4, 5}); err != nil {
		t.Fatal(err)
	}

	got, err := n.All()
	if err != nil {
		t.Fatal(err)
	}

	read := utils.Filter(got, func(notification models.Notification) bool {
		return notification.Read
	})

	if len(read) != 5 {
		t.Fatalf("got %d, want %d", len(read), 5)
	}
}

func TestNotifications_MarkUnread(t *testing.T) {
	n := Notifications(databaseHelper(t))

	if err := n.New(models.Notification{
		Read: true,
	}); err != nil {
		t.Fatal(err)
	}

	if err := n.MarkUnread(1); err != nil {
		t.Fatal(err)
	}

	got, err := n.Get(1)
	if err != nil {
		t.Fatal(err)
	}

	if got.Read {
		t.Fatalf("got %v, want false", got)
	}
}

func TestNotifications_New(t *testing.T) {
	n := Notifications(databaseHelper(t))
	if err := n.New(models.Notification{}); err != nil {
		t.Fatal(err)
	}

	_, err := n.Get(1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNotifications_Unread(t *testing.T) {
	n := Notifications(databaseHelper(t))

	for range 5 {
		if err := n.New(models.Notification{}); err != nil {
			t.Fatal(err)
		}
	}

	for range 5 {
		if err := n.New(models.Notification{
			Read: true,
		}); err != nil {
			t.Fatal(err)
		}
	}

	want := int64(5)
	got, err := n.Unread()
	if err != nil {
		t.Fatal(err)
	}

	if want != got {
		t.Fatalf("got %d, want %d", got, want)
	}

	if err = n.New(models.Notification{}); err != nil {
		t.Fatal(err)
	}
	want++

	got, err = n.Unread()
	if err != nil {
		t.Fatal(err)
	}

	if want != got {
		t.Fatalf("got %d, want %d", got, want)
	}

}
