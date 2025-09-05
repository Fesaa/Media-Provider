package services

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestCronServiceProvider(t *testing.T) {
	_, err := CronServiceProvider(zerolog.Nop())
	if err != nil {
		t.Error(err)
	}
}
