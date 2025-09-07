package core

import (
	"slices"
	"strings"

	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
)

const (
	DownloadOneShotKey       string = "download_one_shot"
	IncludeNotMatchedTagsKey string = "include_not_matched_tags"
	IncludeCover             string = "include_cover"
	UpdateCover              string = "update_cover"
	TitleOverride            string = "title_override"
)

type Tag interface {
	Value() string
	Identifier() string
}

type ScopedTag interface {
	IsGenre() bool
}

func NewStringTag(value string) Tag {
	return NewStringTagWithId(value, "")
}

func NewStringTagWithId(value string, id string) Tag {
	return &stringTag{value, id}
}

func NewStringTagWithIdAndGenre(value, id string, genre bool) Tag {
	return &stringTagWithScope{
		stringTag: stringTag{
			tag: value,
			id:  id,
		},
		genre: genre,
	}
}

type stringTag struct {
	tag string
	id  string
}

type stringTagWithScope struct {
	stringTag
	genre bool
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

func (t *stringTagWithScope) IsGenre() bool {
	return t.genre
}

// MapTags transforms all tags as configured by the tag mappings. This method keeps scoped tags, scoped
func (c *Core[C, S]) MapTags(mappings models.TagMaps, tags []Tag) []Tag {
	return utils.Map(tags, func(t Tag) Tag {
		val := mappings.MapTag(t.Value())
		id := mappings.MapTag(t.Identifier())

		if scoped, ok := t.(ScopedTag); ok {
			return NewStringTagWithIdAndGenre(val, id, scoped.IsGenre())
		}

		return NewStringTagWithId(val, id)
	})
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
func (c *Core[C, S]) GetGenreAndTags(tags []Tag) (string, string) {
	var genres, blackList, whitelist models.Tags
	var tagMappings models.TagMaps
	preferencesLoaded := c.Preference != nil

	if preferencesLoaded {
		genres = c.Preference.DynastyGenreTags
		blackList = c.Preference.BlackListedTags
		whitelist = c.Preference.WhiteListedTags
		tagMappings = c.Preference.TagMappings
	} else {
		c.Log.Warn().Msg("No genres or tags will be set, blacklist couldn't be loaded")
		c.WarnPreferencesFailedToLoad()
		if config.SkipTagsOnFailure {
			return "", ""
		}
	}

	tags = c.MapTags(tagMappings, tags)

	tagContains := func(slice models.Tags, tag Tag) bool {
		return slice.Contains(tag.Value()) || slice.Contains(tag.Identifier())
	}

	forceGenre := func(tag Tag) bool {
		if scoped, ok := tag.(ScopedTag); ok {
			return scoped.IsGenre()
		}
		return false
	}

	// Not blacklisted, configured as genre or forced
	tagAllowedAsGenre := func(tag Tag) bool {
		return preferencesLoaded &&
			!tagContains(blackList, tag) &&
			(tagContains(genres, tag) || forceGenre(tag))
	}
	// not blacklisted, whitelisted or include all, not a genre
	tagAllowedAsTag := func(tag Tag) bool {
		return preferencesLoaded &&
			!tagContains(blackList, tag) &&
			(tagContains(whitelist, tag) || c.Req.GetBool(IncludeNotMatchedTagsKey, false)) &&
			!tagContains(genres, tag) &&
			!forceGenre(tag)
	}

	filterTags := func(tags []Tag, f func(Tag) bool) []string {
		return utils.MaybeMap(tags, func(tag Tag) (string, bool) {
			return tag.Value(), f(tag)
		})
	}

	filteredGenres := filterTags(tags, tagAllowedAsGenre)
	filteredTags := filterTags(tags, tagAllowedAsTag)

	return strings.Join(filteredGenres, ", "), strings.Join(filteredTags, ", ")
}

// GetAgeRating returns the highest comicinfo.AgeRating that is mapped under the models.AgeRatingMappings
// Returns false if no Tag was mapped
func (c *Core[C, S]) GetAgeRating(tags []Tag) (comicinfo.AgeRating, bool) {
	if c.Preference == nil {
		c.Log.Warn().Msg("Could not load age rate mapping, not setting age rating")
		return "", false
	}

	tags = c.MapTags(c.Preference.TagMappings, tags)

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
