package services

import (
	"archive/zip"
	"errors"
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/rs/zerolog"
	"io"
	"strings"
)

var ErrNoMatch = errors.New("zip file does not the wanted content")

type ArchiveService interface {
	GetComicInfo(archive string) (*comicinfo.ComicInfo, error)
	GetCover(archive string) ([]byte, error)
}

type archiveService struct {
	log zerolog.Logger
}

func ArchiveServiceProvider(log zerolog.Logger) ArchiveService {
	return &archiveService{
		log: log.With().Str("handler", "archive-service").Logger(),
	}
}

func (a *archiveService) GetComicInfo(archive string) (*comicinfo.ComicInfo, error) {
	rc, err := a.findInArchive(archive, "comicinfo.xml")
	if err != nil {
		return nil, err
	}
	defer func(rc io.ReadCloser) {
		if err = rc.Close(); err != nil {
			a.log.Warn().Err(err).Msg("failed to close file reader")
		}
	}(rc)

	return comicinfo.Read(rc)
}

func (a *archiveService) GetCover(archive string) ([]byte, error) {
	rc, err := a.findInArchive(archive, "cover")
	if err != nil {
		return nil, err
	}
	defer func(rc io.ReadCloser) {
		if err = rc.Close(); err != nil {
			a.log.Warn().Err(err).Msg("failed to close file reader")
		}
	}(rc)

	return io.ReadAll(rc)
}

func (a *archiveService) findInArchive(archive string, match string) (io.ReadCloser, error) {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return nil, err
	}
	defer func(reader *zip.ReadCloser) {
		if err = reader.Close(); err != nil {
			a.log.Warn().Err(err).Msg("unable to close zip reader")
		}
	}(reader)

	var ciFile *zip.File
	for _, file := range reader.File {
		if strings.Contains(strings.ToLower(file.Name), strings.ToLower(match)) {
			ciFile = file
			break
		}
	}

	if ciFile == nil {
		return nil, ErrNoMatch
	}

	f, err := ciFile.Open()
	if err != nil {
		return nil, err
	}
	return f, nil
}
