package dynasty

import (
	"reflect"
	"testing"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
)

func TestBuilder_Transform(t *testing.T) {
	b := &Builder{}

	got := b.Transform(t.Context(), payload.SearchRequest{
		Query: "test",
	})

	want := SearchOptions{
		Query: "test",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %v, want: %v", got, want)
	}
}

func TestBuilder_NormalizeNil(t *testing.T) {
	b := &Builder{}

	got := b.Normalize(t.Context(), nil)
	if got == nil {
		t.Errorf("got: %v, want: %v", got, []payload.Info{})
	}
	if len(got) != 0 {
		t.Errorf("got: %v, want: %v", len(got), 0)
	}
}

func TestBuilder_Normalize(t *testing.T) {
	b := &Builder{}

	in := []SearchData{
		{
			Id:    "id1",
			Title: "Title1",
			Authors: []Author{{
				DisplayName: "Author1",
				Id:          "AuthorId1",
			}},
			Tags: []Tag{{
				DisplayName: "Tag1",
				Id:          "TagId1",
			}},
		},
		{
			Id:    "id2",
			Title: "Title2",
			Authors: []Author{{
				DisplayName: "Author2",
				Id:          "AuthorId2",
			}},
			Tags: []Tag{},
		},
	}

	got := b.Normalize(t.Context(), in)

	want := payload.Info{
		Name:     "Title1",
		Tags:     []payload.InfoTag{{Name: "Tag1", Value: ""}},
		InfoHash: "id1",
		RefUrl:   DOMAIN + "/series/id1",
		Provider: models.DYNASTY,
	}

	if len(got) != 2 {
		t.Errorf("got: %v, want: %v", len(got), 2)
	}

	if !reflect.DeepEqual(got[0], want) {
		t.Errorf("got: %v, want: %v", got[0], want)
	}
}
