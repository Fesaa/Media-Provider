package core

type SeriesMock struct {
	id       string
	title    string
	chapters []ChapterMock
}

func (s SeriesMock) GetId() string {
	return s.id
}

func (s SeriesMock) GetTitle() string {
	return s.title
}

func (s SeriesMock) AllChapters() []ChapterMock {
	return s.chapters
}
