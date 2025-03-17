package services

import (
	"archive/zip"
	"bytes"
	"errors"
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
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
	fs  afero.Afero
}

func ArchiveServiceProvider(log zerolog.Logger, fs afero.Afero) ArchiveService {
	return &archiveService{
		log: log.With().Str("handler", "archive-service").Logger(),
		fs:  fs,
	}
}

func (a *archiveService) GetComicInfo(archive string) (*comicinfo.ComicInfo, error) {
	rc, err := a.findInArchive(archive, "comicinfo.xml")
	if err != nil {
		return nil, err
	}
	return comicinfo.Read(rc)
}

func (a *archiveService) GetCover(archive string) ([]byte, error) {
	rc, err := a.findInArchive(archive, "cover")
	if err != nil {
		return nil, err
	}
	return io.ReadAll(rc)
}

func (a *archiveService) findInArchive(archive string, match string) (io.Reader, error) {
	f, err := a.fs.Open(archive)
	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}

	defer func(file afero.File) {
		if err = file.Close(); err != nil {
			a.log.Warn().Err(err).Msg("failed to close file")
		}
	}(f)

	reader, err := zip.NewReader(f, fi.Size())
	if err != nil {
		return nil, err
	}

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

	zipFile, err := ciFile.Open()
	if err != nil {
		return nil, err
	}
	defer func(zipFile io.ReadCloser) {
		if err = zipFile.Close(); err != nil {
			a.log.Warn().Err(err).Msg("unable to close zip file")
		}
	}(zipFile)
	b, err := io.ReadAll(zipFile)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}
