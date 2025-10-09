package utils

import "testing"

func TestProgress(t *testing.T) {
	tracker := NewSpeedTracker(10)

	// 0% progress
	if progress := tracker.Progress(); progress != 0 {
		t.Errorf("Expected 0%%, got %f%%", progress)
	}

	// 10% progress
	tracker.Increment()
	if progress := tracker.Progress(); progress != 10 {
		t.Errorf("Expected 10%%, got %f%%", progress)
	}

	// 50% progress
	for i := 0; i < 4; i++ {
		tracker.Increment()
	}
	if progress := tracker.Progress(); progress != 50 {
		t.Errorf("Expected 50%%, got %f%%", progress)
	}

	// 100% progress
	for i := 0; i < 5; i++ {
		tracker.Increment()
	}
	if progress := tracker.Progress(); progress != 100 {
		t.Errorf("Expected 100%%, got %f%%", progress)
	}
}

func TestProgressWithIntermediate(t *testing.T) {
	tracker := NewSpeedTracker(10)

	// Set intermediate tracker with 100 items
	tracker.SetIntermediate(100)

	// With no main progress and no intermediate progress: 0%
	if progress := tracker.Progress(); progress != 0 {
		t.Errorf("Expected 0%%, got %f%%", progress)
	}

	// Increment intermediate by 50 (50% of intermediate)
	for i := 0; i < 50; i++ {
		tracker.IncrementIntermediate()
	}
	// Progress should be 50% of 1 item out of 10 = 5%
	expected := 5.0
	if progress := tracker.Progress(); progress != expected {
		t.Errorf("Expected %f%%, got %f%%", expected, progress)
	}

	// Complete intermediate (100/100)
	for i := 0; i < 50; i++ {
		tracker.IncrementIntermediate()
	}
	// Progress should be 100% of 1 item out of 10 = 10%
	expected = 10.0
	if progress := tracker.Progress(); progress != expected {
		t.Errorf("Expected %f%%, got %f%%", expected, progress)
	}
}

func TestProgressNeverDecreases(t *testing.T) {
	tracker := NewSpeedTracker(52)

	var lastProgress float64 = 0

	// Complete first item
	tracker.SetIntermediate(100)

	// Increment intermediate to 81.40%
	for i := 0; i < 81; i++ {
		tracker.IncrementIntermediate()
	}

	progress1 := tracker.Progress()
	if progress1 <= lastProgress {
		t.Errorf("Progress decreased: %f -> %f", lastProgress, progress1)
	}
	lastProgress = progress1

	// Now complete the item: increment main counter and reset intermediate
	tracker.Increment()
	tracker.ClearIntermediate()

	progress2 := tracker.Progress()
	if progress2 < lastProgress {
		t.Errorf("Progress decreased after increment: %f -> %f", lastProgress, progress2)
	}
	lastProgress = progress2

	// Start next item
	tracker.SetIntermediate(100)

	progress3 := tracker.Progress()
	if progress3 < lastProgress {
		t.Errorf("Progress decreased after setting intermediate: %f -> %f", lastProgress, progress3)
	}
}

func TestProgressMonotonicallyIncreasing(t *testing.T) {
	tracker := NewSpeedTracker(10)

	var lastProgress float64 = 0

	for item := 0; item < 10; item++ {
		tracker.SetIntermediate(50)

		for intermediate := 0; intermediate < 50; intermediate++ {
			tracker.IncrementIntermediate()

			progress := tracker.Progress()
			if progress < lastProgress {
				t.Errorf("Progress decreased at item=%d, intermediate=%d: %f -> %f",
					item, intermediate, lastProgress, progress)
			}
			lastProgress = progress
		}

		tracker.Increment()
		tracker.ClearIntermediate()

		progress := tracker.Progress()
		if progress < lastProgress {
			t.Errorf("Progress decreased after completing item %d: %f -> %f",
				item, lastProgress, progress)
		}
		lastProgress = progress
	}

	// Final progress should be 100%
	if lastProgress != 100 {
		t.Errorf("Expected final progress=100%%, got %f%%", lastProgress)
	}
}
