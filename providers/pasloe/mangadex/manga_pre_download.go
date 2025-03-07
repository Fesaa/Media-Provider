package mangadex

import (
	"context"
	"errors"
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
	"github.com/Fesaa/Media-Provider/utils"
	"os"
	"path"
	"regexp"
	"slices"
	"strconv"
)

func (m *manga) LoadInfo(ctx context.Context) chan struct{} {
	out := make(chan struct{})
	go func() {
		mangaInfo, err := m.repository.GetManga(ctx, m.id)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				m.Log.Error().Err(err).Msg("error while loading manga info")
			}
			m.Cancel()
			close(out)
			return
		}
		m.info = &mangaInfo.Data

		chapters, err := m.repository.GetChapters(ctx, m.id)
		if err != nil || chapters == nil {
			if !errors.Is(err, context.Canceled) {
				m.Log.Error().Err(err).Msg("error while loading chapter info")
			}
			m.Cancel()
			close(out)
			return
		}

		m.chapters = m.FilterChapters(chapters)
		m.SetSeriesStatus()

		if m.Req.GetBool(IncludeCover, true) {
			covers, err := m.repository.GetCoverImages(ctx, m.id)
			if err != nil || covers == nil {
				m.Log.Warn().Err(err).Msg("error while loading manga coverFactory, ignoring")
				m.coverFactory = defaultCoverFactory
			} else {
				m.coverFactory = m.getCoverFactoryLang(covers)
			}
		}

		close(out)
	}()
	return out
}

func (m *manga) SetSeriesStatus() {
	var maxVolume int64 = -1
	var maxChapter int64 = -1

	// If there is a last chapter present, but no last volume is given. We assume that the series does not use volumes
	m.foundLastVolume = m.info.Attributes.LastVolume == "" && m.info.Attributes.LastChapter != ""
	for _, ch := range m.chapters.Data {
		if ch.Attributes.Volume == m.info.Attributes.LastVolume && m.info.Attributes.LastVolume != "" {
			m.foundLastVolume = true
		}
		if ch.Attributes.Chapter == m.info.Attributes.LastChapter && m.info.Attributes.LastChapter != "" {
			m.foundLastChapter = true
		}

		if val, err := strconv.ParseInt(ch.Attributes.Volume, 10, 64); err == nil {
			maxVolume = max(maxVolume, val)
		} else {
			m.Log.Trace().Str("volume", ch.Attributes.Volume).Str("chapter", ch.Attributes.Chapter).
				Msg("not adding chapter, as Volume string isn't an int")
		}

		if val, err := strconv.ParseInt(ch.Attributes.Chapter, 10, 64); err == nil {
			maxChapter = max(maxChapter, val)
		} else {
			m.Log.Trace().Str("volume", ch.Attributes.Volume).Str("chapter", ch.Attributes.Chapter).
				Msg("not adding chapter, as Chapter string isn't an int")
		}
	}
	// We can set these safely as they're only written when found
	m.lastFoundVolume = int(maxVolume)
	m.lastFoundChapter = int(maxChapter)
}

func (m *manga) FilterChapters(c *ChapterSearchResponse) ChapterSearchResponse {
	scanlation := func() string {
		if scanlationGroup, ok := m.Req.GetString(ScanlationGroupKey); ok {
			m.Log.Debug().Str("scanlationGroup", scanlationGroup).
				Msg("loading manga info, prioritizing chapters from a specific Scanlation group or user")
			return scanlationGroup
		}

		return ""
	}()
	chaptersMap := utils.GroupBy(c.Data, func(v ChapterSearchData) string {
		return v.Attributes.Chapter
	})

	newData := make([]ChapterSearchData, 0)
	for _, chapters := range chaptersMap {
		chapter := utils.Find(chapters, m.chapterSearchFunc(scanlation, true))

		// Retry by skipping scanlation check
		if chapter == nil && scanlation != "" {
			chapter = utils.Find(chapters, m.chapterSearchFunc("", true))
		}

		if chapter != nil {
			newData = append(newData, *chapter)
		}
	}

	if m.Req.GetBool(DownloadOneShotKey) {
		// OneShots do not have a chapter, so will be mapped under the empty string
		if chapters, ok := chaptersMap[""]; ok {
			newData = append(newData, utils.Filter(chapters, m.chapterSearchFunc(scanlation, false))...)
		}
	}

	c.Data = newData
	return *c
}

func (m *manga) chapterSearchFunc(scanlation string, skipOneShot bool) func(ChapterSearchData) bool {
	return func(data ChapterSearchData) bool {
		if data.Attributes.TranslatedLanguage != m.language {
			return false
		}
		// Skip over official publisher chapters, we cannot download these from mangadex
		if data.Attributes.ExternalUrl != "" {
			return false
		}

		if data.Attributes.Chapter == "" && skipOneShot {
			return false
		}

		if scanlation == "" {
			return true
		}

		return slices.ContainsFunc(data.Relationships, func(relationship Relationship) bool {
			if relationship.Type != "scanlation_group" && relationship.Type != "user" {
				return false
			}

			return relationship.Id == scanlation
		})
	}
}

var (
	contentRegex = regexp.MustCompile(".* (?:Ch|Vol)\\. ([\\d|\\.]+).cbz")
	oneShotRegex = regexp.MustCompile(".+ OneShot .+\\.cbz")
)

func (m *manga) IsContent(name string) bool {
	if contentRegex.MatchString(name) {
		return true
	}

	if oneShotRegex.MatchString(name) {
		return true
	}

	return false
}

func (m *manga) ShouldDownload(chapter ChapterSearchData) bool {
	// Backwards compatibility check if volume has been downloaded
	if _, ok := m.GetContentByName(m.volumeDir(chapter.Attributes.Volume) + ".cbz"); ok {
		return false
	}

	content, ok := m.GetContentByName(m.ContentDir(chapter) + ".cbz")
	if !ok {
		return true
	}

	if chapter.Attributes.Volume == "" {
		return m.hasNewCover(chapter, content)
	}

	return m.replaceAndShouldDownload(chapter, content)
}

func (m *manga) replaceAndShouldDownload(chapter ChapterSearchData, content api.Content) bool {
	l := m.ContentLogger(chapter)
	fullPath := path.Join(m.Client.GetBaseDir(), content.Path)

	ci, err := m.metadataService.GetComicInfo(fullPath)
	if err != nil {
		l.Warn().Err(err).Str("path", fullPath).Msg("unable to read comic info in zip")
		return false
	}

	if strconv.Itoa(ci.Volume) == chapter.Attributes.Volume {
		l.Trace().Str("path", fullPath).Msg("Volume on disk matches, checking cover")
		return m.hasNewCover(chapter, content)
	}

	l.Debug().Int("onDiskVolume", ci.Volume).Str("path", fullPath).
		Msg("Loose chapter has been assigned to a volume, replacing")

	// Opted to remove, and redownload the entire chapter if the volume marker changes
	// One could argue that only the comicinfo.xml should be replaced if this happens.
	// Making the assumption that new content may be added in a chapter once it's added to a volume.
	// Especially the first, and last chapter of the volume.
	if err = os.Remove(fullPath); err != nil {
		l.Error().Err(err).Str("path", fullPath).Msg("unable to remove old chapter, not downloading new")
		return false
	}

	return true
}

func (m *manga) hasNewCover(chapter ChapterSearchData, content api.Content) bool {
	l := m.ContentLogger(chapter)
	fullPath := path.Join(m.Client.GetBaseDir(), content.Path)
	if !m.Req.GetBool(IncludeCover, true) {
		return false
	}

	l.Trace().Msg("Checking if a new cover is available")
	cover, ok := m.coverFactory(chapter.Attributes.Volume)
	if !ok {
		l.Trace().Str("volume", chapter.Attributes.Volume).Str("path", fullPath).Msg("no cover found")
		return false
	}

	existingCover, ok := m.metadataService.GetCover(fullPath)
	if !ok {
		l.Trace().Str("path", fullPath).Msg("no cover found in archive")
		return false
	}

	betterCover, _, err := m.getBetterChapterCover(chapter, cover)
	if err != nil {
		l.Error().Err(err).Msg("failed to get better cover")
		return false
	}

	if m.imageService.Equal(betterCover, existingCover) {
		l.Trace().Str("path", fullPath).Msg("cover is up to date")
		return false
	}

	l.Debug().Str("path", fullPath).Msg("a new cover has been found, re-downloading chapter")

	// Opted to remove, and redownload the entire chapter if the volume marker changes
	// One could argue that only the cover should be replaced if this happens.
	// Making the assumption that new content may be added in a chapter once it's added to a volume.
	// Especially the first, and last chapter of the volume.
	if err = os.Remove(fullPath); err != nil {
		l.Error().Err(err).Str("path", fullPath).Msg("unable to remove old chapter, not downloading new")
		return false
	}

	return true
}
