package limetorrents

import (
	"net/url"
	"testing"
)

func TestReq(t *testing.T) {
	opt := SearchOptions{
		Category: ALL,
		Query:    url.QueryEscape("Spice and Wolf S01"),
		Page:     0,
	}
	search, err := Search(opt)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, torrent := range search {
		if torrent.Name == "Spice and Wolf S01 1080p BluRay 10-Bit Dual-Audio FLAC5 1 x265-YURASUKA" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("Unable to find: Spice and Wolf S01 1080p BluRay 10-Bit Dual-Audio FLAC5 1 x265-YURASUKA")
	}

}
