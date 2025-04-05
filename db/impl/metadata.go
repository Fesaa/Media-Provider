package impl

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"gorm.io/gorm"
)

type metadata struct {
	db *gorm.DB
}

func Metadata(db *gorm.DB) models.Metadata {
	return &metadata{
		db: db,
	}
}

func (m *metadata) UpdateRow(metadata models.MetadataRow) error {
	return m.db.Where("key = ?", metadata.Key).Save(metadata).Error
}

func (m *metadata) Update(rows []models.MetadataRow) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		for _, row := range rows {
			if err := m.UpdateRow(row); err != nil {
				return err
			}
		}
		return nil
	})
}

func (m *metadata) All() ([]models.MetadataRow, error) {
	var rows []models.MetadataRow
	err := m.db.Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (m *metadata) GetRow(key models.MetadataKey) (*models.MetadataRow, error) {
	var row models.MetadataRow
	err := m.db.Where(&models.MetadataRow{Key: key}).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}
