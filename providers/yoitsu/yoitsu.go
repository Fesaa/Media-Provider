package yoitsu

import (
	"errors"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"github.com/anacrolix/torrent/types/infohash"
	"github.com/rs/zerolog"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"
)

type yoitsuImpl struct {
	dir         string
	maxTorrents int

	client   *torrent.Client
	torrents utils.SafeMap[string, Torrent]
	baseDirs utils.SafeMap[string, string]
	queue    utils.Queue[payload.QueueStat]

	log zerolog.Logger
}

func New(c *config.Config, log zerolog.Logger) (Yoitsu, error) {
	dir := config.OrDefault(c.GetRootDir(), "temp")

	impl := &yoitsuImpl{
		dir:         dir,
		maxTorrents: c.GetMaxConcurrentTorrents(),

		torrents: utils.NewSafeMap[string, Torrent](),
		baseDirs: utils.NewSafeMap[string, string](),
		queue:    utils.NewQueue[payload.QueueStat](),

		log: log.With().Str("handler", "yoitsu").Logger(),
	}

	opts := storage.NewFileClientOpts{
		ClientBaseDir:   dir,
		TorrentDirMaker: impl.GetTorrentDirFilePathMaker(),
	}
	conf := torrent.NewDefaultClientConfig()
	conf.DefaultStorage = storage.NewFileOpts(opts)
	conf.ListenPort = rand.Intn(65535-49152) + 49152 //nolint:gosec
	conf.DisableIPv6 = c.Downloader.DisableIpv6

	client, err := torrent.NewClient(conf)
	if err != nil {
		return nil, err
	}
	impl.client = client

	go impl.cleaner()
	return impl, nil
}

func (y *yoitsuImpl) Content(id string) services.Content {
	content, ok := y.torrents.Get(id)
	if !ok {
		return nil
	}
	return content
}

func (y *yoitsuImpl) GetBackingClient() *torrent.Client {
	return y.client
}

func (y *yoitsuImpl) GetBaseDir() string {
	return y.dir
}

func (y *yoitsuImpl) Download(req payload.DownloadRequest) error {
	if y.maxTorrents <= 0 {
		return y.addDownload(req)
	}
	if y.torrents.Len() >= y.maxTorrents {
		y.queue.Enqueue(req.ToQueueStat())
		return nil
	}

	y.log.Info().Str("infoHash", req.Id).
		Str("into", req.BaseDir).
		Str("title?", req.TempTitle).
		Msg("downloading torrent")
	return y.addDownload(req)
}

func (y *yoitsuImpl) addDownload(req payload.DownloadRequest) error {
	torrentInfo, nTorrent := y.client.AddTorrentInfoHash(infohash.FromHexString(strings.ToLower(req.Id)))
	if !nTorrent {
		return errors.New("torrent already exists")
	}
	y.processTorrent(torrentInfo, req)
	return nil
}

func (y *yoitsuImpl) processTorrent(torrentInfo *torrent.Torrent, req payload.DownloadRequest) {
	nTorrent := newTorrent(torrentInfo, req, y.log)
	y.torrents.Set(torrentInfo.InfoHash().String(), nTorrent)
	y.baseDirs.Set(torrentInfo.InfoHash().String(), req.BaseDir)
	nTorrent.WaitForInfoAndDownload()
}

func (y *yoitsuImpl) RemoveDownload(req payload.StopRequest) error {
	infoHashString := strings.ToLower(req.Id)
	tor, ok := y.torrents.Get(infoHashString)
	if !ok {
		ok = y.queue.RemoveFunc(func(item payload.QueueStat) bool {
			return item.Id == infoHashString
		})
		if ok {
			y.log.Info().Str("infoHash", infoHashString).Msg("torrent removed from queue")
			return nil
		}

		return errors.New("torrent does not exist, or has already completed")
	}

	// We may assume that it is present as long as the torrent is present
	baseDir, _ := y.baseDirs.Get(infoHashString)

	err := tor.Cancel()
	if err != nil {
		y.log.Error().Err(err).Str("infoHash", infoHashString).Msg("torrent failed to cancel")
	}
	backingTorrent := tor.GetTorrent()
	y.log.Info().
		Str("name", backingTorrent.Name()).
		Str("infoHash", backingTorrent.InfoHash().HexString()).
		Bool("deleteFiles", req.DeleteFiles).
		Int64("downloaded", backingTorrent.BytesCompleted()).
		Int64("total", backingTorrent.Length()).
		Msg("dropping torrent")
	backingTorrent.Drop()

	y.torrents.Delete(infoHashString)
	y.baseDirs.Delete(infoHashString)
	if req.DeleteFiles {
		go y.deleteTorrentFiles(backingTorrent, baseDir)
	} else {
		go y.cleanup(backingTorrent, baseDir)
	}
	y.startNext()
	return nil
}

func (y *yoitsuImpl) startNext() {
	if y.queue.IsEmpty() {
		return
	}

	added := false
	for !added && !y.queue.IsEmpty() {
		item, _ := y.queue.Dequeue()
		if err := y.Download(item.ToDownloadRequest()); err != nil {
			y.log.Warn().Err(err).Msg("error while adding torrent from queue")
			continue
		}
		added = true
	}
}

func (y *yoitsuImpl) cleanup(tor *torrent.Torrent, baseDir string) {
	if tor == nil {
		return
	}
	infoHash := tor.InfoHash().HexString()
	hashDir := path.Join(y.dir, baseDir, infoHash)
	info, err := os.ReadDir(hashDir)
	if err != nil {
		y.log.Error().Err(err).Str("infoHash", infoHash).Str("dir", hashDir).Msg("error reading torrent dir")
		return
	}

	if len(info) == 0 {
		y.log.Warn().Msg("downloaded torrent was empty, removing directory")
		if err = os.Remove(hashDir); err != nil {
			y.log.Error().Err(err).Str("dir", hashDir).Msg("error removing torrent dir")
		}
		return
	}

	firstDirEntry := info[0]
	if len(info) == 1 && firstDirEntry.IsDir() {
		y.log.Debug().Str("infoHash", infoHash).Msg("torrent only has one directory, moving everything up")
		src := path.Join(hashDir, firstDirEntry.Name())
		dest := path.Join(y.dir, baseDir, firstDirEntry.Name())
		if err = os.Rename(src, dest); err != nil {
			y.log.Error().Err(err).Str("src", src).Str("dest", dest).Msg("error while renaming directory")
			return
		}

		if err = os.RemoveAll(hashDir); err != nil {
			y.log.Error().Err(err).Str("dir", hashDir).Msg("error removing torrent dir")
		}
		return
	}

	y.log.Debug().Str("infoHash", infoHash).Msg("torrent downloaded more than one dirEntry, or a file; renaming directory")
	src := hashDir
	dest := path.Join(y.dir, baseDir, tor.Name())
	if err = os.Rename(src, dest); err != nil {
		y.log.Error().Err(err).Str("src", src).Str("dest", dest).Msg("error while renaming directory")
		return
	}
}

func (y *yoitsuImpl) deleteTorrentFiles(tor *torrent.Torrent, baseDir string) {
	if tor == nil {
		return
	}

	infoHash := tor.InfoHash().HexString()
	dir := path.Join(y.dir, baseDir, infoHash)
	y.log.Debug().Str("infoHash", infoHash).Str("dir", dir).Msg("deleting directory")
	if err := os.RemoveAll(dir); err != nil {
		y.log.Error().Err(err).Str("infoHash", infoHash).Str("dir", dir).Msg("error removing torrent dir")
	}
}

func (y *yoitsuImpl) GetRunningTorrents() utils.SafeMap[string, Torrent] {
	return y.torrents
}

func (y *yoitsuImpl) GetQueuedTorrents() []payload.InfoStat {
	return utils.Map(y.queue.Items(), func(item payload.QueueStat) payload.InfoStat {
		return payload.InfoStat{
			Provider:     item.Provider,
			Id:           item.Id,
			ContentState: payload.ContentStateQueued,
			Name:         item.Name,
			Progress:     0,
		}
	})
}

// GetTorrentDirFilePathMaker appending the infohash allows us to always clean up the torrent files on delete
func (y *yoitsuImpl) GetTorrentDirFilePathMaker() storage.TorrentDirFilePathMaker {
	return func(baseDir string, info *metainfo.Info, infoHash metainfo.Hash) string {
		d, ok := y.baseDirs.Get(infoHash.HexString())
		if !ok {
			return path.Join(baseDir, infoHash.HexString())
		}
		return path.Join(baseDir, d, infoHash.HexString())
	}
}

func (y *yoitsuImpl) cleaner() {
	for range time.Tick(time.Second * 5) {
		i := 0
		y.torrents.ForEach(func(s string, m Torrent) {
			tor := m.GetTorrent()
			if tor.BytesCompleted() == tor.Length() && tor.BytesCompleted() > 0 {
				i++
				err := y.RemoveDownload(payload.StopRequest{
					Provider:    -1,
					Id:          s,
					DeleteFiles: false,
				})
				if err != nil {
					y.log.Error().Err(err).Str("file", s).Msg("error while cleaning up torrent")
				}
			}
		})
		if i > 0 {
			y.log.Trace().Int("amount", i).Msg("auto removing torrents")
		}
	}
}
