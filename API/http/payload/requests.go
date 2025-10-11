package payload

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
)

type SearchRequest struct {
	Provider  []models.Provider `json:"provider" validate:"required,min=1,dive,provider"`
	Query     string            `json:"query"`
	Modifiers utils.SmartMap    `json:"modifiers,omitempty" validate:"dive,keys,required,endkeys,dive,required"`
}

type DownloadRequest struct {
	Provider         models.Provider                `json:"provider" validate:"required,provider"`
	Id               string                         `json:"id" validate:"required"`
	BaseDir          string                         `json:"dir" validate:"required"`
	TempTitle        string                         `json:"title" validate:"required"`
	DownloadMetadata models.DownloadRequestMetadata `json:"downloadMetadata,omitempty"`
	OwnerId          int                            `json:"-"` // Set by MP

	// Internal communication
	IsSubscription bool `json:"-"`
	Sub            *models.Subscription
}

// IncludesMetadataSlice returns true if the request includes metadata for all passed keys, false otherwise
func (r DownloadRequest) IncludesMetadataSlice(keys []string) bool {
	return r.IncludesMetadata(keys...)
}

// IncludesMetadata returns true if the request includes metadata for all passed keys, false otherwise
func (r DownloadRequest) IncludesMetadata(keys ...string) bool {
	return r.DownloadMetadata.Extra.HasKeys(keys...)
}

// GetStrings returns the metadata associated with the key as a slice of strings
// An empty slice will return false
func (r DownloadRequest) GetStrings(key string) ([]string, bool) {
	return r.DownloadMetadata.Extra.GetStrings(key)
}

// GetString returns the metadata associated with the key as a string,
// an empty string is returned if not present
func (r DownloadRequest) GetString(key string, fallback ...string) (string, bool) {
	return r.DownloadMetadata.Extra.GetString(key, fallback...)
}

func (r DownloadRequest) GetStringOrDefault(key string, fallback string) string {
	return r.DownloadMetadata.Extra.GetStringOrDefault(key, fallback)
}

// GetInt returns the metadata associated with the key as an int,
// zero is returned if the value is not present or if conversion failed
func (r DownloadRequest) GetInt(key string, fallback ...int) (int, error) {
	return r.DownloadMetadata.Extra.GetInt(key, fallback...)
}

// GetBool returns the metadata associated with the key as a bool,
// returns true if the value is equal to "true" while ignoring case
func (r DownloadRequest) GetBool(key string, fallback ...bool) bool {
	return r.DownloadMetadata.Extra.GetBool(key, fallback...)
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
	Id     int           `json:"id" validate:"required"`
	Name   string        `json:"name" validate:"required"`
	Email  string        `json:"email" validate:"required"`
	ApiKey string        `json:"apiKey,omitempty"`
	Roles  []models.Role `json:"roles,omitempty"`
}
type UpdateUserRequest struct {
	Name  string `json:"username" validate:"required"`
	Email string `json:"email"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"oldPassword" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required"`
}

type ResetPasswordRequest struct {
	Key      string `json:"key" validate:"required"`
	Password string `json:"password" validate:"required"`
}
