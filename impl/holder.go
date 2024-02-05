package impl

import (
	"github.com/Fesaa/Media-Provider/models"
)

type HolderImpl struct {
	auth    *AuthImpl
	torrent *TorrentImpl
}

func New() (models.Holder, error) {
	client, err := newTorrent()
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
