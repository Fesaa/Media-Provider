/*
	Package comicinfo

# MIT License

# Copyright (c) 2023 Felipe Martin

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

https://github.com/fmartingr/go-comicinfo/blob/latest/io.go
*/
package comicinfo

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var ErrNoComicInfo = errors.New("zip file does not contain comic info")

// ReadInZip tries to find a comicinfo.xml file inside a zip archive
func ReadInZip(path string) (*ComicInfo, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var ciFile *zip.File
	for _, file := range reader.File {
		if strings.ToLower(file.Name) == "comicinfo.xml" {
			ciFile = file
			break
		}
	}

	if ciFile == nil {
		return nil, ErrNoComicInfo
	}

	f, err := ciFile.Open()
	if err != nil {
		return nil, err
	}

	defer f.Close()
	return Read(f)
}

// Read reads the ComicInfo spec from the specified reader
func Read(r io.Reader) (*ComicInfo, error) {
	var ci ComicInfo
	err := xml.NewDecoder(r).Decode(&ci)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling ComicInfo: %w", err)
	}

	return &ci, nil
}

// Write writes the ComicInfo spec to the specified writter
func Write(ci *ComicInfo, w io.Writer) error {
	contents, err := xml.Marshal(ci)
	if err != nil {
		return fmt.Errorf("error marshalling ComicInfo: %w", err)
	}

	_, err = w.Write(bytes.Join([][]byte{xmlHeader, contents}, []byte("")))
	if err != nil {
		return fmt.Errorf("error writing ComicInfo to file: %w", err)
	}
	return nil
}

// Open reads the ComicInfo spec from the specified path.
func Open(path string) (*ComicInfo, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var ci ComicInfo
	err = xml.Unmarshal(f, &ci)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling ComicInfo: %w", err)
	}

	return &ci, nil
}

// Save writes the ComicInfo spec to the specified path.
func Save(ci *ComicInfo, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer f.Close()

	err = Write(ci, f)
	if err != nil {
		return fmt.Errorf("error writing ComicInfo to file: %w", err)
	}

	return nil
}
