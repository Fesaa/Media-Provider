package services

import (
	"bytes"
	"testing"

	"github.com/Fesaa/Media-Provider/internal/comicinfo"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
)

func TestArchiveService_GetComicInfo(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	ds := DirectoryServiceProvider(zerolog.Nop(), fs)
	as := ArchiveServiceProvider(zerolog.Nop(), fs)

	if err := fs.Mkdir("/testFile", 0755); err != nil {
		t.Fatal(err)
	}

	ci := comicinfo.NewComicInfo()
	ci.Series = "Spice and Wolf"
	if err := comicinfo.Save(fs, ci, "/testFile/comicinfo.xml"); err != nil {
		t.Fatal(err)
	}

	if err := ds.ZipToCbz("/testFile"); err != nil {
		t.Fatal(err)
	}

	ci, err := as.GetComicInfo("/testFile.cbz")
	if err != nil {
		t.Fatal(err)
	}

	if ci.Series != "Spice and Wolf" {
		t.Fatalf("Got %s; expected %s", ci.Series, "Spice and Wolf")
	}

}

func TestArchiveService_GetCover(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	ds := DirectoryServiceProvider(zerolog.Nop(), fs)
	as := ArchiveServiceProvider(zerolog.Nop(), fs)

	if err := fs.Mkdir("/testFile", 0755); err != nil {
		t.Fatal(err)
	}

	cover := []byte{1, 2, 3}
	if err := fs.WriteFile("/testFile/!0000 cover.jpg", cover, 0644); err != nil {
		t.Fatal(err)
	}

	if err := ds.ZipToCbz("/testFile"); err != nil {
		t.Fatal(err)
	}

	foundCover, err := as.GetCover("/testFile.cbz")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(foundCover, cover) {
		t.Fatalf("Got %s; expected %s", foundCover, cover)
	}
}
