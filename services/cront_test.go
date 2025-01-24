package services

import (
	"github.com/rs/zerolog"
	"testing"
)

func TestCronServiceProvider(t *testing.T) {
	_, err := CronServiceProvider(zerolog.Logger{})
	if err != nil {
		t.Error(err)
	}
}
