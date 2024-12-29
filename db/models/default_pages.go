package models

import (
	"github.com/lib/pq"
)

var DefaultPages = []Page{
	{
		Title:     "Anime",
		SortValue: 1,
		Providers: pq.Int64Array{int64(NYAA), int64(SUBSPLEASE)},
		Modifiers: []Modifier{
			{
				Title: "Category",
				Key:   "categories",
				Type:  DROPDOWN,
				Values: []ModifierValue{
					{
						Key:   "anime",
						Value: "Anime",
					},
					{
						Key:   "anime-amv",
						Value: "Music Video",
					},
					{
						Key:   "anime-eng",
						Value: "English Translated",
					},
					{
						Key:   "anime-non-eng",
						Value: "Non-English Translated",
					},
				},
			},
			{
				Title: "Sort",
				Key:   "sortBys",
				Type:  DROPDOWN,
				Values: []ModifierValue{
					{
						Key:   "downloads",
						Value: "Downloads",
					},
					{
						Key:   "date",
						Value: "Date",
					},
					{
						Key:   "size",
						Value: "Size",
					},
					{
						Key:   "seeders",
						Value: "Seeders",
					},
					{
						Key:   "leechers",
						Value: "Leechers",
					},
					{
						Key:   "comments",
						Value: "Comments",
					},
				},
			},
		},
		Dirs:          []string{"Anime"},
		CustomRootDir: "Anime",
	},
	{
		Title:     "Mangadex",
		SortValue: 2,
		Providers: pq.Int64Array{int64(MANGADEX)},
		Modifiers: []Modifier{
			{
				Title: "Include Tags",
				Key:   "includeTags",
				Type:  MULTI,
				Values: []ModifierValue{
					{
						Key:   "Romance",
						Value: "Romance",
					},
					{
						Key:   "Ninja",
						Value: "Ninja",
					},
					{
						Key:   "Comedy",
						Value: "Comedy",
					},
					{
						Key:   "Mecha",
						Value: "Mecha",
					},
					{
						Key:   "Anthology",
						Value: "Anthology",
					},
				},
			},
			{
				Title: "Exclude Tags",
				Key:   "excludeTags",
				Type:  MULTI,
				Values: []ModifierValue{
					{
						Key:   "Cooking",
						Value: "Cooking",
					},
					{
						Key:   "Supernatural",
						Value: "Supernatural",
					},
					{
						Key:   "Mystery",
						Value: "Mystery",
					},
					{
						Key:   "Adaptation",
						Value: "Adaptation",
					},
					{
						Key:   "Music",
						Value: "Music",
					},
					{
						Key:   "Full Color",
						Value: "Full Color",
					},
				},
			},
			{
				Title: "Status",
				Key:   "status",
				Type:  MULTI,
				Values: []ModifierValue{
					{
						Key:   "ongoing",
						Value: "Ongoing",
					},
					{
						Key:   "completed",
						Value: "Completed",
					},
					{
						Key:   "hiatus",
						Value: "Hiatus",
					},
					{
						Key:   "cancelled",
						Value: "Cancelled",
					},
				},
			},
			{
				Title: "Content Rating",
				Key:   "contentRating",
				Type:  MULTI,
				Values: []ModifierValue{
					{
						Key:   "safe",
						Value: "Safe",
					},
					{
						Key:   "suggestive",
						Value: "Suggestive",
					},
				},
			},
			{
				Title: "Demographic",
				Key:   "publicationDemographic",
				Type:  MULTI,
				Values: []ModifierValue{
					{
						Key:   "shounen",
						Value: "Shounen",
					},
					{
						Key:   "shoujo",
						Value: "Shoujo",
					},
					{
						Key:   "josei",
						Value: "Josei",
					},
					{
						Key:   "seinen",
						Value: "Seinen",
					},
				},
			},
		},
		Dirs:          []string{"Manga"},
		CustomRootDir: "Manga",
	},
	{
		Title:     "Manga & Light Novels",
		SortValue: 3,
		Providers: pq.Int64Array{int64(NYAA)},
		Modifiers: []Modifier{
			{
				Title: "Category",
				Key:   "categories",
				Type:  DROPDOWN,
				Values: []ModifierValue{
					{
						Key:   "literature-eng",
						Value: "English Literature",
					},
					{
						Key:   "literature",
						Value: "Literature",
					},
					{
						Key:   "literature-non-eng",
						Value: "Non English Literature",
					},
					{
						Key:   "literature-raw",
						Value: "Raw Literature",
					},
				},
			},
			{
				Title: "Sort by",
				Key:   "sortBys",
				Type:  DROPDOWN,
				Values: []ModifierValue{
					{
						Key:   "downloads",
						Value: "Downloads",
					},
					{
						Key:   "date",
						Value: "Date",
					},
					{
						Key:   "size",
						Value: "Size",
					},
					{
						Key:   "seeders",
						Value: "Seeders",
					},
					{
						Key:   "leechers",
						Value: "Leechers",
					},
					{
						Key:   "comments",
						Value: "Comments",
					},
				},
			},
		},
		Dirs:          []string{"Manga", "LightNovels"},
		CustomRootDir: "",
	},
	{
		Title:     "Movies",
		SortValue: 4,
		Providers: pq.Int64Array{int64(NYAA)},
		Modifiers: []Modifier{
			{
				Title: "Sort By",
				Key:   "sortBys",
				Type:  DROPDOWN,
				Values: []ModifierValue{
					{
						Key:   "title",
						Value: "Title",
					},
					{
						Key:   "year",
						Value: "Year",
					},
					{
						Key:   "rating",
						Value: "Rating",
					},
					{
						Key:   "peers",
						Value: "Peers",
					},
					{
						Key:   "seeds",
						Value: "Seeders",
					},
					{
						Key:   "download_count",
						Value: "Downloads",
					},
					{
						Key:   "like_count",
						Value: "Likes",
					},
					{
						Key:   "date_added",
						Value: "Date Added",
					},
				},
			},
		},
		Dirs:          []string{"Movies"},
		CustomRootDir: "Movies",
	},
	{
		Title:     "Lime",
		SortValue: 5,
		Providers: pq.Int64Array{int64(LIME)},
		Modifiers: []Modifier{
			{
				Title: "Category",
				Key:   "categories",
				Type:  DROPDOWN,
				Values: []ModifierValue{
					{
						Key:   "ALL",
						Value: "All categories",
					},
					{
						Key:   "MOVIES",
						Value: "Movies",
					},
					{
						Key:   "TV",
						Value: "TV",
					},
					{
						Key:   "ANIME",
						Value: "Anime",
					},
					{
						Key:   "OTHER",
						Value: "Other",
					},
				},
			},
		},
		Dirs:          []string{"Anime", "Movies", "Manga", "Series", "LightNovels"},
		CustomRootDir: "",
	},
	{
		Title:         "WebToon",
		SortValue:     6,
		Providers:     pq.Int64Array{int64(WEBTOON)},
		Modifiers:     []Modifier{},
		Dirs:          []string{"Manga"},
		CustomRootDir: "Manga",
	},
}
