package utils

import (
	"reflect"
	"slices"
	"strconv"
	"testing"
)

func TestMap(t *testing.T) {

	got := Map([]bool{true, false}, func(b bool) bool {
		return !b
	})

	if !slices.Equal(got, []bool{false, true}) {
		t.Fatalf("got %v; want %v", got, []bool{false, true})
	}

	got2 := Map([]string{"", "abc"}, func(t string) int {
		return len(t)
	})

	if !slices.Equal(got2, []int{0, 3}) {
		t.Fatalf("got %v; want %v", got2, []int{0, 3})
	}

}

func TestFlatMapMany(t *testing.T) {
	got := FlatMapMany([]bool{true, false, false}, []bool{true, false, true})
	want := []bool{true, false, false, true, false, true}

	if !slices.Equal(got, want) {
		t.Fatalf("got %v; want %v", got, want)
	}
}

func TestFlatMap(t *testing.T) {
	got := FlatMap([][]string{
		{"a", "b"},
		{"c", "d"},
	})

	want := []string{"a", "b", "c", "d"}

	if !slices.Equal(got, want) {
		t.Fatalf("got %v; want %v", got, want)
	}
}

func TestMaybeMap(t *testing.T) {
	got := MaybeMap([]string{"", "abc", "3"}, func(t string) (int, bool) {
		l := len(t)
		if l == 0 {
			return 0, false
		}
		return l, true
	})

	want := []int{3, 1}

	if !slices.Equal(got, want) {
		t.Fatalf("got %v; want %v", got, want)
	}
}

func TestFilter(t *testing.T) {
	got := Filter([]string{"", "abc"}, func(s string) bool {
		return len(s) > 0
	})

	want := []string{"abc"}

	if !slices.Equal(got, want) {
		t.Fatalf("got %v; want %v", got, want)
	}
}

func TestFind(t *testing.T) {
	got := Find([]string{"", "abc", "ab"}, func(s string) bool {
		return len(s) > 0
	})

	if got == nil {
		t.Fatalf("got %v; want %v", got, nil)
	}

	if *got != "abc" {
		t.Fatalf("got %v; want %v", got, "abc")
	}

	got = Find([]string{"abcd", "abc"}, func(s string) bool {
		return len(s) == 0
	})

	if got != nil {
		t.Fatalf("got %v; want %v", got, nil)
	}
}

func TestAny(t *testing.T) {
	got := Any([]bool{true, false}, func(b bool) bool {
		return b
	})

	if !got {
		t.Fatalf("got %v; want %v", got, true)
	}

	got = Any([]bool{false, false}, func(b bool) bool {
		return b
	})

	if got {
		t.Fatalf("got %v; want %v", got, false)
	}

	got = Any([]string{"", "abc"}, func(s string) bool {
		return len(s) > 0
	})

	if !got {
		t.Fatalf("got %v; want %v", got, true)
	}
}

func TestKeys(t *testing.T) {
	got := Keys(map[string]bool{"a": true, "b": true, "c": true})
	want := []string{"a", "b", "c"}

	for _, w := range want {
		if !slices.Contains(got, w) {
			t.Fatalf("got %v; want %v", got, want)
		}
	}

	got = Keys(map[string]bool{})
	if len(got) != 0 {
		t.Fatalf("got %v; want %v", got, want)
	}
}

func TestValues(t *testing.T) {
	got := Values(map[string]bool{"a": true, "b": true, "c": false})
	want := []bool{true, true, false}

	for _, w := range want {
		if !slices.Contains(got, w) {
			t.Fatalf("got %v; want %v", got, want)
		}
	}

	got = Values(map[string]bool{})
	if len(got) != 0 {
		t.Fatalf("got %v; want %v", got, want)
	}
}

func TestMapKeys(t *testing.T) {
	got := MapKeys(map[string]bool{"a": true, "ab": true, "abc": false}, func(k string) int {
		return len(k)
	})
	want := []int{1, 2, 3}

	for _, w := range want {
		if !slices.Contains(got, w) {
			t.Fatalf("got %v; want %v", got, want)
		}
	}

	got = MapKeys(map[string]bool{}, func(k string) int {
		panic("SHOULD NOT HAPPEN")
	})

	if len(got) != 0 {
		t.Fatalf("got %v; want %v", got, want)
	}
}

func TestMapValues(t *testing.T) {
	got := MapValues(map[string]int{"a": 1, "b": 2, "c": 3}, strconv.Itoa)

	want := []string{"1", "2", "3"}

	for _, w := range want {
		if !slices.Contains(got, w) {
			t.Fatalf("got %v; want %v", got, want)
		}
	}

	got = MapValues(map[string]int{}, func(v int) string {
		panic("SHOULD NOT HAPPEN")
	})

	if len(got) != 0 {
		t.Fatalf("got %v; want %v", got, want)
	}
}

func TestGroupBy(t *testing.T) {
	type testType struct {
		name string
		in   []string
		want map[int][]string
		f    func(string) int
	}

	tests := []testType{
		{
			name: "empty",
			in:   []string{},
			want: map[int][]string{},
			f:    func(s string) int { return len(s) },
		},
		{
			name: "single",
			in:   []string{"a"},
			want: map[int][]string{1: {"a"}},
			f:    func(s string) int { return len(s) },
		},
		{
			name: "multiple",
			in:   []string{"a", "b", "c", "ab", "abc"},
			want: map[int][]string{1: {"a", "b", "c"}, 2: {"ab"}, 3: {"abc"}},
			f:    func(s string) int { return len(s) },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := GroupBy(test.in, test.f)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("got %v; want %v", got, test.want)
			}
		})
	}
}
