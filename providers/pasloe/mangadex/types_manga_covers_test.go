package mangadex

import (
	"testing"
)

func TestMangaCoverResponse_GetCoverFactory(t *testing.T) {

	m := &MangaCoverResponse{
		Data: []MangaCoverData{
			{
				Attributes: MangaCoverAttributes{
					Volume:   "1",
					FileName: "UseAsDefault",
				},
			},
			{
				Attributes: MangaCoverAttributes{
					Volume:   "2",
					FileName: "SecondVolumeCover",
				},
			},
		},
	}

	factory := m.GetCoverFactoryLang("en", "myId")

	got, ok := factory("2")
	if !ok {
		t.Error("expected cover to exist")
	}
	want := "https://uploads.mangadex.org/covers/myId/SecondVolumeCover.512.jpg"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	got, ok = factory("3")
	if !ok {
		t.Error("expected cover not to exist")
	}

	want = "https://uploads.mangadex.org/covers/myId/UseAsDefault.512.jpg"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

}

func TestMangaCoverResponse_GetCoverFactoryNoDefault(t *testing.T) {
	m := &MangaCoverResponse{Data: []MangaCoverData{}}

	factory := m.GetCoverFactoryLang("en", "myId")

	got, ok := factory("1")
	if ok {
		t.Error("expected cover to not exist")
	}

	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}
