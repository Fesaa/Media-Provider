package common

import "github.com/Fesaa/Media-Provider/providers/pasloe/core"

func CbzCleanupFunc[C core.Chapter, S core.Series[C]](c *core.Core[C, S], path string) error {
	if err := c.DirService.ZipToCbz(path); err != nil {
		return err
	}

	return c.Fs.RemoveAll(path)
}
