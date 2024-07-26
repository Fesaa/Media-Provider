package yoitsu

import (
	"errors"
	"github.com/Fesaa/Media-Provider/log"
	"os"
)

func removeAll(path, infoHash string, re ...bool) {
	reTry := func() bool {
		if len(re) == 0 {
			return true
		}
		return re[0]
	}()
	log.Trace("removing directory", "dir", path, "infoHash", infoHash, "reTry", reTry)

	err := os.RemoveAll(path)
	if err != nil {
		log.Error("error while deleting directory", "dir", path, "err", err, "infoHash", infoHash)
	}

	stat, err := os.Stat(path)
	log.Trace("os.Stat", "err", func() string {
		if err != nil {
			return err.Error()
		}
		return "<nil>"
	}())
	if !errors.Is(err, os.ErrNotExist) {
		log.Error("error while deleting directory. Directory is still present??", "dir", path, "err", err, "infoHash", infoHash)

		if stat != nil {
			log.Debug("stat info", "size", stat.Size(), "dir", stat.IsDir())

			info, err := os.ReadDir(path)
			if err != nil {
				log.Error("error while deleting directory", "dir", path, "err", err, "infoHash", infoHash)
			}

			if log.IsTraceEnabled() {
				for _, d := range info {
					log.Trace("Found file", "name", d.Name(), "dir", d.IsDir(), "type", d.Type())
				}
			}
		}

		if reTry {
			log.Warn("retrying to remove directory", "dir", path, "infoHash", infoHash)
			removeAll(path, infoHash, false)
		}
	}
}
