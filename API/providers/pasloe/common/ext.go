package common

import "github.com/Fesaa/Media-Provider/providers/pasloe/core"

func CbzExt[C core.Chapter, S core.Series[C]]() core.Ext[C, S] {
	return core.Ext[C, S]{
		IoTaskFunc:         ImageIoTask[C, S],
		ContentCleanupFunc: CbzCleanupFunc[C, S],
		IsContentFunc:      IsCbz,
		VolumeFunc:         GetVolumeFromComicInfo[C, S],
	}
}
