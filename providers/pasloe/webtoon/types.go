package webtoon

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/comicinfo"
	"net/url"
	"strings"
)

type SearchOptions struct {
	Query string
}

type Response struct {
	Result  SearchResult `json:"result"`
	Success bool         `json:"success"`
}

type SearchResult struct {
	Query        string       `json:"query"`
	Start        int          `json:"start"`
	Display      int          `json:"display"`
	Total        int          `json:"total"`
	SearchedList []SearchData `json:"searchedList"`
}

type SearchData struct {
	Id              int      `json:"titleNo"`
	Name            string   `json:"title"`
	ReadCount       int      `json:"readCount"`
	ThumbnailMobile string   `json:"thumbnailMobile"`
	AuthorNameList  []string `json:"authorNameList"`
	Genre           string   `json:"representGenre"`
	Rating          bool     `json:"titleUnsuitableForChildren"`
}

func (s *SearchData) Url() string {
	return fmt.Sprintf(BaseUrl+"%s/%s/list?title_no=%d", s.Genre, url.PathEscape(s.Name), s.Id)
}

func (s *SearchData) ProxiedImage() string {
	parts := strings.Split(strings.TrimPrefix(s.ThumbnailMobile, ImagePrefix), "/")
	if len(parts) != 3 {
		return ""
	}
	date := parts[0]
	id := parts[1]
	fileName := strings.TrimSuffix(parts[2], "?type=q90")
	return fmt.Sprintf("proxy/webtoon/covers/%s/%s/%s", date, id, fileName)
}

func (s *SearchData) ComicInfoRating() comicinfo.AgeRating {
	if s.Rating {
		return comicinfo.AgeRatingMaturePlus17
	}
	return comicinfo.AgeRatingEveryone
}

type Series struct {
	Id          string
	Name        string
	Authors     []string
	Description string
	Genre       string
	Completed   bool
	Chapters    []Chapter
}

type Chapter struct {
	Url      string
	ImageUrl string
	Title    string
	Number   string
	Date     string
}

func (c Chapter) ID() string {
	return c.Number
}
