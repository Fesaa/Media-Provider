package config

import (
	"log/slog"
	"os"
	"path"
)

func defaultConfig() *Config {
	secret, err := GenerateSecret(64)
	if err != nil {
		panic(err)
	}

	return &Config{
		SyncId:  0,
		RootDir: path.Join(OrDefault(os.Getenv("CONFIG_DIR"), "."), "temp"),
		BaseUrl: "",
		Secret:  secret,
		Cache: CacheConfig{
			Type: MEMORY,
		},
		Logging: Logging{
			Level:   slog.LevelInfo,
			Source:  true,
			Handler: LogHandlerText,
		},
		Downloader: Downloader{
			MaxConcurrentTorrents:       5,
			MaxConcurrentMangadexImages: 4,
		},
		Pages: []Page{
			{
				Title:    "Anime",
				Provider: []Provider{NYAA, SUBSPLEASE},
				Modifiers: map[string]Modifier{
					"categories": {
						Title: "Category",
						Type:  DROPDOWN,
						Values: map[string]string{
							"anime":         "Anime",
							"anime-amv":     "Music Video",
							"anime-eng":     "English Translated",
							"anime-non-eng": "Non-English Translated",
						},
					},
					"sortBys": {
						Title: "Sort",
						Type:  DROPDOWN,
						Values: map[string]string{
							"downloads": "Downloads",
							"date":      "Date",
							"size":      "Size",
							"seeders":   "Seeders",
							"leechers":  "Leechers",
							"comments":  "Comments",
						},
					},
				},
				Dirs:          []string{"Anime"},
				CustomRootDir: "Anime",
			},
			{
				Title:    "Mangadex",
				Provider: []Provider{MANGADEX},
				Modifiers: map[string]Modifier{
					"includeTags": {
						Title: "Include Tags",
						Type:  MULTI,
						Values: map[string]string{
							"Romance":   "Romance",
							"Ninja":     "Ninja",
							"Comedy":    "Comedy",
							"Mecha":     "Mecha",
							"Anthology": "Anthology",
						},
					},
					"excludeTags": {
						Title: "Exclude Tags",
						Type:  MULTI,
						Values: map[string]string{
							"Cooking":      "Cooking",
							"Supernatural": "Supernatural",
							"Mystery":      "Mystery",
							"Adaptation":   "Adaptation",
							"Music":        "Music",
							"Full Color":   "Full Color",
						},
					},
					"status": {
						Title: "Status",
						Type:  MULTI,
						Values: map[string]string{
							"ongoing":   "Ongoing",
							"completed": "Completed",
							"hiatus":    "Hiatus",
							"cancelled": "Cancelled",
						},
					},
					"contentRating": {
						Title: "Content Rating",
						Type:  MULTI,
						Values: map[string]string{
							"safe":       "Safe",
							"suggestive": "Suggestive",
						},
					},
					"publicationDemographic": {
						Title: "Demographic",
						Type:  MULTI,
						Values: map[string]string{
							"shounen": "Shounen",
							"shoujo":  "Shoujo",
							"josei":   "Josei",
							"seinen":  "Seinen",
						},
					},
				},
				Dirs:          []string{"Manga"},
				CustomRootDir: "Manga",
			},
			{
				Title:    "Manga & Light Novels",
				Provider: []Provider{NYAA},
				Modifiers: map[string]Modifier{
					"categories": {
						Title: "Category",
						Type:  DROPDOWN,
						Values: map[string]string{
							"literature-eng":     "English Literature",
							"literature":         "Literature",
							"literature-non-eng": "Non English Literature",
							"literature-raw":     "Raw Literature",
						},
					},
					"sortBys": {
						Title: "Sort by",
						Type:  DROPDOWN,
						Values: map[string]string{
							"downloads": "Downloads",
							"date":      "Date",
							"size":      "Size",
							"seeders":   "Seeders",
							"leechers":  "Leechers",
							"comments":  "Comments",
						},
					},
				},
				Dirs:          []string{"Manga", "LightNovels"},
				CustomRootDir: "",
			},
			{
				Title:    "Movies",
				Provider: []Provider{YTS},
				Modifiers: map[string]Modifier{
					"sortBys": {
						Title: "Sort By",
						Type:  DROPDOWN,
						Values: map[string]string{
							"title":          "Title",
							"year":           "Year",
							"rating":         "Rating",
							"peers":          "Peers",
							"seeds":          "Seeders",
							"download_count": "Downloads",
							"like_count":     "Likes",
							"date_added":     "Date Added",
						},
					},
				},
				Dirs:          []string{"Movies"},
				CustomRootDir: "Movies",
			},
			{
				Title:    "Lime",
				Provider: []Provider{LIME},
				Modifiers: map[string]Modifier{
					"categories": {
						Title: "Category",
						Type:  DROPDOWN,
						Values: map[string]string{
							"ALL":    "All categories",
							"MOVIES": "Movies",
							"TV":     "TV",
							"ANIME":  "Anime",
							"OTHER":  "Other",
						},
					},
				},
				Dirs:          []string{"Anime", "Movies", "Manga", "Series", "LightNovels"},
				CustomRootDir: "",
			},
		},
	}
}
