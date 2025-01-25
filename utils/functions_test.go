package utils

import (
	"errors"
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
			expected: 100,
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
