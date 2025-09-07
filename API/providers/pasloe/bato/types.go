package bato

import (
	"fmt"
	"strconv"

	"github.com/Fesaa/Media-Provider/internal/comicinfo"
)

type SearchOptions struct {
	Query              string
	Genres             []string
	OriginalLang       []string
	TranslatedLang     []string
	OriginalWorkStatus []Publication
	BatoUploadStatus   []Publication
}

type Publication string

const (
	PublicationPending   Publication = "pending"
	PublicationOngoing   Publication = "ongoing"
	PublicationCompleted Publication = "completed"
	PublicationHiatus    Publication = "hiatus"
	PublicationCancelled Publication = "cancelled"
)

func toPublication(s string) (Publication, bool) {
	switch s {
	case "pending":
		return PublicationPending, true
	case "ongoing":
		return PublicationOngoing, true
	case "completed":
		return PublicationCompleted, true
	case "hiatus":
		return PublicationHiatus, true
	case "cancelled":
		return PublicationCancelled, true
	}
	return "", false
}

type SearchResult struct {
	Id            string
	ImageUrl      string
	Title         string
	Authors       []string
	Tags          []string
	LatestChapter string
	UploaderImg   string
	LastUploaded  string
}

type Series struct {
	Id                string
	Title             string
	CoverUrl          string
	OriginalTitle     string
	Authors           []Author
	Tags              []string
	PublicationStatus Publication
	BatoUploadStatus  Publication
	Summary           string
	WebLinks          []string
	Chapters          []Chapter
}

func (s *Series) AllChapters() []Chapter {
	return s.Chapters
}

func (s *Series) GetId() string {
	return s.Id
}

func (s *Series) GetTitle() string {
	return s.Title
}

func (s *Series) RefUrl() string {
	return fmt.Sprintf("%s/title/%s", Domain, s.Id)
}

type Author struct {
	Name  string
	Roles comicinfo.Roles
}

type Chapter struct {
	Id      string
	Title   string
	Volume  string
	Chapter string
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

func (c Chapter) GetId() string {
	return c.Id
}
