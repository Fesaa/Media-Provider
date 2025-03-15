package models

import (
	"testing"
)

func TestProvider_String(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
		expected string
	}{
		{
			name:     "NYAA",
			provider: NYAA,
			expected: "Nyaa",
		},
		{
			name:     "YTS",
			provider: YTS,
			expected: "YTS",
		},
		{
			name:     "LIME",
			provider: LIME,
			expected: "Lime",
		},
		{
			name:     "SUBSPLEASE",
			provider: SUBSPLEASE,
			expected: "SubsPlease",
		},
		{
			name:     "MANGADEX",
			provider: MANGADEX,
			expected: "MangaDex",
		},
		{
			name:     "WEBTOON",
			provider: WEBTOON,
			expected: "Webtoon",
		},
		{
			name:     "DYNASTY",
			provider: DYNASTY,
			expected: "Dynasty",
		},
		{
			name:     "Unknown Provider",
			provider: Provider(100), // An unknown provider
			expected: "Unknown Provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.provider.String()
			if result != tt.expected {
				t.Errorf("String() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
