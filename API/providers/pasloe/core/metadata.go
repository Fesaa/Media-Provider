package core

import (
	"context"
	"slices"
	"strings"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/internal/comicinfo"
	"github.com/Fesaa/Media-Provider/utils"
)

const (
	DownloadOneShotKey       string = "download_one_shot"
	IncludeNotMatchedTagsKey string = "include_not_matched_tags"
	IncludeCover             string = "include_cover"
	UpdateCover              string = "update_cover"
	TitleOverride            string = "title_override"
	AssignEmptyVolumes       string = "assign_empty_volumes"
	ScanlationGroupKey       string = "scanlation_group"
)

type Tag interface {
	Value() string
	Identifier() string
}

func NoneEmptyTag[T Tag](t T) bool {
	return len(t.Value()) > 0
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

func mapTag(mappings []models.TagMapping, tag string) string {
	tagN := utils.Normalize(tag)

	for _, m := range mappings {
		if m.OriginTag == tagN {
			return m.DestinationTag
		}
	}

	return tag
}

// MapTags transforms all tags as configured by the tag mappings. This method keeps scoped tags, scoped
func (c *Core[C, S]) MapTags(mappings []models.TagMapping, tags []Tag) []Tag {
	return utils.Map(tags, func(t Tag) Tag {
		val := mapTag(mappings, t.Value())
		id := mapTag(mappings, t.Identifier())

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
func (c *Core[C, S]) GetGenreAndTags(ctx context.Context, tags []Tag) (string, string) {
	var genres, blackList, whitelist []string
	var tagMappings []models.TagMapping
	preferencesLoaded := c.Preference != nil

	if preferencesLoaded {
		genres = utils.Map(c.Preference.GenreList, utils.Normalize)
		blackList = utils.Map(c.Preference.BlackList, utils.Normalize)
		whitelist = utils.Map(c.Preference.WhiteList, utils.Normalize)
		tagMappings = utils.Map(c.Preference.TagMappings, func(t models.TagMapping) models.TagMapping {
			return models.TagMapping{
				OriginTag:      utils.Normalize(t.OriginTag),
				DestinationTag: t.DestinationTag,
			}
		})
	} else {
		c.Log.Warn().Msg("No genres or tags will be set, blacklist couldn't be loaded")
		c.WarnPreferencesFailedToLoad(ctx)
		if config.SkipTagsOnFailure {
			return "", ""
		}
	}

	tags = c.MapTags(tagMappings, tags)

	tagContains := func(slice []string, tag Tag) bool {
		return slices.Contains(slice, utils.Normalize(tag.Value())) ||
			slices.Contains(slice, utils.Normalize(tag.Identifier()))
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

	tagMappings := utils.Map(c.Preference.TagMappings, func(t models.TagMapping) models.TagMapping {
		return models.TagMapping{
			OriginTag:      utils.Normalize(t.OriginTag),
			DestinationTag: t.DestinationTag,
		}
	})
	tags = c.MapTags(tagMappings, tags)

	mappings := c.Preference.AgeRatingMappings
	weights := utils.MaybeMap(tags, func(t Tag) (int, bool) {
		ar, ok := GetAgeRating(mappings, utils.Normalize(t.Value()))
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

func GetAgeRating(arm []models.AgeRatingMapping, tag string) (comicinfo.AgeRating, bool) {
	ageRating := -1
	for _, ageRatingMapping := range arm {
		if utils.Normalize(ageRatingMapping.Tag) != tag {
			continue
		}

		ageRating = max(ageRating, comicinfo.AgeRatingIndex[ageRatingMapping.ComicInfoAgeRating])
	}

	if ageRating > -1 {
		return comicinfo.IndexToAgeRating[ageRating], true
	}

	return "", false
}
