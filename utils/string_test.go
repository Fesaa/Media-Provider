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

func TestPadFloat(t *testing.T) {
	type args struct {
		f float64
		n int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "0 pad",
			args: args{
				f: 0,
				n: 3,
			},
			want: "000",
		},
		{
			name: "Correct length",
			args: args{
				f: 123,
				n: 3,
			},
			want: "123",
		},
		{
			name: "Too long",
			args: args{
				f: 1234,
				n: 3,
			},
			want: "1234",
		},
		{
			name: "Too short",
			args: args{
				f: 12,
				n: 3,
			},
			want: "012",
		},
		{
			name: "Decimal",
			args: args{
				f: 123.456,
				n: 4,
			},
			want: "0123.5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PadFloat(tt.args.f, tt.args.n); got != tt.want {
				t.Errorf("PadFloat() = %v, want %v", got, tt.want)
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
