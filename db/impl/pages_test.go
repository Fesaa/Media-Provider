package impl

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"testing"
)

func seedPages(t *testing.T, p models.Pages) {
	t.Helper()

	for _, page := range models.DefaultPages {
		if err := p.New(&page); err != nil {
			t.Fatal(err)
		}
	}
}

func TestPages_All(t *testing.T) {
	p := Pages(databaseHelper(t))
	seedPages(t, p)

	want := 6
	got, err := p.All()
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != want {
		t.Errorf("got %d pages, want %d", len(got), want)
	}

	mangadex := utils.Find(got, func(page models.Page) bool {
		return page.Title == "Mangadex"
	})

	if mangadex == nil {
		t.Fatal("mangadex not found")
	}
}

func TestPages_Delete(t *testing.T) {
	p := Pages(databaseHelper(t))
	seedPages(t, p)

	if err := p.Delete(1); err != nil {
		t.Fatal(err)
	}

	want := 5
	got, err := p.All()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != want {
		t.Errorf("got %d pages, want %d", len(got), want)
	}

}

func TestPages_Get(t *testing.T) {
	p := Pages(databaseHelper(t))
	seedPages(t, p)

	want := "Mangadex"
	got, err := p.Get(2)
	if err != nil {
		t.Fatal(err)
	}

	if got.Title != want {
		t.Errorf("got %s, want %s", got.Title, want)
	}
}

func TestPages_GetNoResults(t *testing.T) {
	p := Pages(databaseHelper(t))

	got, err := p.Get(1)
	if err != nil {
		t.Fatal(err)
	}

	if got != nil {
		t.Errorf("got %v, want nil", got)
	}

}

func TestPages_New(t *testing.T) {
	p := Pages(databaseHelper(t))

	if err := p.New(&models.DefaultPages[0]); err != nil {
		t.Fatal(err)
	}

	got, err := p.Get(1)
	if err != nil {
		t.Fatal(err)
	}

	if got.Title != "Anime" {
		t.Errorf("got %s, want %s", got.Title, "Anime")
	}

	if got.Providers[0] != int64(models.NYAA) {
		t.Errorf("got %d, want %d", got.Providers[1], int64(models.NYAA))
	}

	if len(got.Modifiers) != 2 {
		t.Errorf("got %d, want %d", len(got.Modifiers), 3)
	}

	if got.Modifiers[0].Title != "Category" {
		t.Errorf("got %s, want %s", got.Modifiers[0].Title, "Category")
	}

	if len(got.Modifiers[0].Values) != 4 {
		t.Errorf("got %d, want %d", len(got.Modifiers[0].Values), 4)
	}

	if got.Modifiers[0].Values[3].Key != "anime-non-eng" {
		t.Errorf("got %s, want %s", got.Modifiers[0].Key, "anime-non-eng")
	}
}

func TestPages_Update(t *testing.T) {
	p := Pages(databaseHelper(t))
	seedPages(t, p)

	first, err := p.Get(1)
	if err != nil {
		t.Fatal(err)
	}

	// Update depth 0
	first.Icon = "MyIcon"
	if err = p.Update(first); err != nil {
		t.Fatal(err)
	}

	first, err = p.Get(1)
	if err != nil {
		t.Fatal(err)
	}

	if first.Icon != "MyIcon" {
		t.Errorf("got %s, want %s", first.Icon, "MyIcon")
	}

	// Update depth 1
	first.Modifiers[0].Title = "MyUpdatedModifier"
	if err = p.Update(first); err != nil {
		t.Fatal(err)
	}

	first, err = p.Get(1)
	if err != nil {
		t.Fatal(err)
	}

	if first.Modifiers[0].Title != "MyUpdatedModifier" {
		t.Errorf("got %s, want %s", first.Modifiers[0].Title, "MyUpdatedModifier")
	}

	// Update depth 2
	first.Modifiers[1].Values[2].Value = "MyUpdatedModifierValue"
	if err = p.Update(first); err != nil {
		t.Fatal(err)
	}

	first, err = p.Get(1)
	if err != nil {
		t.Fatal(err)
	}

	if first.Modifiers[1].Values[2].Value != "MyUpdatedModifierValue" {
		t.Errorf("got %s, want %s", first.Modifiers[1].Values[2].Value, "MyUpdatedModifierValue")
	}

}
