package core

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"io"
	"path/filepath"
	"testing"
)

// SimpleChapter returns a chapter mock, args are
// chapter - volume - title - label
func SimpleChapter(id string, args ...string) ChapterMock {
	return ChapterMock{
		Id:       id,
		Chapter:  utils.AtIdx(args, 0),
		Volume:   utils.AtIdx(args, 1),
		Title:    utils.AtIdx(args, 2),
		LabelStr: utils.AtIdx(args, 3),
	}
}

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
			expected: filepath.Join(baseDir, "Spice and Wolf", "Spice and Wolf Christmas Special (One Shot)"),
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
			want: "Spice and Wolf Oneshot Title (One Shot)",
		},
		{
			name: "Numeric Chapter",
			chapter: ChapterMock{
				Chapter: "12.5",
				Title:   "Something",
			},
			want: "Spice and Wolf Ch. 0012.5",
		},
		{
			name: "Non-Numeric Chapter",
			chapter: ChapterMock{
				Chapter: "extra-a",
				Title:   "Extra",
			},
			want: "Spice and Wolf Ch. extra-a",
		},
		{
			name: "Empty Chapter String",
			chapter: ChapterMock{
				Chapter: "",
				Title:   "Bonus",
			},
			want: "Spice and Wolf Bonus (One Shot)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := testBase(t, req(), io.Discard, ProviderMock{
				title: "Spice and Wolf",
			})
			if got := core.ContentDir(tt.chapter); got != tt.want {
				t.Errorf("ContentDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCore_IsContent(t *testing.T) {
	type testCase[T Chapter] struct {
		name     string
		diskName string
		want     bool
	}
	tests := []testCase[ChapterMock]{
		{
			name:     "Valid Chapter Format",
			diskName: "My Manga Ch. 0012.cbz",
			want:     true,
		},
		{
			name:     "Valid Volume Format",
			diskName: "My Manga Vol. 05.cbz",
			want:     true,
		},
		{
			name:     "Valid OneShot Format (new)",
			diskName: "My Manga Oneshot Title (OneShot).cbz",
			want:     true,
		},
		{
			name:     "Valid OneShot Format (old)",
			diskName: "My Manga OneShot Oneshot Title.cbz",
			want:     true,
		},
		{
			name:     "Invalid Format - no match",
			diskName: "Random_File_Name.zip",
			want:     false,
		},
		{
			name:     "Invalid Format - wrong extension",
			diskName: "My Manga Ch. 0012.pdf",
			want:     false,
		},
		{
			name:     "Valid format with Volume",
			diskName: "My Manga Vol. 5 Ch. 0007.cbz",
			want:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := testBase(t, req(), io.Discard, ProviderMock{})

			if got := core.IsContent(tt.diskName); got != tt.want {
				t.Errorf("IsContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCore_ContentKey(t *testing.T) {
	core := testBase(t, req(), io.Discard, ProviderMock{})

	want := "Spice and Wolf"
	got := core.ContentKey(ChapterMock{
		Id: "Spice and Wolf",
	})

	if got != want {
		t.Errorf("ContentKey() = %v, want %v", got, want)
	}

}
