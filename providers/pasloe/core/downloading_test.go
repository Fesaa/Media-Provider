package core

import (
	"io"
	"reflect"
	"testing"
)

func TestCore_SetUserFiltered(t *testing.T) {
	chapter1 := SimpleChapter("1", "1")
	chapter2 := SimpleChapter("2", "2")
	chapter3 := SimpleChapter("3", "3")

	c := testBase(t, req(), io.Discard, ProviderMock{})
	c.ToDownload = []ChapterMock{chapter1, chapter2, chapter3}
	c.ToDownloadUserSelected = []string{"2"}
	c.ToRemoveContent = []string{
		c.ContentPath(chapter1) + ".cbz",
		c.ContentPath(chapter2) + ".cbz",
		c.ContentPath(chapter3) + ".cbz",
	}

	c.filterContentByUserSelection()

	if len(c.ToDownload) != 1 || c.ToDownload[0].Id != "2" {
		t.Errorf("Expected only '2' to remain in ToDownload")
	}

	expectedRemovals := []string{c.ContentPath(chapter2) + ".cbz"}
	if !reflect.DeepEqual(c.ToRemoveContent, expectedRemovals) {
		t.Errorf("Expected ToRemoveContent %v,\n got %v", expectedRemovals, c.ToRemoveContent)
	}
}
