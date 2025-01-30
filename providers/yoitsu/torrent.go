package yoitsu

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/anacrolix/torrent"
	"github.com/rs/zerolog"
	"os"
	"path"
	"slices"
	"strings"
	"time"
)

// torrentImpl wrapper around the torrent.Torrent struct
// Providers some specific functionality
type torrentImpl struct {
	t      *torrent.Torrent
	log    zerolog.Logger
	client Yoitsu

	req       payload.DownloadRequest
	key       string
	baseDir   string
	tempTitle string
	provider  models.Provider
	state     payload.ContentState

	userFilter []string

	ctx    context.Context
	cancel context.CancelFunc

	lastTime time.Time
	lastRead int64
}

func newTorrent(t *torrent.Torrent, req payload.DownloadRequest, log zerolog.Logger, client Yoitsu) Torrent {
	tor := &torrentImpl{
		t:         t,
		client:    client,
		key:       t.InfoHash().HexString(),
		req:       req,
		baseDir:   req.BaseDir,
		tempTitle: req.TempTitle,
		provider:  req.Provider,
		lastTime:  time.Now(),
		lastRead:  0,
		state:     payload.ContentStateQueued,
	}

	tor.log = log.With().Str("infoHash", tor.key).Logger()
	return tor
}

func (t *torrentImpl) Id() string {
	return t.key
}

func (t *torrentImpl) Title() string {
	if t.t.Info() != nil {
		return t.t.Info().BestName()
	}
	return t.tempTitle
}

func (t *torrentImpl) Provider() models.Provider {
	return t.provider
}

func (t *torrentImpl) State() payload.ContentState {
	return t.state
}

func (t *torrentImpl) Message(msg payload.Message) (payload.Message, error) {
	var jsonData []byte
	var err error
	switch msg.MessageType {
	case payload.MessageListContent:
		jsonData, err = json.Marshal(t.ContentList())
	case payload.SetToDownload:
		err = t.SetUserFiltered(msg.Data)
	case payload.StartDownload:
		err = t.MarkReady()
	default:
		err = services.ErrUnknownMessageType
	}

	if err != nil {
		return payload.Message{}, err
	}

	return payload.Message{
		Provider:    t.Provider(),
		ContentId:   t.key,
		MessageType: msg.MessageType,
		Data:        jsonData,
	}, nil
}

func (t *torrentImpl) MarkReady() error {
	if t.state != payload.ContentStateWaiting {
		return services.ErrWrongState
	}
	if t.client.CanStartNext() {
		go t.StartDownload()
		return nil
	}

	t.state = payload.ContentStateReady
	return nil
}

func (t *torrentImpl) SetUserFiltered(data json.RawMessage) error {
	if t.state != payload.ContentStateWaiting &&
		t.state != payload.ContentStateReady {
		return services.ErrWrongState
	}

	var filter []string
	if err := json.Unmarshal(data, &filter); err != nil {
		return err
	}

	t.userFilter = filter
	return nil
}

func (t *torrentImpl) ContentList() []payload.ListContentData {
	if t.t.Info() == nil {
		return nil
	}

	paths := utils.Map(t.t.Files(), func(file *torrent.File) []string {
		return strings.Split(file.Path(), "/")
	})

	return t.buildTree(paths)
}

func (t *torrentImpl) buildTree(paths [][]string, depths ...int) []payload.ListContentData {
	depth := utils.OrDefault(depths, 0)
	var tree []payload.ListContentData
	pathByFirstDir := utils.GroupBy(paths, func(v []string) string {
		if depth >= len(v) {
			return ""
		}
		return v[depth]
	})

	for dir, subPaths := range pathByFirstDir {
		if dir == "" {
			continue
		}

		if len(subPaths[0]) == depth+1 {
			id := path.Join(subPaths[0]...)
			tree = append(tree, payload.ListContentData{
				Label:        dir,
				Selected:     len(t.userFilter) == 0 || slices.Contains(t.userFilter, id),
				SubContentId: id,
			})
			continue
		}

		children := t.buildTree(subPaths, depth+1)
		tree = append(tree, payload.ListContentData{
			Label:    dir,
			Children: children,
		})
	}

	if len(tree) == 1 {
		return tree[0].Children
	}

	return tree
}

func (t *torrentImpl) GetTorrent() *torrent.Torrent {
	return t.t
}

func (t *torrentImpl) LoadInfo() {
	if t.cancel != nil {
		t.log.Debug().Msg("already loading info")
		return
	}

	t.state = payload.ContentStateLoading
	ctx, cancel := context.WithCancel(context.Background())
	t.ctx = ctx
	t.cancel = cancel
	t.log.Trace().Msg("loading torrent info")
	select {
	case <-t.ctx.Done():
		return
	case <-t.t.GotInfo():
		t.log = t.log.With().Str("name", t.t.Info().BestName()).Logger()
	}

	t.log.Info().Msg("torrent has downloaded all info")

	t.state = utils.Ternary(t.req.DownloadMetadata.StartImmediately,
		payload.ContentStateReady,
		payload.ContentStateWaiting)
}

func (t *torrentImpl) StartDownload() {
	t.log.Info().Str("infoHash", t.key).
		Str("into", t.GetDownloadDir()).
		Str("title", t.Title()).
		Msg("downloading torrent")
	t.state = payload.ContentStateDownloading

	if len(t.userFilter) == 0 {
		t.t.DownloadAll()
		return
	}

	for _, file := range t.t.Files() {
		if slices.Contains(t.userFilter, file.Path()) {
			file.SetPriority(torrent.PiecePriorityNormal)
		} else {
			file.SetPriority(torrent.PiecePriorityNone)
		}
	}
}

func (t *torrentImpl) Cancel() error {
	t.log.Trace().Msg("cancelling torrent")
	if t.cancel == nil {
		return fmt.Errorf("torrent is not downloading")
	}
	t.cancel()
	return nil
}

func (t *torrentImpl) GetDownloadDir() string {
	return path.Join(t.baseDir, t.key)
}

func (t *torrentImpl) GetInfo() payload.InfoStat {
	c := t.t.Stats().BytesReadData
	bytesRead := c.Int64()
	bytesDiff := bytesRead - t.lastRead
	timeDiff := max(time.Since(t.lastTime).Seconds(), 1)
	speed := int64(float64(bytesDiff) / timeDiff)
	t.lastRead = bytesRead
	t.lastTime = time.Now()

	size := t.size()
	return payload.InfoStat{
		Provider:     t.provider,
		Id:           t.key,
		ContentState: t.state,
		Name:         t.Title(),
		Size:         utils.BytesToSize(float64(size)),
		Downloading:  t.state == payload.ContentStateDownloading,
		Progress:     utils.Percent(t.t.BytesCompleted(), size),
		Estimated: func() *int64 {
			if speed == 0 {
				return nil
			}
			es := (size - bytesRead) / speed
			return &es
		}(),
		SpeedType:   payload.BYTES,
		Speed:       speed,
		DownloadDir: t.GetDownloadDir(),
	}
}

// Cleanup is needed in case of user filtered content. While the pieces priority is set to none,
// the file is still added as zero bytes, and the pieces closeby pieces with priority have some overflow
func (t *torrentImpl) Cleanup(root string) {
	if len(t.userFilter) == 0 {
		return
	}

	for _, file := range t.t.Files() {
		if slices.Contains(t.userFilter, file.Path()) {
			continue
		}

		filePath := path.Join(root, file.Path())
		t.log.Debug().Str("path", filePath).Msg("removing file, as it wasn't wanted")
		if err := os.Remove(filePath); err != nil {
			t.log.Error().Str("path", filePath).Err(err).Msg("failed to remove file")
		}
	}
}

func (t *torrentImpl) IsDone() bool {
	if t.state != payload.ContentStateDownloading {
		return false
	}

	if len(t.userFilter) == 0 {
		return t.t.Length() == t.t.BytesCompleted()
	}

	// Since we have to check with >= below (overflow), lets make sure every file is completely downloaded
	// and that we do cross the threshold with overflow, and leave a wanted file corrupted
	for _, file := range t.t.Files() {
		if !slices.Contains(t.userFilter, file.Path()) {
			continue
		}

		if file.BytesCompleted() != file.Length() {
			return false
		}
	}

	return t.t.BytesCompleted() >= t.size()
}

func (t *torrentImpl) size() int64 {
	if t.t.Info() == nil {
		return 0
	}

	if len(t.userFilter) == 0 {
		return t.t.Length()
	}

	var size int64
	for _, file := range t.t.Files() {
		if slices.Contains(t.userFilter, file.Path()) {
			size += file.Length()
		}
	}
	return size
}
