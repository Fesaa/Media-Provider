package providers

import (
	"github.com/Fesaa/Media-Provider/limetorrents"
	"github.com/Fesaa/Media-Provider/subsplease"
	"github.com/Fesaa/Media-Provider/yts"
	"github.com/irevenko/go-nyaa/nyaa"
)

var providers = map[SearchProvider]searchProvider{}

func init() {
	register(NYAA, nyaaTransformer, nyaa.Search, nyaaNormalizer)
	register(LIME, limeTransformer, limetorrents.Search, limeNormalizer)
	register(YTS, ytsTransformer, yts.Search, ytsNormalizer)
	register(SUBSPLEASE, subsPleaseTransformer, subsplease.Search, subsPleaseNormalizer)
}

func register[T, S any](name SearchProvider, t requestTransformerFunc[S], s searchFunc[S, T], n responseNormalizerFunc[T]) {
	providers[name] = &searchProviderImpl[T, S]{
		transformer: t,
		normalizer:  n,
		searcher:    s,
	}
}

type responseNormalizerFunc[T any] func(T) []TorrentInfo
type requestTransformerFunc[S any] func(SearchRequest) S
type searchFunc[S, T any] func(S) (T, error)

type searchProviderImpl[T any, S any] struct {
	transformer requestTransformerFunc[S]
	normalizer  responseNormalizerFunc[T]
	searcher    searchFunc[S, T]
}

func (s *searchProviderImpl[T, S]) Search(req SearchRequest) ([]TorrentInfo, error) {
	t := s.transformer(req)
	data, err := s.searcher(t)
	if err != nil {
		return nil, err
	}
	return s.normalizer(data), nil
}
