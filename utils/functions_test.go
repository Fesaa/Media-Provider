package utils

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestPercent(t *testing.T) {
	type testCase struct {
		name     string
		inputA   int64
		inputB   int64
		expected int64
	}

	testCases := []testCase{
		{
			name:     "Normal",
			inputA:   5,
			inputB:   10,
			expected: 50,
		},
		{
			name:     "BZero",
			inputA:   5,
			inputB:   0,
			expected: 100,
		},
		{
			name:     "AllZero",
			inputA:   0,
			inputB:   0,
			expected: 0,
		},
		{
			name:     "A > B",
			inputA:   5,
			inputB:   2,
			expected: 100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Percent(tc.inputA, tc.inputB)
			if got != tc.expected {
				t.Errorf("got %d, want %d", got, tc.expected)
			}
		})
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

	got = BytesToSize(0)
	if got != "0 Byte" {
		t.Fatalf("BytesToSize(0) = %q; want \"0 Byte\"", got)
	}
}

func TestStringify(t *testing.T) {
	if Stringify(1) != "1" {
		t.Errorf("Stringify(1) = %v; want \"1\"", Stringify(1))
	}
}

func TestIdentity(t *testing.T) {
	f := Identity('1')

	if f() != '1' {
		t.Errorf("Identity('1') = %v; want \"1\"", f())
	}
}

func TestMustNoErr(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Must() did panic")
		}
	}()

	Must(nil)
}

func TestMustErr(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Must() did not panic")
		}
	}()

	Must(errors.New("foo"))
}

func TestMustReturnNoErr(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Must() did panic")
		}
	}()

	MustReturn(func() (string, error) {
		return "foo", nil
	}())
}

func TestMustReturnErr(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Must() did not panic")
		}
	}()
	MustReturn(func() (string, error) {
		return "", errors.New("foo")
	}())

}

func TestMustHave(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Must() did panic")
		}
	}()
	MustHave(func() (string, bool) {
		return "", true
	}())
}

func TestMustHavePanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Must() did not panic")
		}
	}()
	MustHave(func() (string, bool) {
		return "", false
	}())
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

func TestHumanReadableSpeed(t *testing.T) {
	type args struct {
		s int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Small",
			args: args{
				s: 1000,
			},
			want: "1000.00 B/s",
		},
		{
			name: "Medium",
			args: args{
				s: 2000,
			},
			want: "1.95 KB/s",
		},
		{
			name: "Large",
			args: args{
				s: 3456789,
			},
			want: "3.30 MB/s",
		},
		{
			name: "Zero",
			args: args{
				s: 0,
			},
			want: "0.00 B/s",
		},
		{
			name: "Negative",
			args: args{
				s: -1,
			},
			want: "-1.00 B/s",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HumanReadableSpeed(tt.args.s); got != tt.want {
				t.Errorf("HumanReadableSpeed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateSecret(t *testing.T) {
	_, err := GenerateSecret(1024)
	if err != nil {
		t.Errorf("GenerateSecret(1024) = %v; want nil", err)
	}
}

func TestGenerateApiKey(t *testing.T) {
	_, err := GenerateApiKey()
	if err != nil {
		t.Errorf("GenerateApiKey() = %v; want nil", err)
	}
}

func TestGenerateSecretNegative(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("GenerateSecret(-1) = %v; want panic", r)
		}
	}()
	_, _ = GenerateSecret(-1)
}

func TestSortFloats(t *testing.T) {
	type args struct {
		a string
		b string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Equal",
			args: args{
				a: "1",
				b: "1",
			},
			want: 0,
		},
		{
			name: "Less",
			args: args{
				a: "1",
				b: "2",
			},
			want: 1,
		},
		{
			name: "More",
			args: args{
				a: "2",
				b: "1",
			},
			want: -1,
		},
		{
			name: "Empty a",
			args: args{
				a: "",
				b: "1",
			},
			want: 1,
		},
		{
			name: "Empty b",
			args: args{
				a: "1",
				b: "",
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SortFloats(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("SortFloats() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShorten(t *testing.T) {
	type args struct {
		s      string
		length int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no-op",
			args: args{
				s:      "foo",
				length: 10,
			},
			want: "foo",
		},
		{
			name: "shorten",
			args: args{
				s:      "foo",
				length: 2,
			},
			want: "fo",
		},
		{
			name: "equal",
			args: args{
				s:      "foo",
				length: 3,
			},
			want: "foo",
		},
		{
			name: "long",
			args: args{
				s:      "thisisaverylongsentence",
				length: 10,
			},
			want: "thisisa...",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Shorten(tt.args.s, tt.args.length); got != tt.want {
				t.Errorf("Shorten() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTryCatch(t *testing.T) {
	type args[T any, U any] struct {
		producer      func() (T, error)
		mapper        func(T) U
		fallback      U
		errorHandlers []func(error)
	}
	type testCase[T any, U any] struct {
		name string
		args args[T, U]
		want U
	}
	tests := []testCase[string, int]{
		{
			name: "success",
			args: args[string, int]{
				producer: func() (string, error) {
					return "foo", nil
				},
				mapper: func(s string) int {
					return len(s)
				},
				fallback:      2,
				errorHandlers: nil,
			},
			want: 3,
		},
		{
			name: "error",
			args: args[string, int]{
				producer: func() (string, error) {
					return "foo", fmt.Errorf("foo")
				},
				mapper: func(s string) int {
					return len(s)
				},
				fallback: 2,
				errorHandlers: []func(error){
					func(e error) {

					},
				},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TryCatch(tt.args.producer, tt.args.mapper, tt.args.fallback, tt.args.errorHandlers...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TryCatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
