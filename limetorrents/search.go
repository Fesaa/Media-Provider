package limetorrents

import (
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

var base_url string = "https://www.limetorrents.lol/search/%s/%s/%d/"

func Search(searchOptions SearchOptions) ([]SearchResult, error) {
	url := formatUrl(searchOptions)

	doc, err := getSearch(url)
	if err != nil {
		return nil, err
	}

	torrents := doc.Find(".table2 tbody tr")
	return parseResults(torrents), nil
}

func parseResults(torrents *goquery.Selection) []SearchResult {
	results := make([]SearchResult, 0, torrents.Length()-1)

	torrents.Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return
		}

		name := s.Find("td:nth-child(1) a").Text()
		url, _ := s.Find("td:nth-child(1) a").Attr("href")
		added := s.Find("td:nth-child(2)").Text()
		size := s.Find("td:nth-child(3)").Text()
		seed := s.Find("td:nth-child(4)").Text()
		leach := s.Find("td:nth-child(5)").Text()

		results = append(results, SearchResult{
			Name:  name,
			Url:   url,
			Size:  size,
			Seed:  seed,
			Leach: leach,
			Added: added,
		})
	})

	return results
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
