package core

import (
	"context"
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/utils"
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

		content, matches := c.IsContent(entry.Name())
		if !matches {
			c.Log.Trace().Str("file", entry.Name()).Msg("skipping non content file")
			continue
		}
		c.Log.Trace().Str("file", entry.Name()).Msg("found  content on disk")
		out = append(out, Content{
			Name:    entry.Name(),
			Path:    path.Join(p, entry.Name()),
			Volume:  content.Volume,
			Chapter: content.Chapter,
		})
	}

	return out, nil
}

// StartIOWorkers starts cap(IOCh) IOWorker threads
func (c *Core[C, S]) StartIOWorkers(ctx context.Context) {
	for worker := range cap(c.IOWorkCh) {
		go c.IOWorker(ctx, fmt.Sprintf("IOWorker#%d", worker))
	}
}

// IOWorker reads from IOCh; converts to webp and writes to disk
func (c *Core[C, S]) IOWorker(ctx context.Context, id string) {
	c.IoWg.Add(1)
	defer c.IoWg.Done()

	log := c.Log.With().Str("IOWorker#", id).Logger()

	for task := range c.IOWorkCh {
		select {
		case <-ctx.Done():
			return
		default:
		}

		data, ok := c.imageService.ConvertToWebp(task.data)

		ext := utils.Ternary(ok, ".webp", utils.Ext(task.dTask.url))
		filePath := path.Join(task.path, fmt.Sprintf("page %s"+ext, utils.PadInt(task.dTask.idx, 4)))

		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := c.fs.WriteFile(filePath, data, 0755); err != nil {
			select {
			case <-ctx.Done():
				log.Debug().Err(err).Msg("ignoring write error due to cancellation")
				return
			default:
			}
			log.Error().Err(err).Msg("error writing file")
			c.abortDownload(fmt.Errorf("error writing file %s: %w", filePath, err))
			return
		}
	}
}
