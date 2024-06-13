package subsplease

import (
	"net/url"
	"testing"
)

func TestReq(t *testing.T) {
	opt := SearchOptions{
		Query: url.QueryEscape("Spice and Wolf"),
	}
	search, err := Search(opt)
	if err != nil {
		t.Fatal(err)
	}
	if search == nil || len(search) == 0 {
		t.Fatal("No torrents found")
	}
	found := false
	for _, data := range search {
		if data.Show == "Spice and Wolf (2024)" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Unable to find show: Spice and Wolf (2024)")
	}
}
