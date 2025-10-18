package bato

import "github.com/Fesaa/Media-Provider/providers/pasloe/publication"

type SearchOptions struct {
	Query              string
	Genres             []string
	IgnoredGenres      []string
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

func toPublicationStatus(s string) publication.Status {
	switch s {
	case "pending":
		return publication.StatusOngoing
	case "ongoing":
		return publication.StatusOngoing
	case "completed":
		return publication.StatusCompleted
	case "hiatus":
		return publication.StatusPaused
	case "cancelled":
		return publication.StatusCancelled
	}
	return ""
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
