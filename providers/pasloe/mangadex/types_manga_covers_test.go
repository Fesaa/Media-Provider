package mangadex

import (
	"context"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"io"
	"testing"
)

var coverReq = payload.DownloadRequest{
	Id:        "d0f9e331-e022-4b49-8399-e14091d8b703",
	TempTitle: "Listening to the Stars",
	BaseDir:   "Manga",
	Provider:  models.MANGADEX,
}

func TestManga_CoverSkipWrongFormatAndFirstAsDefault(t *testing.T) {
	m := tempManga(t, coverReq, io.Discard)

	covers, err := m.repository.GetCoverImages(context.Background(), m.id)
	if err != nil {
		t.Fatal(err)
	}

	factory := m.getCoverFactoryLang(covers)
	if factory == nil {
		t.Fatal("Cover factory not found")
	}

	cover, ok := factory("NotAValidChapter")
	if !ok {
		t.Fatal("there should have been a default cover")
	}

	got := cover.Data.Id
	want := "b77cbccb-82b3-43e9-a92e-dd39a4fcb6fc"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

}
