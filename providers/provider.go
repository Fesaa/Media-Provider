package providers

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/limetorrents"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/mangadex"
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/Fesaa/Media-Provider/subsplease"
	"github.com/Fesaa/Media-Provider/yts"
	"time"
)

var providers = map[config.Provider]provider{}

func init() {
	register(config.SUKEBEI, nyaaTransformer(config.SUKEBEI), nyaaSearch, nyaaNormalizer(config.SUKEBEI), yoitsuDownloader, yoitsuStopper)
	register(config.NYAA, nyaaTransformer(config.NYAA), nyaaSearch, nyaaNormalizer(config.NYAA), yoitsuDownloader, yoitsuStopper)
	register(config.LIME, limeTransformer, limetorrents.Search, limeNormalizer, yoitsuDownloader, yoitsuStopper)
	register(config.YTS, ytsTransformer, yts.Search, ytsNormalizer, yoitsuDownloader, yoitsuStopper)
	register(config.SUBSPLEASE, subsPleaseTransformer, subsplease.Search, subsPleaseNormalizer, yoitsuDownloader, yoitsuStopper)
	register(config.MANGADEX, mangadexTransformer, mangadex.SearchManga, mangadexNormalizer, mangadexDownloader, mangadexStopper)
}

func register[T, S any](name config.Provider, t requestTransformerFunc[S], s searchFunc[S, T], n responseNormalizerFunc[T], d downloadFunc, stop stopFunc) {
	providers[name] = &providerImpl[T, S]{
		transformer: t,
		normalizer:  n,
		searcher:    s,
		downloader:  d,
		stopper:     stop,
		provider:    name,
	}
}

type responseNormalizerFunc[T any] func(T) []Info
type requestTransformerFunc[S any] func(payload.SearchRequest) S
type searchFunc[S, T any] func(S) (T, error)
type downloadFunc func(payload.DownloadRequest) error
type stopFunc func(payload.StopRequest) error

type providerImpl[T any, S any] struct {
	transformer requestTransformerFunc[S]
	normalizer  responseNormalizerFunc[T]
	searcher    searchFunc[S, T]
	downloader  downloadFunc
	stopper     stopFunc
	provider    config.Provider
}

func (s *providerImpl[T, S]) Download(req payload.DownloadRequest) error {
	return s.downloader(req)
}

func (s *providerImpl[T, S]) Stop(req payload.StopRequest) error {
	return s.stopper(req)
}

func (s *providerImpl[T, S]) Search(req payload.SearchRequest) ([]Info, error) {
	t := s.transformer(req)

	start := time.Now()
	data, err := s.searcher(t)
	since := time.Since(start)

	log.Debug("Search done", "duration", since, "provider", s.provider, "request", fmt.Sprintf("%+v", req))
	if since > time.Second*1 {
		log.Warn("Searching took more than one second", "duration", since, "provider", s.provider)
	}

	if err != nil {
		log.Debug("error while searching", "req", fmt.Sprintf("%+v", req), "err", err)
		return nil, err
	}
	return s.normalizer(data), nil
}
