package core

import (
	"github.com/Fesaa/Media-Provider/utils"
	"io"
	"path"
	"testing"
)

type testCase struct {
	name           string
	files          []string
	wantedChapters []ChapterMock
	isNew          []ChapterMock
}

func TestCore_loadContentOnDisk(t *testing.T) {

	const (
		MangaName = "Spice and Wolf"
	)

	file := func(s string) string {
		return path.Join(MangaName, MangaName+" "+s)
	}

	testCases := []testCase{
		{
			name:  "No content on disk",
			files: []string{},
			isNew: []ChapterMock{
				{
					Title: "Hi",
				},
			},
		},
		{
			name:  "All chapters on disk",
			files: []string{file("Ch. 001.cbz"), file("Ch. 002.cbz"), file("Ch. 003.cbz")},
			wantedChapters: []ChapterMock{
				{
					Chapter: "1",
				},
				{
					Chapter: "2",
				},
				{
					Chapter: "3",
				},
			},
			isNew: []ChapterMock{
				{
					Chapter: "4",
				},
			},
		},
		{
			name: "Volume and Chapter on disk",
			files: []string{
				file("Vol. 1 Ch. 001.cbz"), file("Vol. 1 Ch. 002.cbz"),
				file("Vol. 2 Ch. 003.cbz"), file("Vol. 2 Ch. 004.cbz"),
			},
			wantedChapters: []ChapterMock{
				SimpleChapter("", "1", "1"),
				SimpleChapter("", "2", "1"),
				SimpleChapter("", "3", "2"),
				SimpleChapter("", "4", "2"),
			},
			isNew: []ChapterMock{
				SimpleChapter("", "5", "3"),
				SimpleChapter("", "6", "3"),
			},
		},
		{
			name: "No volume on disk, match on chapter",
			files: []string{
				file("Ch. 001.cbz"), file("Ch. 002.cbz"),
				file("Ch. 003.cbz"), file("Ch. 004.cbz"),
			},
			wantedChapters: []ChapterMock{
				SimpleChapter("", "1", "1"),
				SimpleChapter("", "2", "1"),
				SimpleChapter("", "3", "2"),
				SimpleChapter("", "4", "2"),
			},
			isNew: []ChapterMock{
				SimpleChapter("", "5", "3"),
				SimpleChapter("", "6", "3"),
			},
		},
		{
			name: "Mixed on disk",
			files: []string{
				file("Vol. 1 Ch. 001.cbz"), file("Vol. 1 Ch. 002.cbz"),
				file("Ch. 003.cbz"), file("Ch. 004.cbz"),
			},
			wantedChapters: []ChapterMock{
				SimpleChapter("", "1", "1"),
				SimpleChapter("", "2", "1"),
				SimpleChapter("", "3", "1"),
				SimpleChapter("", "4", "1"),
			},
			isNew: []ChapterMock{
				SimpleChapter("", "5", "2"),
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			core := testBase(t, req(), io.Discard, ProviderMock{})

			write := func(s string) {
				utils.Must(core.fs.WriteFile(s, []byte{}, 0))
			}
			utils.ForEach(tt.files, write)

			core.SeriesInfo = &SeriesMock{
				chapters: append(tt.wantedChapters, tt.isNew...),
			}

			// Load content
			core.loadContentOnDisk()

			for _, wanted := range tt.wantedChapters {
				if core.ShouldDownload(wanted) {
					t.Errorf("Did not find chapter %v", wanted)
				}
			}

			for _, notWanted := range tt.isNew {
				if !core.ShouldDownload(notWanted) {
					t.Errorf("Chapter %v should have been downloaded", notWanted)
				}
			}
		})
	}

}
