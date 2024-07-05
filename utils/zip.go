package utils

import (
	"archive/zip"
	"github.com/Fesaa/Media-Provider/log"
	"io"
	"os"
	"path/filepath"
)

func addFileToZip(zipWriter *zip.Writer, filename string, baseDir string) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func(fileToZip *os.File) {
		if err = fileToZip.Close(); err != nil {
			log.Warn("failed to close file", "filename", filename, "error", err)
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

func ZipFolder(folderPath string, zipFileName string) error {
	log.Trace("zipping folder", "path", folderPath, "filename", zipFileName)
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return err
	}
	defer func(zipFile *os.File) {
		if err = zipFile.Close(); err != nil {
			log.Warn("failed to close zip file", "filename", zipFileName, "error", err)
		}
	}(zipFile)

	zipWriter := zip.NewWriter(zipFile)
	defer func(zipWriter *zip.Writer) {
		if err = zipWriter.Close(); err != nil {
			log.Warn("failed to close zip writer", "error", err)
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

		return addFileToZip(zipWriter, path, folderPath)
	})

	return err
}
