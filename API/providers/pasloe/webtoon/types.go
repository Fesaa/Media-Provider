package webtoon

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Fesaa/Media-Provider/internal/comicinfo"
)

type SearchOptions struct {
	Query string
}

type SearchData struct {
	Id              string   `json:"titleNo"`
	Name            string   `json:"title"`
	ReadCount       string   `json:"readCount"`
	ThumbnailMobile string   `json:"thumbnailMobile"`
	AuthorNameList  []string `json:"authorNameList"`
	Genre           string   `json:"representGenre"`
	Rating          bool     `json:"titleUnsuitableForChildren"`
}

func (s *SearchData) Url() string {
	return fmt.Sprintf(BaseUrl+"%s/%s/list?title_no=%s", s.Genre, url.PathEscape(s.Name), s.Id)
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

func (s *Series) RefUrl() string {
	return fmt.Sprintf(BaseUrl+"%s/%s/list?title_no=%s", s.Genre, url.PathEscape(s.Name), s.Id)
}

func (s *Series) GetId() string {
	return s.Id
}

func (s *Series) GetTitle() string {
	return s.Name
}

func (s *Series) AllChapters() []Chapter {
	return s.Chapters
}

type Chapter struct {
	Url      string
	ImageUrl string
	Title    string
	Number   string
	Date     string
}

func (c Chapter) GetChapter() string {
	return c.Number
}

func (c Chapter) GetVolume() string {
	return ""
}

func (c Chapter) GetTitle() string {
	return c.Title
}

func (c Chapter) GetId() string {
	return c.Number
}
