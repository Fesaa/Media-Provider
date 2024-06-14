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
	register(config.NYAA, nyaaTransformer, nyaa.Search, nyaaNormalizer, yoitsuDownloader, yoitsuStopper)
	register(config.LIME, limeTransformer, limetorrents.Search, limeNormalizer, yoitsuDownloader, yoitsuStopper)
	register(config.YTS, ytsTransformer, yts.Search, ytsNormalizer, yoitsuDownloader, yoitsuStopper)
	register(config.SUBSPLEASE, subsPleaseTransformer, subsplease.Search, subsPleaseNormalizer, yoitsuDownloader, yoitsuStopper)
}

func register[T, S any](name config.Provider, t requestTransformerFunc[S], s searchFunc[S, T], n responseNormalizerFunc[T], d downloadFunc, stop stopFunc) {
	providers[name] = &providerImpl[T, S]{
		transformer: t,
		normalizer:  n,
		searcher:    s,
		downloader:  d,
		stopper:     stop,
	}
}

type responseNormalizerFunc[T any] func(T) []TorrentInfo
type requestTransformerFunc[S any] func(SearchRequest) S
type searchFunc[S, T any] func(S) (T, error)
type downloadFunc func(DownloadRequest) error
type stopFunc func(StopRequest) error

type providerImpl[T any, S any] struct {
	transformer requestTransformerFunc[S]
	normalizer  responseNormalizerFunc[T]
	searcher    searchFunc[S, T]
	downloader  downloadFunc
	stopper     stopFunc
}

func (s *providerImpl[T, S]) Download(req DownloadRequest) error {
	return s.downloader(req)
}

func (s *providerImpl[T, S]) Stop(req StopRequest) error {
	return s.stopper(req)
}

func (s *providerImpl[T, S]) Search(req SearchRequest) ([]TorrentInfo, error) {
	t := s.transformer(req)
	data, err := s.searcher(t)
	if err != nil {
		return nil, err
	}
	return s.normalizer(data), nil
}
