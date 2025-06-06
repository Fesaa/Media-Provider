package mangadex

import (
	"testing"
)

func TestChapterSearchData_ID(t *testing.T) {
	chapter := ChapterSearchData{Id: "test-id"}
	if chapter.GetId() != "test-id" {
		t.Errorf("Expected ID to be 'test-id', got '%s'", chapter.GetId())
	}
}

func TestChapterSearchData_Label(t *testing.T) {
	tests := []struct {
		name     string
		chapter  ChapterSearchData
		expected string
	}{
		{
			name: "One-shot",
			chapter: ChapterSearchData{
				Attributes: ChapterAttributes{
					Title: "One-shot title",
				},
			},
			expected: "One-shot title (OneShot)",
		},
		{
			name: "Chapter only",
			chapter: ChapterSearchData{
				Attributes: ChapterAttributes{
					Title:   "Chapter title",
					Chapter: "10",
				},
			},
			expected: "Chapter title (Ch. 10)",
		},
		{
			name: "Volume and chapter",
			chapter: ChapterSearchData{
				Attributes: ChapterAttributes{
					Title:   "Volume chapter title",
					Volume:  "5",
					Chapter: "20",
				},
			},
			expected: "Volume chapter title (Vol. 5 - Ch. 20)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.chapter.Label() != test.expected {
				t.Errorf("Expected label to be '%s', got '%s'", test.expected, test.chapter.Label())
			}
		})
	}
}

func TestChapterSearchData_Volume(t *testing.T) {
	tests := []struct {
		name     string
		chapter  ChapterSearchData
		expected float64
	}{
		{
			name: "Empty volume",
			chapter: ChapterSearchData{
				Attributes: ChapterAttributes{
					Volume: "",
				},
			},
			expected: -1,
		},
		{
			name: "Valid volume",
			chapter: ChapterSearchData{
				Attributes: ChapterAttributes{
					Volume: "10.5",
				},
			},
			expected: 10.5,
		},
		{
			name: "Invalid volume",
			chapter: ChapterSearchData{
				Attributes: ChapterAttributes{
					Volume: "abc",
				},
			},
			expected: -1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.chapter.Volume() != test.expected {
				t.Errorf("Expected volume to be %f, got %f", test.expected, test.chapter.Volume())
			}
		})
	}
}

func TestChapterSearchData_Chapter(t *testing.T) {
	tests := []struct {
		name     string
		chapter  ChapterSearchData
		expected float64
	}{
		{
			name: "Empty chapter",
			chapter: ChapterSearchData{
				Attributes: ChapterAttributes{
					Chapter: "",
				},
			},
			expected: -1,
		},
		{
			name: "Valid chapter",
			chapter: ChapterSearchData{
				Attributes: ChapterAttributes{
					Chapter: "20.5",
				},
			},
			expected: 20.5,
		},
		{
			name: "Invalid chapter",
			chapter: ChapterSearchData{
				Attributes: ChapterAttributes{
					Chapter: "xyz",
				},
			},
			expected: -1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.chapter.Chapter() != test.expected {
				t.Errorf("Expected chapter to be %f, got %f", test.expected, test.chapter.Chapter())
			}
		})
	}
}

func TestChapterSearchData_Volume_Chapter_Combined(t *testing.T) {
	tests := []struct {
		name            string
		chapter         ChapterSearchData
		expectedVol     float64
		expectedChapter float64
	}{
		{
			name: "Valid Volume and Chapter",
			chapter: ChapterSearchData{
				Attributes: ChapterAttributes{
					Volume:  "10.0",
					Chapter: "1.0",
				},
			},
			expectedVol:     10.0,
			expectedChapter: 1.0,
		},
		{
			name: "Invalid Volume and valid Chapter",
			chapter: ChapterSearchData{
				Attributes: ChapterAttributes{
					Volume:  "abc",
					Chapter: "1.0",
				},
			},
			expectedVol:     -1,
			expectedChapter: 1.0,
		},
		{
			name: "valid Volume and Invalid Chapter",
			chapter: ChapterSearchData{
				Attributes: ChapterAttributes{
					Volume:  "10.0",
					Chapter: "abc",
				},
			},
			expectedVol:     10.0,
			expectedChapter: -1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.chapter.Volume() != test.expectedVol {
				t.Errorf("Expected Volume to be %f, got %f", test.expectedVol, test.chapter.Volume())
			}
			if test.chapter.Chapter() != test.expectedChapter {
				t.Errorf("Expected Chapter to be %f, got %f", test.expectedChapter, test.chapter.Chapter())
			}
		})
	}
}
