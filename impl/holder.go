package impl

import (
	"database/sql"

	"github.com/Fesaa/Media-Provider/impl/database"
	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/storage"
)

type HolderImpl struct {
	auth     *AuthImpl
	storage  *StorageImpl
	database models.DatabaseProvider
}

func New(storage storage.Storage, pool *sql.DB) (models.Holder, error) {
	db, err := database.NewDatabase(pool)
	if err != nil {
		return nil, err
	}

	return &HolderImpl{
		auth:     newAuth(),
		storage:  newStorage(storage),
		database: db,
	}, nil
}

func (h *HolderImpl) GetAuthProvider() models.AuthProvider {
	return h.auth
}

func (h *HolderImpl) GetStorageProvider() models.StorageProvider {
	return h.storage
}

func (h *HolderImpl) GetDatabaseProvider() models.DatabaseProvider {
	return h.database
}

func (h *HolderImpl) Shutdown() error {
	return h.GetStorageProvider().GetStorage().Close()
}
