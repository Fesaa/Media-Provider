package core

import (
	"path"
	"strconv"
)

func (c *Core[C, S]) ShouldDownload(chapter C) bool {
	// Backwards compatibility check if volume has been downloaded
	if _, ok := c.GetContentByName(c.VolumeDir(chapter) + ".cbz"); ok {
		return false
	}

	content, ok := c.GetContentByName(c.ContentFileName(chapter) + ".cbz")
	if !ok {
		return true
	}

	// Redownload
	if chapter.GetVolume() != "" && c.hasBeenAssignedVolume(chapter, content) {
		fullPath := path.Join(c.Client.GetBaseDir(), content.Path)
		c.ToRemoveContent = append(c.ToRemoveContent, fullPath)
		return true
	}

	return false
}

func (c *Core[C, S]) hasBeenAssignedVolume(chapter C, content Content) bool {
	l := c.ContentLogger(chapter)
	fullPath := path.Join(c.Client.GetBaseDir(), content.Path)

	ci, err := c.archiveService.GetComicInfo(fullPath)
	if err != nil {
		l.Warn().Err(err).Str("path", fullPath).Msg("unable to read comic info in zip")
		return false
	}

	if strconv.Itoa(ci.Volume) == chapter.GetVolume() {
		l.Trace().Str("path", fullPath).Msg("Volume on disk matches, not replacing")
		return false
	}

	l.Debug().Int("onDiskVolume", ci.Volume).Str("path", fullPath).
		Msg("Loose chapter has been assigned to a volume, replacing")
	return true
}

/**
func (m *manga) hasOutdatedCover(chapter ChapterSearchData, content core.Content) bool {
	if !m.Req.GetBool(UpdateCover, false) || !m.Req.GetBool(IncludeCover, true) {
		return false
	}

	l := m.ContentLogger(chapter)
	fullPath := path.Join(m.Client.GetBaseDir(), content.Path)

	wantedCover, firstPage := m.getChapterCover(chapter)
	if wantedCover == nil {
		l.Debug().Str("path", fullPath).Msg("no cover found")
		return false
	}

	coverOnDisk, err := m.archiveService.GetCover(fullPath)
	if err != nil {
		l.Debug().Err(err).Str("path", fullPath).Bool("firstPage", firstPage).
			Msg("unable to read cover, may be first page")
		// If no cover was found in the archive, and there is a wanted cover. Lets re-download
		// If the cover is the first page, ArchiveService.GetCover will return ErrNoMatch.
		return errors.Is(err, services.ErrNoMatch) && !firstPage
	}

	return m.coverShouldBeReplaced(chapter, wantedCover, coverOnDisk)
}

func (m *manga) coverShouldBeReplaced(chapter ChapterSearchData, wantedCover, coverOnDisk []byte) bool {
	l := m.ContentLogger(chapter)
	wantedImg, err := m.imageService.ToImage(wantedCover)
	if err != nil {
		l.Warn().Err(err).Msg("unable to convert wanted cover to image")
		return false
	}

	onDiskImg, err := m.imageService.ToImage(coverOnDisk)
	if err != nil {
		l.Warn().Err(err).Msg("unable to convert on disk cover to image")
		return false
	}

	similar := m.imageService.Similar(onDiskImg, wantedImg)
	if similar > 0.85 {
		l.Trace().Float64("similar", similar).Msg("on disk image is similar to wanted image, not re-downloading")
		return false
	}

	l.Debug().Msg("on disk image is different from wanted image, re-downloading")
	return true
}

*/
