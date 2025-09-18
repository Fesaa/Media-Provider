package models

type ManualMigration struct {
	Model

	Success bool
	Name    string
}
