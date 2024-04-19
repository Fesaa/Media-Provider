package limetorrents

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Fesaa/Media-Provider/utils"
	"github.com/PuerkitoBio/goquery"
)

var base_url string = "https://www.limetorrents.lol/search/%s/%s/%d/"
var cache utils.Cache[[]SearchResult] = *utils.NewCache[[]SearchResult](5 * time.Minute)

func Search(searchOptions SearchOptions) ([]SearchResult, error) {
	url := formatUrl(searchOptions)
	if res := cache.Get(url); res != nil {
		return *res, nil
	}

	doc, err := getSearch(url)
	if err != nil {
		return nil, err
	}

	torrents := doc.Find(".table2 tbody tr")
	res := parseResults(torrents)
	cache.Set(url, res)
	return res, nil
}

func parseResults(torrents *goquery.Selection) []SearchResult {
	// Either no results or only the header
	if torrents.Length() < 2 {
		return []SearchResult{}
	}
	results := goquery.Map(torrents, searchFromNode)
	// Don't return the header
	return results[1:]
}

func searchFromNode(i int, s *goquery.Selection) SearchResult {
	name := s.Find("td:nth-child(1) a").Text()
	url, _ := s.Find("td:nth-child(1) a").Attr("href")
	added := s.Find("td:nth-child(2)").Text()
	size := s.Find("td:nth-child(3)").Text()
	seed := s.Find("td:nth-child(4)").Text()
	leach := s.Find("td:nth-child(5)").Text()

	return SearchResult{
		Name:  name,
		Url:   url,
		Hash:  hashFromUrl(url),
		Size:  size,
		Seed:  seed,
		Leach: leach,
		Added: added,
	}
}

func hashFromUrl(url string) string {
	if url == "" {
		return ""
	}
	s1 := strings.Split(url, "torrent/")
	s2 := strings.Split(s1[1], ".torrent")
	return s2[0]
}

func getSearch(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func formatUrl(s SearchOptions) string {
	return fmt.Sprintf(base_url, s.Category, s.Query, s.Page)
}
