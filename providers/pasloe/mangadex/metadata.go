package mangadex

import "github.com/Fesaa/Media-Provider/http/payload"

const (
	LanguageKey string = "tl-lang"
	// ScanlationGroupKey filter chapters on the group that translated it
	ScanlationGroupKey string = "scanlation_group"
	// AllowNonMatchingScanlationGroupKey if we should use chapters from groups not matching ScanlationGroupKey
	// Only takes effect if ScanlationGroupKey is set
	AllowNonMatchingScanlationGroupKey string = "allow_non_matching_scanlation_group"
)

var languages = []payload.MetadataOption{
	{
		Key:   "en",
		Value: "English",
	},
	{
		Key:   "zh",
		Value: "Simplified Chinese",
	},
	{
		Key:   "zh-hk",
		Value: "Traditional Chinese",
	},
	{
		Key:   "es",
		Value: "Castilian Spanish",
	},
	{
		Key:   "fr",
		Value: "French",
	},
	{
		Key:   "ja",
		Value: "Japanese",
	}}
