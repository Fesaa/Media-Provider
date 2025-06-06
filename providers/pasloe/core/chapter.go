package core

// Chapter represents any downloadable group of images
//
// A chapter is considered standalone/OneShot if GetVolume and GetChapter both return an empty string
type Chapter interface {
	GetId() string
	Label() string

	GetChapter() string
	GetVolume() string
	GetTitle() string
}

func IsOneShot(chapter Chapter) bool {
	return chapter.GetChapter() == "" && chapter.GetVolume() == ""
}
