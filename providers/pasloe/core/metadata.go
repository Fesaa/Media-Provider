package core

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
	return &stringTag{value, ""}
}

func NewStringTagWithId(value string, id string) Tag {
	return &stringTag{value, id}
}

type stringTag struct {
	tag string
	id  string
}

func (t *stringTag) Value() string {
	return t.tag
}

func (t *stringTag) Identifier() string {
	if t.id != "" {
		return t.id
	}
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
func (c *Core[T]) GetGenreAndTags(tags []Tag) (string, string) {
	var genres, blackList, whitelist models.Tags
	var tagMappings models.TagMaps

	p, err := c.preferences.GetComplete()
	if err != nil {
		c.Log.Error().Err(err).Msg("failed to get mapped genre tags, not setting any genres")
		if !c.hasWarnedTags {
			c.hasWarnedTags = true
			c.Notifier.NotifyContentQ(
				c.TransLoco.GetTranslation("blacklist-failed-to-load-title", c.infoProvider.Title()),
				c.TransLoco.GetTranslation("blacklist-failed-to-load-summary"),
				models.Orange)
		}
	} else {
		genres = p.DynastyGenreTags
		blackList = p.BlackListedTags
		whitelist = p.WhiteListedTags
		tagMappings = p.TagMappings
	}

	tags = utils.Map(tags, func(t Tag) Tag {
		val := tagMappings.MapTag(t.Value())
		id := tagMappings.MapTag(t.Identifier())
		return NewStringTagWithId(val, id)
	})

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

		if c.Req.GetBool(IncludeNotMatchedTagsKey, false) &&
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
func (c *Core[T]) GetAgeRating(tags []Tag) (comicinfo.AgeRating, bool) {
	if c.Preference == nil {
		c.Log.Warn().Msg("Could not load age rate mapping, not setting age rating")
		return "", false
	}

	tags = utils.Map(tags, func(t Tag) Tag {
		val := models.TagMaps(c.Preference.TagMappings).MapTag(t.Value())
		id := models.TagMaps(c.Preference.TagMappings).MapTag(t.Identifier())
		return NewStringTagWithId(val, id)
	})

	var mappings models.AgeRatingMappings = c.Preference.AgeRatingMappings
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
