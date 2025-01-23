package utils

import (
	"sync"
	"testing"
)

func TestSafeMapNew(t *testing.T) {
	m1 := NewSafeMap[int, string]()
	if m1.Len() != 0 {
		t.Errorf("Expected length 0, got %d", m1.Len())
	}

	initial := map[int]string{1: "one", 2: "two"}
	m2 := NewSafeMap(initial)
	if m2.Len() != 2 {
		t.Errorf("Expected length 2, got %d", m2.Len())
	}
}

func TestSafeMapHas(t *testing.T) {
	m := NewSafeMap[int, string]()
	m.Set(1, "one")
	if !m.Has(1) {
		t.Errorf("Expected key 1 to exist")
	}
	if m.Has(2) {
		t.Errorf("Did not expect key 2 to exist")
	}
}

func TestSafeMapGet(t *testing.T) {
	m := NewSafeMap[int, string]()
	m.Set(1, "one")
	value, ok := m.Get(1)
	if !ok || value != "one" {
		t.Errorf("Expected to get 'one', got %v, %v", value, ok)
	}
	_, ok = m.Get(2)
	if ok {
		t.Errorf("Expected key 2 to not exist")
	}
}

func TestSafeMapSet(t *testing.T) {
	m := NewSafeMap[int, string]()
	m.Set(1, "one")
	value, ok := m.Get(1)
	if !ok || value != "one" {
		t.Errorf("Expected to get 'one', got %v, %v", value, ok)
	}
	m.Set(1, "uno")
	value, _ = m.Get(1)
	if value != "uno" {
		t.Errorf("Expected to get 'uno', got %v", value)
	}
}

func TestSafeMapDelete(t *testing.T) {
	m := NewSafeMap[int, string]()
	m.Set(1, "one")
	m.Delete(1)
	if m.Has(1) {
		t.Errorf("Expected key 1 to be deleted")
	}
}

func TestSafeMapLen(t *testing.T) {
	m := NewSafeMap[int, string]()
	if m.Len() != 0 {
		t.Errorf("Expected length 0, got %d", m.Len())
	}
	m.Set(1, "one")
	if m.Len() != 1 {
		t.Errorf("Expected length 1, got %d", m.Len())
	}
	m.Set(2, "two")
	if m.Len() != 2 {
		t.Errorf("Expected length 2, got %d", m.Len())
	}
}

func TestSafeMapForEachSafe(t *testing.T) {
	m := NewSafeMap[int, string]()
	m.Set(1, "one")
	m.Set(2, "two")

	keys := map[int]bool{}
	m.ForEachSafe(func(k int, v string) {
		keys[k] = true
	})
	if len(keys) != 2 || !keys[1] || !keys[2] {
		t.Errorf("Expected keys 1 and 2 to be iterated")
	}
}

func TestSafeMapForEach(t *testing.T) {
	m := NewSafeMap[int, string]()
	m.Set(1, "one")
	m.Set(2, "two")

	iterated := 0
	m.ForEach(func(k int, v string) {
		iterated++
		if k == 1 {
			m.Set(3, "three")
		}
	})
	if !m.Has(3) {
		t.Errorf("Expected key 3 to be added")
	}
}

func TestSafeMapConcurrency(t *testing.T) {
	m := NewSafeMap[int, int]()
	wg := sync.WaitGroup{}

	for i := range 100 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			m.Set(i, i)
		}(i)
	}

	wg.Wait()
	if m.Len() != 100 {
		t.Errorf("Expected length 100, got %d", m.Len())
	}

	for i := range 100 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			m.Delete(i)
		}(i)
	}

	wg.Wait()
	if m.Len() != 0 {
		t.Errorf("Expected length 0 after deletions, got %d", m.Len())
	}
}
