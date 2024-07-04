/*
MIT License

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
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

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

// Read reads the ComicInfo spec from the specified path.
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

	err = Write(ci, f)
	if err != nil {
		return fmt.Errorf("error writing ComicInfo to file: %w", err)
	}

	return nil
}
