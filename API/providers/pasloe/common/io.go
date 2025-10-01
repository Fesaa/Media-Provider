package common

import (
	"context"
	"fmt"
	"path"

	"github.com/Fesaa/Media-Provider/providers/pasloe/core"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
)

func ImageIoTask[C core.Chapter, S core.Series[C]](c *core.Core[C, S], ctx context.Context, log zerolog.Logger, task core.IOTask) error {
	data := task.Data
	ok := false

	p, err := c.UnitOfWork.Preferences.GetPreferences(ctx, c.Req.OwnerId)
	if err == nil && p.ConvertToWebp {
		data, ok = c.ImageService.ConvertToWebp(ctx, task.Data)
	}

	ext := utils.Ternary(ok, ".webp", utils.Ext(task.DTask.Url))
	filePath := path.Join(task.Path, fmt.Sprintf("page %s"+ext, utils.PadInt(task.DTask.Idx, 4)))

	select {
	case <-ctx.Done():
		return nil
	default:
	}

	err = c.Fs.WriteFile(filePath, data, 0755)
	if err == nil {
		return nil
	}

	select {
	case <-ctx.Done():
		log.Debug().Err(err).Msg("ignoring write error due to cancellation")
		return nil
	default:
	}
	log.Error().Err(err).Msg("error writing file")
	return err
}
