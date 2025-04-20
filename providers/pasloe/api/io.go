package api

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

func (d *DownloadBase[T]) loadContentOnDisk() {
	d.Log.Debug().Str("dir", d.GetDownloadDir()).Msg("checking content on disk")
	content, err := d.readDirectoryForContent(d.GetDownloadDir())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			d.Log.Trace().Msg("directory not found, fresh download")
		} else {
			d.Log.Warn().Err(err).Msg("unable to check for already downloaded content. Downloading all")
		}
		d.ExistingContent = []Content{}
		return
	}

	d.Log.Trace().Str("content", fmt.Sprintf("%v", content)).Msg("found following content on disk")
	d.ExistingContent = content
}

func (d *DownloadBase[T]) readDirectoryForContent(p string) ([]Content, error) {
	entries, err := d.fs.ReadDir(path.Join(d.Client.GetBaseDir(), p))
	if err != nil {
		return nil, err
	}

	out := make([]Content, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			dirContent, err2 := d.readDirectoryForContent(path.Join(p, entry.Name()))
			if err2 != nil {
				return nil, err2
			}
			out = append(out, dirContent...)
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".cbz") {
			d.Log.Trace().Str("file", entry.Name()).Msg("skipping non content file")
			continue

		}

		matches := d.infoProvider.IsContent(entry.Name())
		if !matches {
			d.Log.Trace().Str("file", entry.Name()).Msg("skipping non content file")
			continue
		}
		d.Log.Trace().Str("file", entry.Name()).Msg("found  content on disk")
		out = append(out, Content{
			Name: entry.Name(),
			Path: path.Join(p, entry.Name()),
		})
	}

	return out, nil
}
