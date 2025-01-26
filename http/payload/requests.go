package payload

import (
	"github.com/Fesaa/Media-Provider/db/models"
)

type SearchRequest struct {
	Provider  []models.Provider   `json:"provider" validate:"required,min=1,dive,provider"`
	Query     string              `json:"query"`
	Modifiers map[string][]string `json:"modifiers,omitempty" validate:"dive,keys,required,endkeys,dive,required"`
}

type DownloadRequest struct {
	Provider  models.Provider `json:"provider" validate:"required,provider"`
	Id        string          `json:"id" validate:"required"`
	BaseDir   string          `json:"dir" validate:"required"`
	TempTitle string          `json:"title" validate:"required"`
}

type StopRequest struct {
	Provider    models.Provider `json:"provider" validate:"required,provider"`
	Id          string          `json:"id" validate:"required"`
	DeleteFiles bool            `json:"delete" validate:"required"`
}

type ListDirsRequest struct {
	Dir       string `json:"dir"`
	ShowFiles bool   `json:"files"`
}

type LoginRequest struct {
	UserName string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Remember bool   `json:"remember,omitempty"`
}

type LoginResponse struct {
	Id          uint   `json:"id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Token       string `json:"token" validate:"required"`
	ApiKey      string `json:"apiKey,omitempty"`
	Permissions int    `json:"permissions" validate:"gte=0"`
}

type UpdatePasswordRequest struct {
	Password string `json:"password" validate:"required"`
}

type SwapPageRequest struct {
	Id1 uint `json:"id1" validate:"required"`
	Id2 uint `json:"id2" validate:"required,diff=Id1"`
}

type ResetPasswordRequest struct {
	Key      string `json:"key" validate:"required"`
	Password string `json:"password" validate:"required"`
}
