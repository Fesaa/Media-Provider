package limetorrents

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Fesaa/Media-Provider/utils"
	"github.com/PuerkitoBio/goquery"
)

const BASE_URl string = "https://www.limetorrents.lol"
const SEARCH_URL string = BASE_URl + "/search/%s/%s/%d/"

var cache utils.Cache[[]SearchResult] = *utils.NewCache[[]SearchResult](5 * time.Minute)

func Search(searchOptions SearchOptions) ([]SearchResult, error) {
	url := formatUrl(searchOptions)
	slog.Info("Searching lime for torrents", "url", url)
	if res := cache.Get(url); res != nil {
		slog.Debug("Cache hit", "url", url)
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
	urlSel := s.Find("td:nth-child(1) a")
	url, _ := urlSel.First().Attr("href")
	pageUrl, _ := urlSel.Last().Attr("href")
	added := s.Find("td:nth-child(2)").Text()
	size := s.Find("td:nth-child(3)").Text()
	seed := s.Find("td:nth-child(4)").Text()
	leach := s.Find("td:nth-child(5)").Text()

	return SearchResult{
		Name:    name,
		Url:     url,
		Hash:    hashFromUrl(url),
		Size:    size,
		Seed:    seed,
		Leach:   leach,
		Added:   added,
		PageUrl: BASE_URl + pageUrl,
	}
}

func hashFromUrl(url string) string {
	if url == "" {
		return ""
	}
	s1 := strings.Split(url, "torrent/")
	if len(s1) < 2 {
		return ""
	}
	s2 := strings.Split(s1[1], ".torrent")
	if len(s2) == 0 {
		return ""
	}
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
	return fmt.Sprintf(SEARCH_URL, s.Category, s.Query, s.Page)
}
