package mangadex

import "fmt"

type ChapterImageSearchResponse struct {
	Result  string      `json:"result"`
	BaseUrl string      `json:"baseUrl"`
	Chapter ChapterInfo `json:"chapter"`
}

// FullImageUrls returns the urls for full quality
func (s *ChapterImageSearchResponse) FullImageUrls() []string {
	urls := make([]string, len(s.Chapter.Data))
	for i, image := range s.Chapter.Data {
		urls[i] = fmt.Sprintf("%s/data/%s/%s", s.BaseUrl, s.Chapter.Hash, image)
	}
	return urls
}

type ChapterInfo struct {
	Hash      string   `json:"hash"`
	Data      []string `json:"data"`
	DataSaver []string `json:"dataSaver"`
}
