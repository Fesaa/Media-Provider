package core

import "github.com/Fesaa/Media-Provider/utils"

type Series[C Chapter] interface {
	GetId() string
	GetTitle() string
	AllChapters() []C
}

func (c *Core[C, S]) Title() string {
	//invalid operation: c.SeriesInfo != nil (mismatched types S and untyped nil)
	// ???????
	if utils.IsNil(c.SeriesInfo) {
		return utils.NonEmpty(c.Req.TempTitle, c.Req.Id)
	}

	return utils.NonEmpty(c.SeriesInfo.GetTitle(), c.Req.TempTitle, c.Req.Id)
}

func (c *Core[C, S]) GetAllLoadedChapters() []C {
	if chapterCustomizer, ok := c.impl.(ChapterCustomizer[C]); ok {
		return chapterCustomizer.CustomizeAllChapters()
	}

	return c.SeriesInfo.AllChapters()
}
