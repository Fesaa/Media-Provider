package services

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/metadata"
	"github.com/rs/zerolog"
	"time"
)

type MetadataService interface {
	Get() (payload.Metadata, error)
	Update(payload.Metadata) error
}

type metadataService struct {
	db  models.Metadata
	log zerolog.Logger
}

func MetadataServiceProvider(db models.Metadata, log zerolog.Logger) MetadataService {
	return &metadataService{
		db:  db,
		log: log.With().Str("handler", "metadata-service").Logger(),
	}
}

func (m *metadataService) Get() (payload.Metadata, error) {
	metadata, err := m.db.All()
	if err != nil {
		return payload.Metadata{}, err
	}

	return m.metadataFromRows(metadata), nil
}

func (m *metadataService) Update(pl payload.Metadata) error {
	metadata, err := m.db.All()
	if err != nil {
		return err
	}

	metadata = m.metadataUpdateRows(pl, metadata)
	return m.db.Update(metadata)
}

func (m *metadataService) metadataUpdateRows(pl payload.Metadata, rows []models.MetadataRow) []models.MetadataRow {
	newRows := make([]models.MetadataRow, 0)

	for _, row := range rows {
		switch row.Key {
		case models.InstalledVersion:
			if !pl.Version.EqualS(row.Value) {
				row.Value = pl.Version.String()
				newRows = append(newRows, row)
			}
		case models.FirstInstalledVersion:
		case models.InstallDate:
		default:
			continue
		}
	}

	return newRows
}

func (m *metadataService) metadataFromRows(rows []models.MetadataRow) payload.Metadata {
	pl := payload.Metadata{}

	for _, row := range rows {
		switch row.Key {
		case models.InstallDate:
			pl.InstallDate, _ = time.Parse(time.DateTime, row.Value)
		case models.InstalledVersion:
			pl.Version = metadata.SemanticVersion(row.Value)
		case models.FirstInstalledVersion:
			pl.FirstInstalledVersion = row.Value
		}
	}

	return pl
}
