package providers

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/limetorrents"
	"github.com/Fesaa/Media-Provider/subsplease"
	"github.com/Fesaa/Media-Provider/yts"
	"github.com/irevenko/go-nyaa/nyaa"
)

var providers = map[config.Provider]provider{}

func init() {
	register(config.NYAA, nyaaTransformer, nyaa.Search, nyaaNormalizer, yoitsuDownloader)
	register(config.LIME, limeTransformer, limetorrents.Search, limeNormalizer, yoitsuDownloader)
	register(config.YTS, ytsTransformer, yts.Search, ytsNormalizer, yoitsuDownloader)
	register(config.SUBSPLEASE, subsPleaseTransformer, subsplease.Search, subsPleaseNormalizer, yoitsuDownloader)
}

func register[T, S any](name config.Provider, t requestTransformerFunc[S], s searchFunc[S, T], n responseNormalizerFunc[T], d downloadFunc) {
	providers[name] = &providerImpl[T, S]{
		transformer: t,
		normalizer:  n,
		searcher:    s,
		downloader:  d,
	}
}

type responseNormalizerFunc[T any] func(T) []TorrentInfo
type requestTransformerFunc[S any] func(SearchRequest) S
type searchFunc[S, T any] func(S) (T, error)
type downloadFunc func(req DownloadRequest) error

type providerImpl[T any, S any] struct {
	transformer requestTransformerFunc[S]
	normalizer  responseNormalizerFunc[T]
	searcher    searchFunc[S, T]
	downloader  downloadFunc
}

func (s *providerImpl[T, S]) Download(req DownloadRequest) error {
	return s.downloader(req)
}

func (s *providerImpl[T, S]) Search(req SearchRequest) ([]TorrentInfo, error) {
	t := s.transformer(req)
	data, err := s.searcher(t)
	if err != nil {
		return nil, err
	}
	return s.normalizer(data), nil
}
