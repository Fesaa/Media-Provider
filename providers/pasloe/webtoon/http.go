package webtoon

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
)

func wrapInDoc(url string, httpClient *http.Client) (*goquery.Document, error) {
	res, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}
