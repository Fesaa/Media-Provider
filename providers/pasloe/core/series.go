package core

type Series[C Chapter] interface {
	GetId() string
	GetTitle() string
	AllChapters() []C
}

func (c *Core[C, S]) GetAllLoadedChapters() []C {
	if chapterCustomizer, ok := c.impl.(ChapterCustomizer[C]); ok {
		return chapterCustomizer.CustomizeAllChapters()
	}

	return c.SeriesInfo.AllChapters()
}
