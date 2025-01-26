package services

import (
	"github.com/rs/zerolog"
	"testing"
)

func TestCronServiceProvider(t *testing.T) {
	_, err := CronServiceProvider(zerolog.Nop())
	if err != nil {
		t.Error(err)
	}
}
