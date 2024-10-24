package models

var DefaultPages = []*Page{
	{
		Title:     "Anime",
		SortValue: 1,
		Providers: []Provider{NYAA, SUBSPLEASE},
		Modifiers: []Modifier{
			{
				Title: "Category",
				Key:   "categories",
				Type:  DROPDOWN,
				Values: map[string]string{
					"anime":         "Anime",
					"anime-amv":     "Music Video",
					"anime-eng":     "English Translated",
					"anime-non-eng": "Non-English Translated",
				},
			},
			{
				Title: "Sort",
				Key:   "sortBys",
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
		Title:     "Mangadex",
		SortValue: 2,
		Providers: []Provider{MANGADEX},
		Modifiers: []Modifier{
			{
				Title: "Include Tags",
				Key:   "includeTags",
				Type:  MULTI,
				Values: map[string]string{
					"Romance":   "Romance",
					"Ninja":     "Ninja",
					"Comedy":    "Comedy",
					"Mecha":     "Mecha",
					"Anthology": "Anthology",
				},
			},
			{
				Title: "Exclude Tags",
				Key:   "excludeTags",
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
			{
				Title: "Status",
				Key:   "status",
				Type:  MULTI,
				Values: map[string]string{
					"ongoing":   "Ongoing",
					"completed": "Completed",
					"hiatus":    "Hiatus",
					"cancelled": "Cancelled",
				},
			},
			{
				Title: "Content Rating",
				Key:   "contentRating",
				Type:  MULTI,
				Values: map[string]string{
					"safe":       "Safe",
					"suggestive": "Suggestive",
				},
			},
			{
				Title: "Demographic",
				Key:   "publicationDemographic",
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
		Title:     "Manga & Light Novels",
		SortValue: 3,
		Providers: []Provider{NYAA},
		Modifiers: []Modifier{
			{
				Title: "Category",
				Key:   "categories",
				Type:  DROPDOWN,
				Values: map[string]string{
					"literature-eng":     "English Literature",
					"literature":         "Literature",
					"literature-non-eng": "Non English Literature",
					"literature-raw":     "Raw Literature",
				},
			},
			{
				Title: "Sort by",
				Key:   "sortBys",
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
		Title:     "Movies",
		SortValue: 4,
		Providers: []Provider{YTS},
		Modifiers: []Modifier{
			{
				Title: "Sort By",
				Key:   "sortBys",
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
		Title:     "Lime",
		SortValue: 5,
		Providers: []Provider{LIME},
		Modifiers: []Modifier{
			{
				Title: "Category",
				Key:   "categories",
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
	{
		Title:         "WebToon",
		SortValue:     6,
		Providers:     []Provider{WEBTOON},
		Modifiers:     []Modifier{},
		Dirs:          []string{"Manga"},
		CustomRootDir: "Manga",
	},
}
