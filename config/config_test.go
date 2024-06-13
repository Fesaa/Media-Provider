package config

import (
	"fmt"
	"testing"
)

func TestConfig(t *testing.T) {
	if err := LoadConfig("../config.yaml.example"); err != nil {
		t.Fatal(err)
	}

	if I().GetRootURl() != "behind-me" {
		t.Fatal(fmt.Errorf("wrong root url: %s wanted behind-me", I().GetRootURl()))
	}

	if len(I().GetPages()) != 4 {
		t.Fatal(fmt.Errorf("wrong number of pages: %d wanted 4", I().GetPages()))
	}
}
