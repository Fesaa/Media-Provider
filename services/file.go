package services

import (
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
)

type FileService interface{}

type fileService struct {
	log zerolog.Logger
	fs  afero.Fs
}

func FileServiceProvider(log zerolog.Logger, fs afero.Fs) FileService {
	return fileService{
		log: log,
		fs:  fs,
	}
}
