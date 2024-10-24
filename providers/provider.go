package providers

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/providers/mangadex"
	"github.com/Fesaa/Media-Provider/providers/webtoon"
	"github.com/Fesaa/Media-Provider/providers/yoitsu/limetorrents"
	"github.com/Fesaa/Media-Provider/providers/yoitsu/subsplease"
	"github.com/Fesaa/Media-Provider/providers/yoitsu/yts"
	"time"
)

var providers = map[models.Provider]provider{}

func init() {
	register(models.SUKEBEI, nyaaTransformer(models.SUKEBEI), nyaaSearch, nyaaNormalizer(models.SUKEBEI), yoitsuDownloader, yoitsuStopper)
	register(models.NYAA, nyaaTransformer(models.NYAA), nyaaSearch, nyaaNormalizer(models.NYAA), yoitsuDownloader, yoitsuStopper)
	register(models.LIME, limeTransformer, limetorrents.Search, limeNormalizer, yoitsuDownloader, yoitsuStopper)
	register(models.YTS, ytsTransformer, yts.Search, ytsNormalizer, yoitsuDownloader, yoitsuStopper)
	register(models.SUBSPLEASE, subsPleaseTransformer, subsplease.Search, subsPleaseNormalizer, yoitsuDownloader, yoitsuStopper)
	register(models.MANGADEX, mangadexTransformer, mangadex.SearchManga, mangadexNormalizer, mangadexDownloader, mangadexStopper)
	register(models.WEBTOON, webtoonTransformer, webtoon.Search, webtoonNormalizer, webToonDownloader, webtoonStopper)
}

func register[T, S any](name models.Provider, t requestTransformerFunc[S], s searchFunc[S, T], n responseNormalizerFunc[T], d downloadFunc, stop stopFunc) {
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
	provider    models.Provider
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
		log.Error("error while searching", "req", fmt.Sprintf("%+v", req), "err", err)
		return nil, err
	}
	return s.normalizer(data), nil
}
