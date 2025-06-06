/*
Package comicinfo
MIT License

# Copyright (c) 2023 Felipe Martin

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

https://github.com/fmartingr/go-comicinfo/blob/latest/schema.go, but with Tags support, and custom Kavitatags
*/
package comicinfo

import "slices"

var xmlHeader = []byte(`<?xml version="1.0" encoding="UTF-8"?>`)

// ComicInfo ...
type ComicInfo struct {
	Title               string    `xml:"Title,omitempty"`
	Series              string    `xml:"Series,omitempty"`
	LocalizedSeries     string    `xml:"LocalizedSeries,omitempty"` // Kavita only
	Number              string    `xml:"Number,omitempty"`
	Count               int       `xml:"Count,omitempty"`
	Volume              int       `xml:"Volume,omitempty"`
	AlternateSeries     string    `xml:"AlternateSeries,omitempty"`
	AlternateNumber     string    `xml:"AlternateNumber,omitempty"`
	AlternateCount      int       `xml:"AlternateCount,omitempty"`
	Summary             string    `xml:"Summary,omitempty"`
	Notes               string    `xml:"Notes,omitempty"`
	Year                int       `xml:"Year,omitempty"`
	Month               int       `xml:"Month,omitempty"`
	Day                 int       `xml:"Day,omitempty"`
	Writer              string    `xml:"Writer,omitempty"`
	Penciller           string    `xml:"Penciller,omitempty"`
	Inker               string    `xml:"Inker,omitempty"`
	Colorist            string    `xml:"Colorist,omitempty"`
	Letterer            string    `xml:"Letterer,omitempty"`
	CoverArtist         string    `xml:"CoverArtist,omitempty"`
	Editor              string    `xml:"Editor,omitempty"`
	Publisher           string    `xml:"Publisher,omitempty"`
	Imprint             string    `xml:"Imprint,omitempty"`
	Genre               string    `xml:"Genre,omitempty"`
	Tags                string    `xml:"Tags,omitempty"`
	Web                 string    `xml:"Web,omitempty"`
	PageCount           int       `xml:"PageCount,omitempty"`
	LanguageISO         string    `xml:"LanguageISO,omitempty"`
	Format              string    `xml:"Format,omitempty"`
	BlackAndWhite       YesNo     `xml:"BlackAndWhite,omitempty"`
	Manga               Manga     `xml:"Manga,omitempty"`
	Characters          string    `xml:"Characters,omitempty"`
	Teams               string    `xml:"Teams,omitempty"`
	Locations           string    `xml:"Locations,omitempty"`
	ScanInformation     string    `xml:"ScanInformation,omitempty"`
	StoryArc            string    `xml:"StoryArc,omitempty"`
	SeriesGroup         string    `xml:"SeriesGroup,omitempty"`
	AgeRating           AgeRating `xml:"AgeRating,omitempty"`
	Pages               Pages     `xml:"Pages,omitempty"`
	CommunityRating     Rating    `xml:"CommunityRating,omitempty"`
	MainCharacterOrTeam string    `xml:"MainCharacterOrTeam,omitempty"`
	Review              string    `xml:"Review,omitempty"`

	// Internal
	XmlnsXsd string `xml:"xmlns:xsd,attr"`
	XmlNsXsi string `xml:"xmlns:xsi,attr"`
	// XsiSchemaLocation string `xml:"xsi:schemaLocation,attr,omitempty"`
}

type Roles []Role

func (r Roles) HasRole(role Role) bool {
	return slices.Contains(r, role)
}

type Role string

const (
	Writer      Role = "writer"
	Penciller   Role = "penciler"
	Inker       Role = "inker"
	Colorist    Role = "colorist"
	Letterer    Role = "letterer"
	CoverArtist Role = "cover_artist"
	Editor      Role = "editor"
)

func (ci *ComicInfo) SetXMLAttributes() {
	ci.XmlnsXsd = "http://www.w3.org/2001/XMLSchema"
	ci.XmlNsXsi = "http://www.w3.org/2001/XMLSchema-instance"
}

// NewComicInfo provides a new ComicInfo struct with the XML attributes set
func NewComicInfo() *ComicInfo {
	ci := ComicInfo{}
	ci.SetXMLAttributes()
	return &ci
}

// YesNo defines the YesNo type
type YesNo string

var (
	YesNoUnknown YesNo = "Unknown"
	Yes          YesNo = "Yes"
	No           YesNo = "No"
)

// Manga defines the Manga type
type Manga string

var (
	MangaUnknown           Manga = "Unknown"
	MangaYes               Manga = "Yes"
	MangaNo                Manga = "No"
	MangeYesAndRightToLeft Manga = "YesAndRightToLeft"
)

// Rating defines the Rating type
type Rating float64

// AgeRating defines the AgeRating type
type AgeRating string

var (
	AgeRatingUnknown          AgeRating = "Unknown"
	AgeRatingPending          AgeRating = "Rating Pending"
	AgeRatingEarlyChildhood   AgeRating = "Early Childhood"
	AgeRatingEveryone         AgeRating = "Everyone"
	AgeRatingG                AgeRating = "G"
	AgeRatingEveryone10Plus   AgeRating = "Everyone 10+"
	AgeRatingPG               AgeRating = "PG"
	AgeRatingKidsToAdults     AgeRating = "Kids to Adults"
	AgeRatingTeen             AgeRating = "Teen"
	AgeRatingMAPlus15         AgeRating = "MA15+"
	AgeRatingMaturePlus17     AgeRating = "Mature 17+"
	AgeRatingM                AgeRating = "M"
	AgeRatingRPlus18          AgeRating = "R18+"
	AgeRatingAdultsOnlyPlus18 AgeRating = "Adults Only 18+"
	AgeRatingXPlus18          AgeRating = "X18+"
)

var AgeRatingIndex = map[AgeRating]int{
	AgeRatingUnknown:          0,
	AgeRatingPending:          1,
	AgeRatingEarlyChildhood:   2,
	AgeRatingEveryone:         3,
	AgeRatingG:                4,
	AgeRatingEveryone10Plus:   5,
	AgeRatingPG:               6,
	AgeRatingKidsToAdults:     7,
	AgeRatingTeen:             8,
	AgeRatingMAPlus15:         9,
	AgeRatingMaturePlus17:     10,
	AgeRatingM:                11,
	AgeRatingRPlus18:          12,
	AgeRatingAdultsOnlyPlus18: 13,
	AgeRatingXPlus18:          14,
}

var IndexToAgeRating = []AgeRating{
	AgeRatingUnknown,
	AgeRatingPending,
	AgeRatingEarlyChildhood,
	AgeRatingEveryone,
	AgeRatingG,
	AgeRatingEveryone10Plus,
	AgeRatingPG,
	AgeRatingKidsToAdults,
	AgeRatingTeen,
	AgeRatingMAPlus15,
	AgeRatingMaturePlus17,
	AgeRatingM,
	AgeRatingRPlus18,
	AgeRatingAdultsOnlyPlus18,
	AgeRatingXPlus18,
}

// Pages defines the Pages type (slice of ComicPageInfo for proper XML marshalling)
type Pages struct {
	Pages []ComicPageInfo `xml:"Page,omitempty"`
}

func (p *Pages) Append(page ComicPageInfo) {
	p.Pages = append(p.Pages, page)
}

func (p *Pages) Len() int {
	return len(p.Pages)
}

// ComicPageInfo defines the ComicPageInfo type
type ComicPageInfo struct {
	Image       int           `xml:"Image,attr"`
	Type        ComicPageType `xml:"Type,omitempty,attr"`
	DoublePage  bool          `xml:"DoublePage,omitempty,attr"`
	ImageSize   int64         `xml:"ImageSize,omitempty,attr"`
	Key         string        `xml:"Key,omitempty,attr"`
	Bookmark    string        `xml:"Bookmark,omitempty,attr"`
	ImageWidth  int           `xml:"ImageWidth,omitempty,attr"`
	ImageHeight int           `xml:"ImageHeight,omitempty,attr"`
}

// NewComicPageInfo provides a new ComicPageInfo struct with the XML attributes set
func NewComicPageInfo() ComicPageInfo {
	return ComicPageInfo{
		Type: ComicPageTypeStory,
	}
}

// ComicPageType defines the ComicPageType type
type ComicPageType string

var (
	ComicPageTypeFrontCover    ComicPageType = "FrontCover"
	ComicPageTypeInnerCover    ComicPageType = "InnerCover"
	ComicPageTypeRoundup       ComicPageType = "Roundup"
	ComicPageTypeStory         ComicPageType = "Story"
	ComicPageTypeAdvertisement ComicPageType = "Advertisement"
	ComicPageTypeEditorial     ComicPageType = "Editorial"
	ComicPageTypeLetters       ComicPageType = "Letters"
	ComicPageTypePreview       ComicPageType = "Preview"
	ComicPageTypeBackCover     ComicPageType = "BackCover"
	ComicPageTypeOther         ComicPageType = "Other"
	ComicPageTypeDeleted       ComicPageType = "Deleted"
)
