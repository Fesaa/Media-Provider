package metadata

import (
	"github.com/Fesaa/Media-Provider/config"
	"testing"
)

func TestMPVersion_String(t *testing.T) {
	v := config.SemanticVersion("1.2.3")
	if v.String() != "1.2.3" {
		t.Errorf("String() = %v, want %v", v.String(), "1.2.3")
	}
}

func TestMPVersion_Older(t *testing.T) {
	tests := []struct {
		v1   config.SemanticVersion
		v2   config.SemanticVersion
		want bool
	}{
		{v1: "1.2.3", v2: "1.2.4", want: true},
		{v1: "1.2.4", v2: "1.2.3", want: false},
		{v1: "1.2.3", v2: "1.2.3", want: false},
		{v1: "1.10.0", v2: "1.2.3", want: false},
		{v1: "1.2.3", v2: "1.10.0", want: true},
	}

	for _, tt := range tests {
		if got := tt.v1.Older(tt.v2); got != tt.want {
			t.Errorf("Older(%v, %v) = %v, want %v", tt.v1, tt.v2, got, tt.want)
		}
	}
}

func TestMPVersion_Newer(t *testing.T) {
	tests := []struct {
		v1   config.SemanticVersion
		v2   config.SemanticVersion
		want bool
	}{
		{v1: "1.2.3", v2: "1.2.4", want: false},
		{v1: "1.2.4", v2: "1.2.3", want: true},
		{v1: "1.2.3", v2: "1.2.3", want: false},
		{v1: "1.10.0", v2: "1.2.3", want: true},
		{v1: "1.2.3", v2: "1.10.0", want: false},
	}

	for _, tt := range tests {
		if got := tt.v1.Newer(tt.v2); got != tt.want {
			t.Errorf("Newer(%v, %v) = %v, want %v", tt.v1, tt.v2, got, tt.want)
		}
	}
}

func TestMPVersion_Equal(t *testing.T) {
	tests := []struct {
		v1   config.SemanticVersion
		v2   config.SemanticVersion
		want bool
	}{
		{v1: "1.2.3", v2: "1.2.3", want: true},
		{v1: "1.2.3", v2: "1.2.4", want: false},
	}

	for _, tt := range tests {
		if got := tt.v1.Equal(tt.v2); got != tt.want {
			t.Errorf("Equal(%v, %v) = %v, want %v", tt.v1, tt.v2, got, tt.want)
		}
	}
}

func TestMPVersion_EqualS(t *testing.T) {
	tests := []struct {
		v1   config.SemanticVersion
		v2   string
		want bool
	}{
		{v1: "1.2.3", v2: "1.2.3", want: true},
		{v1: "1.2.3", v2: "1.2.4", want: false},
	}

	for _, tt := range tests {
		if got := tt.v1.EqualS(tt.v2); got != tt.want {
			t.Errorf("EqualS(%v, %v) = %v, want %v", tt.v1, tt.v2, got, tt.want)
		}
	}
}
