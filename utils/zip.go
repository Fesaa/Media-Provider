package utils

import (
	"archive/zip"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func addFileToZip(zipWriter *zip.Writer, filename string, baseDir string) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

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
	slog.Debug("Zipping folder", "path", folderPath, "filename", zipFileName)
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

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
