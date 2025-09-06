package models

import "time"

// Model is a gorm.Model but without DeletedAt
type Model struct {
	ID        int `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
