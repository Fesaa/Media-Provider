package services

import (
	"testing"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/devfeel/mapper"
	"github.com/rs/zerolog"
)

func tempPageService(t *testing.T) PageService {
	t.Helper()
	log := zerolog.Nop()

	tempDir := t.TempDir()
	config.Dir = tempDir

	database, err := db.DatabaseProvider(log)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		d, err := database.DB()
		if err != nil {
			t.Fatal(err)
		}
		d.Close()
	})
	return PageServiceProvider(db.NewUnitOfWork(database, mapper.NewMapper()), log)
}

func TestPageService_LoadDefaultPages(t *testing.T) {
	ps := tempPageService(t)

	if err := ps.LoadDefaultPages(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func TestPageService_UpdateOrCreate(t *testing.T) {
	ps := tempPageService(t)

	page := models.Page{
		Title:         "MyTitle",
		Icon:          "pi-heart",
		SortValue:     2,
		Providers:     []int64{2},
		Modifiers:     nil,
		Dirs:          []string{"dir1", "dir2"},
		CustomRootDir: "",
	}

	if err := ps.UpdateOrCreate(t.Context(), &page); err != nil {
		t.Fatal(err)
	}
}

func TestPageService_UpdateOrCreateDefaultSortValue(t *testing.T) {
	ps := tempPageService(t)

	page := models.Page{
		Title:         "MyTitle",
		Icon:          "pi-heart",
		SortValue:     DefaultPageSort,
		Providers:     []int64{2},
		Modifiers:     nil,
		Dirs:          []string{"dir1", "dir2"},
		CustomRootDir: "",
	}

	if err := ps.UpdateOrCreate(t.Context(), &page); err != nil {
		t.Fatal(err)
	}

	if page.SortValue != 0 {
		t.Fatalf("expected page.SortValue to be 0, as this was the first page, got %d", page.SortValue)
	}

	secondPage := models.Page{
		Title:         "MyTitle2",
		Icon:          "pi-heart",
		SortValue:     DefaultPageSort,
		Providers:     []int64{2},
		Modifiers:     nil,
		Dirs:          []string{"dir1", "dir2"},
		CustomRootDir: "",
	}

	if err := ps.UpdateOrCreate(t.Context(), &secondPage); err != nil {
		t.Fatal(err)
	}

	if secondPage.SortValue != 1 {
		t.Fatalf("expected page.SortValue to be 1, as this was the second page, got %d", secondPage.SortValue)
	}

}
