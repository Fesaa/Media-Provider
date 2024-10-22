package payload

import (
	"github.com/Fesaa/Media-Provider/db/models"
)

type SearchRequest struct {
	Provider  []models.Provider   `json:"provider"`
	Query     string              `json:"query"`
	Modifiers map[string][]string `json:"modifiers,omitempty"`
}

type DownloadRequest struct {
	Provider  models.Provider `json:"provider"`
	Id        string          `json:"id"`
	BaseDir   string          `json:"dir"`
	TempTitle string          `json:"title"`
}

type StopRequest struct {
	Provider    models.Provider `json:"provider"`
	Id          string          `json:"id"`
	DeleteFiles bool            `json:"delete"`
}

type ListDirsRequest struct {
	Dir       string `json:"dir"`
	ShowFiles bool   `json:"files"`
}

type LoginRequest struct {
	UserName string `json:"username"`
	Password string `json:"password"`
	Remember bool   `json:"remember,omitempty"`
}

type LoginResponse struct {
	Token  string `json:"token"`
	ApiKey string `json:"apiKey,omitempty"`
}

type UpdatePasswordRequest struct {
	Password string `json:"password"`
}

type MovePageRequest struct {
	OldIndex int `json:"oldIndex"`
	NewIndex int `json:"newIndex"`
}
