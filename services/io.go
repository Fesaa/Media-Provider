package services

import (
	"archive/zip"
	"github.com/rs/zerolog"
	"io"
	"os"
	"path/filepath"
)

type IOService interface {
	// ZipToCbz calls ZipFolder(folderPath, folderPath+".cbz")
	ZipToCbz(folderPath string) error
	ZipFolder(folderPath string, zipFileName string) error
}

type ioService struct {
	log zerolog.Logger
}

func IOServiceProvider(log zerolog.Logger) IOService {
	return &ioService{
		log: log,
	}
}

func (ios *ioService) ZipToCbz(folderPath string) error {
	return ios.ZipFolder(folderPath, folderPath+".cbz")
}

func (ios *ioService) ZipFolder(folderPath string, zipFileName string) error {
	_, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		return err
	}
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return err
	}
	defer func(zipFile *os.File) {
		if err = zipFile.Close(); err != nil {
			ios.log.Warn().Err(err).Msg("failed to close zip file")
		}
	}(zipFile)

	zipWriter := zip.NewWriter(zipFile)
	defer func(zipWriter *zip.Writer) {
		if err = zipWriter.Close(); err != nil {
			ios.log.Warn().Err(err).Msg("failed to close zip writer")
		}
	}(zipWriter)

	err = filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == folderPath {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		return ios.addFileToZip(zipWriter, path, folderPath)
	})

	return err
}

func (ios *ioService) addFileToZip(zipWriter *zip.Writer, filename string, baseDir string) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func(fileToZip *os.File) {
		if err = fileToZip.Close(); err != nil {
			ios.log.Warn().Str("filename", filename).Err(err).Msg("failed to close file")
		}
	}(fileToZip)

	relativePath, err := filepath.Rel(baseDir, filename)
	if err != nil {
		return err
	}

	stat, err := fileToZip.Stat()
	if err != nil {
		return err
	}
	zipHeader, err := zip.FileInfoHeader(stat)
	if err != nil {
		return err
	}
	zipHeader.Name = relativePath
	zipHeader.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(zipHeader)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, fileToZip)
	return err
}
