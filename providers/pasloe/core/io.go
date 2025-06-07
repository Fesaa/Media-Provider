package core

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

func (c *Core[C, S]) loadContentOnDisk() {
	c.Log.Debug().Str("dir", c.GetDownloadDir()).Msg("checking content on disk")
	content, err := c.readDirectoryForContent(c.GetDownloadDir())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.Log.Trace().Msg("directory not found, fresh download")
		} else {
			c.Log.Warn().Err(err).Msg("unable to check for already downloaded content. Downloading all")
		}
		c.ExistingContent = []Content{}
		return
	}

	c.Log.Trace().Str("content", fmt.Sprintf("%v", content)).Msg("found following content on disk")
	c.ExistingContent = content
}

func (c *Core[C, S]) readDirectoryForContent(p string) ([]Content, error) {
	entries, err := c.fs.ReadDir(path.Join(c.Client.GetBaseDir(), p))
	if err != nil {
		return nil, err
	}

	out := make([]Content, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			dirContent, err2 := c.readDirectoryForContent(path.Join(p, entry.Name()))
			if err2 != nil {
				return nil, err2
			}
			out = append(out, dirContent...)
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".cbz") {
			c.Log.Trace().Str("file", entry.Name()).Msg("skipping non content file")
			continue

		}

		matches := c.IsContent(entry.Name())
		if !matches {
			c.Log.Trace().Str("file", entry.Name()).Msg("skipping non content file")
			continue
		}
		c.Log.Trace().Str("file", entry.Name()).Msg("found  content on disk")
		out = append(out, Content{
			Name: entry.Name(),
			Path: path.Join(p, entry.Name()),
		})
	}

	return out, nil
}
