package api

import (
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"slices"
	"strings"
)

const (
	DownloadOneShotKey       string = "download_one_shot"
	IncludeNotMatchedTagsKey string = "include_not_matched_tags"
	IncludeCover             string = "include_cover"
)

type Tag interface {
	Value() string
	Identifier() string
}

type ScopedTag interface {
	IsGenre() bool
}

func NewStringTag(value string) Tag {
	return &stringTag{value}
}

type stringTag struct {
	tag string
}

func (t *stringTag) Value() string {
	return t.tag
}

func (t *stringTag) Identifier() string {
	return t.tag
}

// GetGenreAndTags returns two comma-separated strings: one for genres and one for tags.
//
// A Tag is considered a genre if:
//   - It is not in the blacklist.
//   - It is mapped as a genre.
//
// A Tag is considered a tag if:
//   - It is not in the blocklist.
//   - It is not mapped as a genre.
//   - It is either in the whitelist or the request has IncludeNotMatchedTagsKey set to true.
func (d *DownloadBase[T]) GetGenreAndTags(tags []Tag) (string, string) {
	var genres, blackList, whitelist models.Tags
	p, err := d.preferences.GetComplete()
	if err != nil {
		d.Log.Error().Err(err).Msg("failed to get mapped genre tags, not setting any genres")
		if !d.hasWarnedTags {
			d.hasWarnedTags = true
			d.Notifier.NotifyContentQ(
				d.TransLoco.GetTranslation("blacklist-failed-to-load-title", d.infoProvider.Title()),
				d.TransLoco.GetTranslation("blacklist-failed-to-load-summary"),
				models.Orange)
		}
	} else {
		genres = p.DynastyGenreTags
		blackList = p.BlackListedTags
		whitelist = p.WhiteListedTags
	}

	tagContains := func(slice models.Tags, tag Tag) bool {
		return slice.Contains(tag.Value()) || slice.Contains(tag.Identifier())
	}

	forceGenre := func(tag Tag) bool {
		if scoped, ok := tag.(ScopedTag); ok {
			return scoped.IsGenre()
		}
		return false
	}

	tagAllowedAsGenre := func(tag Tag) bool {
		return err == nil &&
			!tagContains(blackList, tag) &&
			(tagContains(genres, tag) || forceGenre(tag))
	}
	tagAllowedAsTag := func(tag Tag) bool {
		return err == nil &&
			!tagContains(blackList, tag) &&
			tagContains(whitelist, tag) &&
			!tagContains(genres, tag) &&
			!forceGenre(tag)
	}

	filteredGenres := utils.MaybeMap(tags, func(t Tag) (string, bool) {
		if tagAllowedAsGenre(t) {
			return t.Value(), true
		}
		return "", false
	})

	filteredTags := utils.MaybeMap(tags, func(t Tag) (string, bool) {
		if tagAllowedAsTag(t) {
			return t.Value(), true
		}

		if d.Req.GetBool(IncludeNotMatchedTagsKey, false) &&
			!tagContains(genres, t) &&
			!tagContains(blackList, t) {
			return t.Value(), true
		}

		return "", false
	})

	return strings.Join(filteredGenres, ", "), strings.Join(filteredTags, ", ")
}

// GetAgeRating returns the highest comicinfo.AgeRating that is mapped under the models.AgeRatingMappings
// Returns false if no Tag was mapped
func (d *DownloadBase[T]) GetAgeRating(tags []Tag) (comicinfo.AgeRating, bool) {
	if d.Preference == nil {
		d.Log.Warn().Msg("Could not load age rate mapping, not setting age rating")
		return "", false
	}

	var mappings models.AgeRatingMappings = d.Preference.AgeRatingMappings
	weights := utils.MaybeMap(tags, func(t Tag) (int, bool) {
		ar, ok := mappings.GetAgeRating(t.Value())
		if !ok {
			return 0, false
		}

		return comicinfo.AgeRatingIndex[ar], true
	})

	if len(weights) == 0 {
		return "", false
	}

	return comicinfo.IndexToAgeRating[slices.Max(weights)], true
}
