package providers

/*func mangadexNormalizer(mangas *mangadex.MangaSearchResponse) []payload.Info {
	if mangas == nil {
		return []payload.Info{}
	}

	info := make([]payload.Info, 0)
	for _, data := range mangas.Data {
		enTitle := data.Attributes.EnTitle()
		if enTitle == "" {
			continue
		}

		info = append(info, payload.Info{
			Name:        enTitle,
			Description: data.Attributes.EnDescription(),
			Size: func() string {
				volumes := config.OrDefault(data.Attributes.LastVolume, "unknown")
				chapters := config.OrDefault(data.Attributes.LastChapter, "unknown")
				return fmt.Sprintf("%s Vol. %s Ch.", volumes, chapters)
			}(),
			Tags: []InfoTag{
				of("Date", strconv.Itoa(data.Attributes.Year)),
			},
			InfoHash: data.Id,
			RefUrl:   data.RefURL(),
			Provider: models.MANGADEX,
			ImageUrl: data.CoverURL(),
		})
	}

	return info
}

func nyaaNormalizer(provider models.Provider) responseNormalizerFunc[[]types.Torrent] {
	return func(torrents []types.Torrent) []payload.Info {
		torrentsInfo := make([]payload.Info, len(torrents))
		for i, t := range torrents {
			torrentsInfo[i] = payload.Info{
				Name:        t.Name,
				Description: "", // The description passed here, is some raw html nonsense. Don't use it
				Size:        t.Size,
				Tags: []InfoTag{
					of("Date", t.Date),
					of("Seeders", t.Seeders),
					of("Leechers", t.Leechers),
					of("Downloads", t.Downloads),
				},
				Link:     t.Link,
				InfoHash: t.InfoHash,
				ImageUrl: "",
				RefUrl:   t.GUID,
				Provider: provider,
			}
		}
		return torrentsInfo
	}
}

func dynastyNormalizer(series []dynasty_scans.SearchData) []payload.Info {
	return utils.Map(series, func(t dynasty_scans.SearchData) payload.Info {
		return payload.Info{
			Name:     t.Title,
			InfoHash: t.Id,
			Tags: utils.Map(t.Tags, func(t string) InfoTag {
				return of("", t)
			}),
			RefUrl:   t.RefUrl(),
			Provider: models.DYNASTY,
		}
	})
}
*/
