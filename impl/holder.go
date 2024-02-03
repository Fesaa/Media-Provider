package impl

import (
	"database/sql"
	"math/rand"

	"github.com/Fesaa/Media-Provider/impl/database"
	"github.com/Fesaa/Media-Provider/models"
	"github.com/anacrolix/torrent"
)

type HolderImpl struct {
	auth     *AuthImpl
	database models.DatabaseProvider
	torrent  *TorrentImpl
}

func New(pool *sql.DB) (models.Holder, error) {
	db, err := database.NewDatabase(pool)
	if err != nil {
		return nil, err
	}
	conf := torrent.NewDefaultClientConfig()
	conf.DataDir = "temp"
	conf.ListenPort = rand.Intn(65535-49152) + 49152

	client, err := newTorrent(conf)
	if err != nil {
		return nil, err
	}

	return &HolderImpl{
		auth:     newAuth(),
		database: db,
		torrent:  client,
	}, nil
}

func (h *HolderImpl) GetAuthProvider() models.AuthProvider {
	return h.auth
}

func (h *HolderImpl) GetDatabaseProvider() models.DatabaseProvider {
	return h.database
}

func (h *HolderImpl) GetTorrentProvider() models.TorrentProvider {
	return h.torrent
}

func (h *HolderImpl) Shutdown() error {
	return nil
}
