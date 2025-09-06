package utils

import (
	"slices"
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

func Test_safeMap_Keys(t *testing.T) {
	m := NewSafeMap[int, string]()

	got := len(m.Keys())
	want := 0
	if got != want {
		t.Errorf("Expected length 0, got %d", got)
	}

	m.Set(1, "one")
	got = len(m.Keys())
	want = 1
	if got != want {
		t.Errorf("Expected length 1, got %d", got)
	}

	m.Set(2, "two")

	got2 := m.Keys()
	want2 := []int{1, 2}
	for _, key := range want2 {
		if !slices.Contains(got2, key) {
			t.Errorf("Expected key %d to exist", key)
		}
	}
}

func Test_safeMap_Values(t *testing.T) {
	m := NewSafeMap[int, string]()

	got := len(m.Values())
	want := 0
	if got != want {
		t.Errorf("Expected length 0, got %d", got)
	}

	m.Set(1, "one")
	got = len(m.Values())
	want = 1
	if got != want {
		t.Errorf("Expected length 1, got %d", got)
	}

	m.Set(2, "two")
	got2 := m.Values()
	want2 := []string{"two", "one"}
	for _, value := range want2 {
		if !slices.Contains(got2, value) {
			t.Errorf("Expected key %s to exist", value)
		}
	}
}

func Test_safeMap_Any(t *testing.T) {
	type args[K comparable, V any] struct {
		f func(k K, v V) bool
	}
	type testCase[K comparable, V any] struct {
		name string
		s    SafeMap[K, V]
		args args[K, V]
		want bool
	}

	s := NewSafeMap[int, string]()
	s.Set(1, "one")
	s.Set(2, "two")
	s.Set(3, "three")

	tests := []testCase[int, string]{
		{
			name: "no match",
			s:    s,
			args: args[int, string]{
				f: func(k int, v string) bool {
					return k == 2 && v == "one"
				},
			},
			want: false,
		},
		{
			name: "match",
			s:    s,
			args: args[int, string]{
				f: func(k int, v string) bool {
					return k == 2 && v == "two"
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Any(tt.args.f); got != tt.want {
				t.Errorf("Any() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_safeMap_Count(t *testing.T) {
	type args[K comparable, V any] struct {
		f func(K, V) bool
	}
	type testCase[K comparable, V any] struct {
		name string
		s    SafeMap[K, V]
		args args[K, V]
		want int
	}

	s := NewSafeMap[int, string]()
	s.Set(1, "one")
	s.Set(2, "two")
	s.Set(3, "three")

	tests := []testCase[int, string]{
		{
			name: "no match",
			s:    s,
			args: args[int, string]{
				f: func(k int, v string) bool {
					return k == 2 && v == "one"
				},
			},
			want: 0,
		},
		{
			name: "match",
			s:    s,
			args: args[int, string]{
				f: func(k int, v string) bool {
					return k > 1
				},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Count(tt.args.f); got != tt.want {
				t.Errorf("Count() = %v, want %v", got, tt.want)
			}
		})
	}
}

//nolint:funlen
func Test_safeMap_Find(t *testing.T) {
	type args[K comparable, V any] struct {
		f func(k K, v V) bool
	}
	type testCase[K comparable, V any] struct {
		name  string
		s     SafeMap[K, V]
		args  args[K, V]
		want  *V
		want1 bool
	}

	s := NewSafeMap[int, string]()
	one := "one"
	s.Set(1, one)
	s.Set(2, "two")
	s.Set(3, "three")

	tests := []testCase[int, string]{
		{
			name: "no match",
			s:    s,
			args: args[int, string]{
				f: func(k int, v string) bool {
					return k == 2 && v == "one"
				},
			},
			want:  nil,
			want1: false,
		},
		{
			name: "match",
			s:    s,
			args: args[int, string]{
				f: func(k int, v string) bool {
					return k == 1
				},
			},
			want:  &one,
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.s.Find(tt.args.f)
			if got != nil && tt.want != nil {
				if *got != *tt.want {
					t.Errorf("Find() got = %v, want %v", got, tt.want)
				}
			}

			if got != nil && tt.want == nil {
				t.Errorf("Find() got = %v, want %v", got, tt.want)
			}

			if got == nil && tt.want != nil {
				t.Errorf("Find() got = %v, want %v", got, tt.want)
			}

			if got1 != tt.want1 {
				t.Errorf("Find() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
