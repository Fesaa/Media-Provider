package core

import "path"

// ContentPath returns the full path for chapter content
func (c *Core[T]) ContentPath(chapter T) string {
	base := path.Join(c.Client.GetBaseDir(), c.GetBaseDir(), c.infoProvider.Title())

	// Check if provider uses volume subdirectories
	if customizer, ok := c.infoProvider.(ContentCustomizer); ok {
		if volumeSubDir := customizer.GetVolumeSubDir(chapter); volumeSubDir != "" {
			base = path.Join(base, volumeSubDir)
		}
	}

	return path.Join(base, c.infoProvider.ContentDir(chapter))
}
