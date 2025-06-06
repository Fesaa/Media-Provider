package core

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"io"
	"path/filepath"
	"testing"
)

type ChapterMock struct {
	Id       string
	LabelStr string
	Chapter  string
	Volume   string
	Title    string
}

func (c ChapterMock) GetId() string {
	return c.Id
}

func (c ChapterMock) Label() string {
	return c.LabelStr
}

func (c ChapterMock) GetChapter() string {
	return c.Chapter
}

func (c ChapterMock) GetVolume() string {
	return c.Volume
}

func (c ChapterMock) GetTitle() string {
	return c.Title
}

func TestCore_ContentPath(t *testing.T) {
	type testCase[T Chapter] struct {
		name     string
		req      payload.DownloadRequest
		chapter  ChapterMock
		expected string
	}

	tmpWriter := io.Discard
	baseDir := t.TempDir()

	tests := []testCase[ChapterMock]{
		{
			name: "With volume",
			req: payload.DownloadRequest{
				TempTitle: "ExampleTitle",
				Provider:  models.MANGADEX,
				BaseDir:   baseDir,
			},
			chapter: ChapterMock{
				Volume:  "5",
				Chapter: "1",
				Title:   "Chapter 42",
			},
			expected: filepath.Join(baseDir, "ExampleTitle", "ExampleTitle Vol. 5", "ExampleTitle Ch. 0001"),
		},
		{
			name: "Without volume (OneShot)",
			req: payload.DownloadRequest{
				TempTitle: "Spice and Wolf",
				Provider:  models.MANGADEX,
				BaseDir:   baseDir,
			},
			chapter: ChapterMock{
				Title: "Christmas Special",
			},
			expected: filepath.Join(baseDir, "Spice and Wolf", "Spice and Wolf Christmas Special (OneShot)"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := testBase(t, tt.req, tmpWriter, ProviderMock{
				title:      tt.req.TempTitle,
				contentDir: tt.chapter.Title,
			})
			got := c.ContentPath(tt.chapter)
			if got != tt.expected {
				t.Errorf("ContentPath() = %v,\n want %v", got, tt.expected)
			}
		})
	}
}

func TestCore_ContentDir(t *testing.T) {
	type testCase[T Chapter] struct {
		name    string
		chapter ChapterMock
		want    string
	}
	tests := []testCase[ChapterMock]{
		{
			name: "OneShot Chapter",
			chapter: ChapterMock{
				Chapter: "",
				Title:   "Oneshot Title",
			},
			want: "My Manga Oneshot Title (OneShot)",
		},
		{
			name: "Numeric Chapter",
			chapter: ChapterMock{
				Chapter: "12.5",
				Title:   "Something",
			},
			want: "My Manga Ch. 0012.5",
		},
		{
			name: "Non-Numeric Chapter",
			chapter: ChapterMock{
				Chapter: "extra-a",
				Title:   "Extra",
			},
			want: "My Manga Ch. extra-a",
		},
		{
			name: "Empty Chapter String",
			chapter: ChapterMock{
				Chapter: "",
				Title:   "Bonus",
			},
			want: "My Manga Bonus (OneShot)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := testBase(t, req(), io.Discard, ProviderMock{
				title: "My Manga",
			})
			if got := core.ContentDir(tt.chapter); got != tt.want {
				t.Errorf("ContentDir() = %v, want %v", got, tt.want)
			}
		})
	}
}
