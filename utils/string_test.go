package utils

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no-op",
			input:    "abdsfiqkl",
			expected: "abdsfiqkl",
		},
		{
			name:     "Complex",
			input:    "5ç!%%ab",
			expected: "5ab",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := Normalize(test.input)
			if actual != test.expected {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestPadInt(t *testing.T) {
	type args struct {
		i int
		n int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Too long",
			args: args{
				i: 1234,
				n: 3,
			},
			want: "1234",
		},
		{
			name: "Too short",
			args: args{
				i: 12,
				n: 4,
			},
			want: "0012",
		},
		{
			name: "Correct length",
			args: args{
				i: 123,
				n: 3,
			},
			want: "123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PadInt(tt.args.i, tt.args.n); got != tt.want {
				t.Errorf("PadInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_pad(t *testing.T) {
	type args struct {
		str string
		n   int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Correct length",
			args: args{
				str: "123",
				n:   3,
			},
			want: "123",
		},
		{
			name: "Too long",
			args: args{
				str: "1234",
				n:   3,
			},
			want: "1234",
		},
		{
			name: "Too short",
			args: args{
				str: "12",
				n:   3,
			},
			want: "012",
		},
		{
			name: "Non number",
			args: args{
				str: "abc",
				n:   5,
			},
			want: "00abc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pad(tt.args.str, tt.args.n); got != tt.want {
				t.Errorf("pad() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPadFloatFromString(t *testing.T) {
	testCases := []struct {
		input    string
		length   int
		expected string
	}{
		{
			input:    "1.2",
			length:   4,
			expected: "0001.2",
		},
		{
			input:    "1.20",
			length:   4,
			expected: "0001.20",
		},
		{
			input:    "123",
			length:   6,
			expected: "000123",
		},
		{
			input:    "1.00",
			length:   3,
			expected: "001.00",
		},
		{
			input:    "12345.6789",
			length:   7,
			expected: "0012345.6789",
		},
		{
			input:    "0.1",
			length:   3,
			expected: "000.1",
		},
		{
			input:    "10",
			length:   4,
			expected: "0010",
		},
		{
			input:    "1.02",
			length:   4,
			expected: "0001.02",
		},
	}

	for _, tc := range testCases {
		actual := PadFloatFromString(tc.input, tc.length)
		if actual != tc.expected {
			t.Errorf("PadFloatFromString(%q, %d) = %q, expected %q", tc.input, tc.length, actual, tc.expected)
		}
	}
}
