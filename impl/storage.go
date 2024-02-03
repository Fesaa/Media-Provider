package impl

import (
	"time"

	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/storage"
)

type StorageImpl struct {
	storage storage.Storage
}

func newStorage(storage storage.Storage) *StorageImpl {
	return &StorageImpl{
		storage: storage,
	}
}

func (s *StorageImpl) GetStorage() storage.Storage {
	return s.storage
}

func (s *StorageImpl) Store(key string, storeable models.Storeable, exp time.Duration) error {
	b, err := storeable.ToBytes()
	if err != nil {
		return err
	}
	return s.storage.Set(key, b, exp)
}

func (s *StorageImpl) Load(key string) ([]byte, error) {
	return s.storage.Get(key)
}
