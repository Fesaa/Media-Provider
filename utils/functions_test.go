package utils

import (
	"testing"
)

func TestPercent(t *testing.T) {
	got := Percent(5, 10)
	if got != 50 {
		t.Errorf("Percent(5, 10) = %v, want 50", got)
	}

	got = Percent(5, 0)
	if got != 100 {
		t.Errorf("Percent(5, 0) = %v, want 5", got)
	}
}

func TestBytesToSize(t *testing.T) {
	got := BytesToSize(1024)

	if got != "1.00 KB" {
		t.Fatalf("BytesToSize(1024) = %q; want \"1.00 KB\"", got)
	}

	got = BytesToSize(873944456)

	if got != "833.46 MB" {
		t.Fatalf("BytesToSize(873944456) = %q; want \"833.46 MB\"", got)
	}
}

func TestOrDefault(t *testing.T) {
	got := OrDefault([]int{1, 2, 3}, 2)
	if got != 1 {
		t.Fatalf("OrDefault([]int{1, 2, 3}, 3) = %d; want 1", got)
	}

	got = OrDefault([]int{}, 3)
	if got != 3 {
		t.Fatalf("OrDefault([]int{}, 3) = %d; want 3", got)
	}
}

func TestTernary(t *testing.T) {
	got := Ternary(true, 1, 2)
	if got != 1 {
		t.Fatalf("Ternary(true, 1, 2) = %d; want 1", got)
	}

	got = Ternary(false, 1, 2)
	if got != 2 {
		t.Fatalf("Ternary(true, 1, 2) = %d; want 2", got)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Ternary(true, 1, 2, 3) should have panicked")
		}
	}()

	Ternary(true, 1, 2, 3)
}
