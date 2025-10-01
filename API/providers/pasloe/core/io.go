package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/Fesaa/Media-Provider/internal/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
	entries, err := c.Fs.ReadDir(path.Join(c.Client.GetBaseDir(), p))
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

	ctx, span := tracing.TracerPasloe.Start(ctx, tracing.SpanPasloeIOWorker,
		trace.WithAttributes(attribute.String("worker.id", id)))
	defer span.End()

	log := c.Log.With().Str("IOWorker#", id).Logger()

	for task := range c.IOWorkCh {
		select {
		case <-ctx.Done():
			return
		default:
		}

		err := c.impl.CoreExt().IoTaskFunc(c, ctx, log, task)
		if err != nil {
			c.abortDownload(fmt.Errorf("unable to run task '%v': %w", task.Path, err))
			return
		}
	}
}
