package services

import (
	"archive/zip"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"io"
	"os"
	"path/filepath"
)

type DirectoryService interface {
	// ZipToCbz calls ZipFolder(folderPath, folderPath+".cbz")
	ZipToCbz(folderPath string) error
	ZipFolder(folderPath string, zipFileName string) error

	MoveDirectoryContent(src, dest string) error
}

type directoryService struct {
	log zerolog.Logger
	fs  afero.Afero
}

func DirectoryServiceProvider(log zerolog.Logger, fs afero.Afero) DirectoryService {
	return &directoryService{
		log: log.With().Str("handler", "directory-service").Logger(),
		fs:  fs,
	}
}

func (ds *directoryService) ZipToCbz(folderPath string) error {
	return ds.ZipFolder(folderPath, folderPath+".cbz")
}

func (ds *directoryService) ZipFolder(folderPath string, zipFileName string) error {
	ok, err := ds.fs.DirExists(folderPath)
	if !ok || err != nil {
		return err
	}

	zipFile, err := ds.fs.Create(zipFileName)
	if err != nil {
		return err
	}
	defer func(zipFile afero.File) {
		if err = zipFile.Close(); err != nil {
			ds.log.Warn().Err(err).Msg("failed to close zip file")
		}
	}(zipFile)

	zipWriter := zip.NewWriter(zipFile)
	defer func(zipWriter *zip.Writer) {
		if err = zipWriter.Close(); err != nil {
			ds.log.Warn().Err(err).Msg("failed to close zip writer")
		}
	}(zipWriter)

	err = ds.fs.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == folderPath {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		return ds.addFileToZip(zipWriter, path, folderPath)
	})

	return err
}

func (ds *directoryService) addFileToZip(zipWriter *zip.Writer, filename string, baseDir string) error {
	fileToZip, err := ds.fs.Open(filename)
	if err != nil {
		return err
	}

	defer func(fileToZip afero.File) {
		if err = fileToZip.Close(); err != nil {
			ds.log.Warn().Str("filename", filename).Err(err).Msg("failed to close file")
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

func (ds *directoryService) MoveDirectoryContent(src, dest string) error {
	files, err := ds.fs.ReadDir(src)
	if err != nil {
		return err
	}

	ds.log.Debug().Str("src", src).Str("dest", dest).Int("count", len(files)).
		Msg("Moving directory content")

	for _, f := range files {
		srcPath := filepath.Join(src, f.Name())
		destPath := filepath.Join(dest, f.Name())
		ds.log.Trace().Str("src", srcPath).Str("dest", destPath).Msg("moving file")

		if err = ds.fs.Rename(srcPath, destPath); err != nil {
			return err
		}
	}

	// Remove is on purpose as src should be empty now, not empty should return an error
	// rather than removing left over content
	if err = ds.fs.Remove(src); err != nil {
		return err
	}

	return nil
}
