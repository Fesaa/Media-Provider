package services

import (
	"archive/zip"
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"io"
)

type MetadataService interface {
	GetComicInfo(p string) (*comicinfo.ComicInfo, error)
	GetCover(p string) ([]byte, bool)
}

type metaDataService struct {
	log zerolog.Logger
}

func MetadataServiceProvider(log zerolog.Logger) MetadataService {
	return &metaDataService{
		log: log.With().Str("handler", "metadata-service").Logger(),
	}
}

func (m *metaDataService) GetCover(p string) ([]byte, bool) {
	zr, err := zip.OpenReader(p)
	if err != nil {
		return nil, false
	}

	defer zr.Close()

	var ff *zip.File
	var cf *zip.File
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}

		if ff == nil && !utils.ContainsIgnoreCase(f.FileInfo().Name(), "comicinfo") {
			ff = f
		}

		if utils.ContainsIgnoreCase(f.FileInfo().Name(), "cover") {
			cf = f
			break
		}
	}

	if cf == nil {
		m.log.Trace().Str("file", p).Msg("no cover found in archive, falling back to first file")
		cf = ff
	}

	rc, err := cf.Open()
	if err != nil {
		m.log.Error().Err(err).Str("file", p).Msg("failed to open file")
		return nil, false
	}

	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		m.log.Error().Err(err).Str("file", p).Msg("failed to read file")
		return nil, false
	}
	return data, true
}

func (m *metaDataService) GetComicInfo(p string) (*comicinfo.ComicInfo, error) {
	return comicinfo.ReadInZip(p)
}
