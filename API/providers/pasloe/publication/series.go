package publication

import (
	"fmt"
	"path"
	"slices"
	"strconv"
	"time"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/internal/comicinfo"
	"github.com/Fesaa/Media-Provider/utils"
)

// Series represents any series that may represent a publication
type Series struct {
	Id          string
	Title       string
	AltTitle    string // Alternative/Original title
	Description string
	CoverUrl    string
	RefUrl      string

	Status            Status
	TranslationStatus utils.Settable[Status]
	Year              int
	OriginalLanguage  string

	HighestVolume  utils.Settable[float64]
	HighestChapter utils.Settable[float64]

	ContentRating comicinfo.AgeRating
	Tags          []Tag
	People        []Person
	Links         []string

	Chapters []Chapter
}

// A Chapter is the smallest bit of info that can be downloaded.
type Chapter struct {
	Id       string
	Title    string
	Volume   string
	Chapter  string
	CoverUrl string
	Url      string

	Summary     string
	ReleaseDate *time.Time
	Translator  []string

	Tags   []Tag
	People []Person
}

// VolumeFloat returns the volume as a float64, or -1 if empty/invalid
func (c Chapter) VolumeFloat() float64 {
	if c.Volume == "" {
		return -1
	}
	if vol, err := strconv.ParseFloat(c.Volume, 64); err == nil {
		return vol
	}
	return -1
}

// ChapterFloat returns the chapter as a float64, or -1 if empty/invalid
func (c Chapter) ChapterFloat() float64 {
	if c.Chapter == "" {
		return -1
	}
	if ch, err := strconv.ParseFloat(c.Chapter, 64); err == nil {
		return ch
	}
	return -1
}

func (c Chapter) Label() string {
	if c.Chapter != "" && c.Volume != "" {
		return fmt.Sprintf("Volume %s Chapter %s: %s", c.Volume, c.Chapter, c.Title)
	}

	if c.Chapter != "" {
		return fmt.Sprintf("Chapter %s: %s", c.Chapter, c.Title)
	}

	return fmt.Sprintf("OneShot: %s", c.Title)
}

func (p *publication) VolumeDir(chapter Chapter) string {
	return fmt.Sprintf("%s Vol. %s", p.Title(), chapter.Volume)
}

// ContentPath returns the full path to the directory where images, and metadata for a chapter
// should be downloaded to
func (p *publication) ContentPath(chapter Chapter) string {
	base := path.Join(p.client.GetBaseDir(), p.req.BaseDir, p.Title())

	if chapter.Volume != "" && !config.DisableVolumeDirs {
		base = path.Join(base, p.VolumeDir(chapter))
	}

	return path.Join(base, p.ContentFileName(chapter))
}

// ContentFileName return the filename for a chapter
func (p *publication) ContentFileName(chapter Chapter) string {
	if chapter.Chapter == "" {
		return p.OneShotFileName(chapter)
	}

	return p.DefaultFileName(chapter)
}

// DefaultFileName return the filename for a chapter with a non-empty chapter marker
func (p *publication) DefaultFileName(chapter Chapter) string {
	fileName := p.Title()

	if chapter.Volume != "" && p.ShouldIncludeVolume() {
		fileName += fmt.Sprintf(" Vol. %s", chapter.Volume)
	}

	if _, err := strconv.ParseFloat(chapter.Chapter, 32); err != nil {
		p.log.Warn().Err(err).Str("chapter", chapter.Chapter).Msg("Failed to parse chapter, not padding")
		return fmt.Sprintf("%s Ch. %s", fileName, chapter.Chapter)
	}

	padded := utils.PadFloatFromString(chapter.Chapter, 4)
	return fmt.Sprintf("%s Ch. %s", fileName, padded)
}

// ShouldIncludeVolume returns true if chapter filenames should include a volume marker
func (p *publication) ShouldIncludeVolume() bool {
	if config.DisableVolumeDirs {
		return true
	}

	if b, ok := p.hasDuplicatedChapters.Get(); ok {
		return b
	}

	groupedByChapter := utils.GroupBy(p.series.Chapters, func(c Chapter) string {
		return c.Chapter
	})

	for _, chapterGroup := range groupedByChapter {
		if len(chapterGroup) > 1 {
			p.hasDuplicatedChapters.Set(true)
			return true
		}
	}

	p.hasDuplicatedChapters.Set(false)
	return false
}

// OneShotFileName returns the filename for a chapter with an empty chapter marker
func (p *publication) OneShotFileName(chapter Chapter) string {
	oneShotPath := fmt.Sprintf("%s %s", p.Title(), chapter.Title)
	if !config.DisableOneShotInFileName {
		oneShotPath += " (One Shot)"
	}

	finalOneShotPath := oneShotPath
	for i := 0; slices.Contains(p.hasDownloaded, finalOneShotPath); i++ {
		finalOneShotPath = fmt.Sprintf("%s (%d)", oneShotPath, i)
		if i >= 25 {
			p.log.Warn().
				Str("chapter", chapter.Title).
				Int("tries", i).
				Msg("Amount of unnamed, or same named OneShots has exceeded 25. Falling back to random generated string")
			finalOneShotPath = fmt.Sprintf("%s (%s)", oneShotPath, utils.MustReturn(utils.GenerateSecret(8)))
		}
	}

	return finalOneShotPath
}

// Person represents a creator (author, artist, translator, etc)
type Person struct {
	Name  string
	Url   string
	Roles comicinfo.Roles
}

// Tag represents a genre or tag
type Tag struct {
	Value      string
	Identifier string
	IsGenre    bool // True if the source marked this as a genre vs a tag
}

func NonEmptyTag(t Tag) bool {
	return t.Value != "" || t.Identifier != ""
}

func NonEmptyPerson(p Person) bool {
	return p.Name != ""
}

type Status string

const (
	StatusOngoing   Status = "ongoing"
	StatusCompleted Status = "completed"
	StatusPaused    Status = "paused" // Hiatus/Pending
	StatusCancelled Status = "cancelled"
)
