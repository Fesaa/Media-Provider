package dynasty

import (
	"strconv"
	"strings"
	"time"
)

type SearchOptions struct {
	Query string
}

type SearchData struct {
	Id      string
	Title   string
	Authors []Author
	Tags    []Tag
}

type Author = Identifiable

type Tag = Identifiable

func (i Identifiable) Value() string {
	return i.DisplayName
}

func (i Identifiable) Identifier() string {
	return i.Id
}

type Identifiable struct {
	DisplayName string
	Id          string
}

func (s *SearchData) RefUrl() string {
	if strings.HasPrefix(s.Id, "/series/") {
		return DOMAIN + s.Id
	}
	return DOMAIN + "/series/" + s.Id
}

type Series struct {
	Id          string
	Title       string
	AltTitle    string
	Description string
	Status      SeriesStatus
	CoverUrl    string

	Authors []Author
	Tags    []Tag

	Chapters []Chapter
}

func (s *Series) GetId() string {
	return s.Id
}

func (s *Series) GetTitle() string {
	return s.Title
}

func (s *Series) AllChapters() []Chapter {
	return s.Chapters
}

func (s *Series) RefUrl() string {
	return seriesURL(s.Id)
}

type Chapter struct {
	Id          string
	Title       string
	Volume      string
	Chapter     string
	ReleaseDate *time.Time
	Tags        []Tag
	Authors     []Author
}

func (c Chapter) GetChapter() string {
	return c.Chapter
}

func (c Chapter) GetVolume() string {
	return c.Volume
}

func (c Chapter) GetTitle() string {
	return c.Title
}

func (c Chapter) GetId() string {
	return c.Id
}

func (c Chapter) VolumeFloat() float64 {
	if c.Volume == "" {
		return -1
	}
	if vol, err := strconv.ParseFloat(c.Volume, 64); err == nil {
		return vol
	}
	return -1
}

func (c Chapter) ChapterFloat() float64 {
	if c.Chapter == "" {
		return -1
	}
	if v, err := strconv.ParseFloat(c.Chapter, 64); err == nil {
		return v
	}
	return -1
}

type SeriesStatus string

const (
	Completed SeriesStatus = "Completed"
	Ongoing   SeriesStatus = "Ongoing"
)
