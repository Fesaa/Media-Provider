package models

import (
	"time"

	"github.com/gofiber/storage"
)

type StorageProvider interface {
	// Returns the underlying gofiber storage interface
	GetStorage() storage.Storage
	// Wrapper around set, which converts the Storeable to bytes
	Store(key string, s Storeable, exp time.Duration) error
	// Wrapper around get, cannot have generics in interface so converting to bytes must be done manually
	Load(key string) ([]byte, error)
}

type Storeable interface {
	ToBytes() ([]byte, error)
}
