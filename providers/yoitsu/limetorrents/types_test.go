package limetorrents

import "testing"

func TestConvertCategory(t *testing.T) {
	type testCase struct {
		input    string
		expected Category
	}

	cases := []testCase{
		{
			input:    "",
			expected: ALL,
		},
		{
			input:    "aNimE",
			expected: ANIME,
		},
		{
			input:    "movies",
			expected: MOVIE,
		},
		{
			input:    "applications",
			expected: APPS,
		},
		{
			input:    "games",
			expected: GAMES,
		},
		{
			input:    "music",
			expected: MUSIC,
		},
		{
			input:    "tv",
			expected: TV,
		},
		{
			input:    "other",
			expected: OTHER,
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			if got := ConvertCategory(c.input); got != c.expected {
				t.Errorf("ConvertCategory(%q): got %q, want %q", c.input, got, c.expected)
			}
		})
	}
}
