package models

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestSubscription_BeforeSave(t *testing.T) {
	sub := Subscription{
		Payload: DownloadRequestMetadata{
			Extra: map[string][]string{
				"key1": {"value1", "value2"},
				"key2": {"value3"},
			},
		},
	}

	err := sub.BeforeSave(nil) // No need for gorm.DB here
	if err != nil {
		t.Errorf("BeforeSave failed: %v", err)
	}

	var metadata map[string][]string
	err = json.Unmarshal(sub.Metadata, &metadata)
	if err != nil {
		t.Errorf("Unmarshal failed: %v", err)
	}

	if len(sub.Payload.Extra) != len(metadata) {
		t.Errorf("Metadata mismatch: expected %v, got %v", sub.Payload.Extra, metadata)
	}

	for key, expectedValues := range sub.Payload.Extra {
		actualValues, ok := metadata[key]
		if !ok {
			t.Errorf("Key %s missing from metadata", key)
		}
		if len(expectedValues) != len(actualValues) {
			t.Errorf("Values mismatch for key %s: expected %v, got %v", key, expectedValues, actualValues)
		}
		for i, expectedValue := range expectedValues {
			if actualValues[i] != expectedValue {
				t.Errorf("Value mismatch for key %s at index %d: expected %s, got %s", key, i, expectedValue, actualValues[i])
			}
		}
	}
}

func TestSubscription_AfterFind(t *testing.T) {
	metadata := map[string][]string{
		"key1": {"value1", "value2"},
		"key2": {"value3"},
	}
	metadataBytes, _ := json.Marshal(metadata)

	sub := Subscription{
		Metadata: metadataBytes,
	}

	err := sub.AfterFind(nil)
	if err != nil {
		t.Errorf("AfterFind failed: %v", err)
	}
	if len(metadata) != len(sub.Payload.Extra) {
		t.Errorf("Metadata mismatch: expected %v, got %v", metadata, sub.Payload.Extra)
	}
	for key, expectedValues := range metadata {
		actualValues, ok := sub.Payload.Extra[key]
		if !ok {
			t.Errorf("Key %s missing from metadata", key)
		}
		if len(expectedValues) != len(actualValues) {
			t.Errorf("Values mismatch for key %s: expected %v, got %v", key, expectedValues, actualValues)
		}
		for i, expectedValue := range expectedValues {
			if actualValues[i] != expectedValue {
				t.Errorf("Value mismatch for key %s at index %d: expected %s, got %s", key, i, expectedValue, actualValues[i])
			}
		}
	}
	if !sub.Payload.StartImmediately {
		t.Errorf("StartImmediately should be true")
	}

	sub2 := Subscription{
		Metadata: nil,
	}

	err2 := sub2.AfterFind(nil)
	if err2 != nil {
		t.Errorf("AfterFind failed: %v", err2)
	}
	if sub2.Payload.Extra != nil {
		t.Errorf("Payload.Extra should be nil")
	}
}

func TestSubscription_ShouldRefresh(t *testing.T) {
	tests := []struct {
		Name     string
		Sub      Subscription
		OldSub   Subscription
		Expected bool
	}{
		{
			Name:     "Same frequency",
			Sub:      Subscription{RefreshFrequency: Day},
			OldSub:   Subscription{RefreshFrequency: Day},
			Expected: false,
		},
		{
			Name:     "Different frequency",
			Sub:      Subscription{RefreshFrequency: Week},
			OldSub:   Subscription{RefreshFrequency: Day},
			Expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := tt.Sub.ShouldRefresh(&tt.OldSub)
			if result != tt.Expected {
				t.Errorf("ShouldRefresh mismatch: expected %v, got %v", tt.Expected, result)
			}
		})
	}
}

func TestSubscription_Normalize(t *testing.T) {
	mockPreferences := mockPreferences{}
	mockPreferences.preferences = Preference{SubscriptionRefreshHour: 10}

	sub := Subscription{
		Info: SubscriptionInfo{
			LastCheck: time.Date(2023, 10, 27, 15, 30, 0, 0, time.Local),
		},
	}

	err := sub.Normalize(&mockPreferences)
	if err != nil {
		t.Errorf("Normalize failed: %v", err)
	}

	expected := time.Date(2023, 10, 27, 10, 0, 0, 0, time.Local)
	if sub.Info.LastCheck != expected {
		t.Errorf("Normalize time mismatch: expected %v, got %v", expected, sub.Info.LastCheck)
	}

	mockPreferences.err = errors.New("test error")
	err = sub.Normalize(&mockPreferences)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestSubscription_normalize(t *testing.T) {
	sub := Subscription{}
	originalTime := time.Date(2023, 10, 27, 15, 30, 0, 0, time.Local)
	normalizedTime := sub.normalize(originalTime, 10)
	expectedTime := time.Date(2023, 10, 27, 10, 0, 0, 0, time.Local)
	if normalizedTime != expectedTime {
		t.Errorf("normalize time mismatch: expected %v, got %v", expectedTime, normalizedTime)
	}
}

func TestSubscription_NextExecution(t *testing.T) {
	mockPreferences := mockPreferences{}
	mockPreferences.preferences = Preference{SubscriptionRefreshHour: 10}

	now := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 15, 30, 0, 0, time.Local)
	time.Local = now.Location()

	tests := []struct {
		Name          string
		Sub           Subscription
		ExpectedTime  time.Time
		ExpectedError error
	}{
		{
			Name: "Refresh due",
			Sub: Subscription{
				RefreshFrequency: Day,
				Info:             SubscriptionInfo{LastCheck: now.Add(-time.Hour * 25)},
			},
			ExpectedTime:  time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, time.Local),
			ExpectedError: nil,
		},
		{
			Name: "Refresh not due",
			Sub: Subscription{
				RefreshFrequency: Day,
				Info:             SubscriptionInfo{LastCheck: now.Add(-time.Hour * 12)},
			},
			ExpectedTime:  time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, time.Local),
			ExpectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			next, err := tt.Sub.NextExecution(&mockPreferences)
			if next != tt.ExpectedTime {
				t.Errorf("NextExecution time mismatch: expected %v, got %v", tt.ExpectedTime, next)
			}
			if err != tt.ExpectedError {
				t.Errorf("NextExecution error mismatch: expected %v, got %v", tt.ExpectedError, err)
			}
		})
	}

	mockPreferences.err = errors.New("test error")
	sub := Subscription{
		RefreshFrequency: Day,
		Info:             SubscriptionInfo{LastCheck: now.Add(-time.Hour * 12)},
	}
	_, err := sub.NextExecution(&mockPreferences)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

type mockPreferences struct {
	preferences Preference
	err         error
}

func (m *mockPreferences) GetComplete() (*Preference, error) {
	return &m.preferences, m.err
}

func (m *mockPreferences) Update(pref Preference) error {
	m.preferences = pref
	return m.err
}

func (m *mockPreferences) Get() (*Preference, error) {
	return &m.preferences, m.err
}
