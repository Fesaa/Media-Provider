package mangadex

import "github.com/Fesaa/Media-Provider/http/payload"

const (
	LanguageKey        string = "tl-lang"
	ScanlationGroupKey string = "scanlation_group"
	DownloadOneShotKey string = "download_one_shot"
	IncludeCover       string = "include_cover"
	UpdateCover        string = "update_cover"
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
