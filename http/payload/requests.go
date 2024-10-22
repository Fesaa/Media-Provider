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
	Id          int64  `json:"id"`
	Token       string `json:"token"`
	ApiKey      string `json:"apiKey,omitempty"`
	Permissions int    `json:"permissions"`
}

type UpdatePasswordRequest struct {
	Password string `json:"password"`
}

type SwapPageRequest struct {
	Id1 int64 `json:"id1"`
	Id2 int64 `json:"id2"`
}

type ResetPasswordRequest struct {
	Key      string `json:"key"`
	Password string `json:"password"`
}
