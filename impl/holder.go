package impl

import (
	"math/rand"
	"os"

	"github.com/Fesaa/Media-Provider/models"
	"github.com/anacrolix/torrent"
)

type HolderImpl struct {
	auth    *AuthImpl
	torrent *TorrentImpl
}

func New() (models.Holder, error) {
	conf := torrent.NewDefaultClientConfig()
	conf.DataDir = os.Getenv("TORRENT_DIR")
	conf.ListenPort = rand.Intn(65535-49152) + 49152

	client, err := newTorrent(conf)
	if err != nil {
		return nil, err
	}

	return &HolderImpl{
		auth:    newAuth(),
		torrent: client,
	}, nil
}

func (h *HolderImpl) GetAuthProvider() models.AuthProvider {
	return h.auth
}

func (h *HolderImpl) GetTorrentProvider() models.TorrentProvider {
	return h.torrent
}

func (h *HolderImpl) Shutdown() error {
	return nil
}
