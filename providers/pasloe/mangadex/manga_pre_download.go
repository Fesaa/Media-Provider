package mangadex

import (
	"context"
	"errors"
	"github.com/Fesaa/Media-Provider/utils"
	"slices"
	"strconv"
)

func (m *manga) LoadInfo(ctx context.Context) chan struct{} {
	out := make(chan struct{})
	go func() {
		defer close(out)
		mangaInfo, err := m.repository.GetManga(ctx, m.id)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				m.Log.Error().Err(err).Msg("error while loading manga info")
			}
			m.Cancel()
			return
		}
		m.SeriesInfo = &mangaInfo.Data

		chapters, err := m.repository.GetChapters(ctx, m.id)
		if err != nil || chapters == nil {
			if !errors.Is(err, context.Canceled) {
				m.Log.Error().Err(err).Msg("error while loading chapter info")
			}
			m.Cancel()
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
	}()
	return out
}

func (m *manga) SetSeriesStatus() {
	var maxVolume int64 = -1
	var maxChapter int64 = -1

	// If there is a last chapter present, but no last volume is given. We assume that the series does not use volumes
	m.foundLastVolume = m.SeriesInfo.Attributes.LastVolume == "" && m.SeriesInfo.Attributes.LastChapter != ""
	for _, ch := range m.chapters.Data {
		if ch.Attributes.Volume == m.SeriesInfo.Attributes.LastVolume && m.SeriesInfo.Attributes.LastVolume != "" {
			m.foundLastVolume = true
		}
		if ch.Attributes.Chapter == m.SeriesInfo.Attributes.LastChapter && m.SeriesInfo.Attributes.LastChapter != "" {
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
