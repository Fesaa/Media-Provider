package payload

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"strconv"
	"strings"
)

type SearchRequest struct {
	Provider  []models.Provider   `json:"provider" validate:"required,min=1,dive,provider"`
	Query     string              `json:"query"`
	Modifiers map[string][]string `json:"modifiers,omitempty" validate:"dive,keys,required,endkeys,dive,required"`
}

type DownloadRequest struct {
	Provider         models.Provider                `json:"provider" validate:"required,provider"`
	Id               string                         `json:"id" validate:"required"`
	BaseDir          string                         `json:"dir" validate:"required"`
	TempTitle        string                         `json:"title" validate:"required"`
	DownloadMetadata models.DownloadRequestMetadata `json:"downloadMetadata,omitempty"`

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
	for _, key := range keys {
		if _, ok := r.DownloadMetadata.Extra[key]; !ok {
			return false
		}
	}
	return true
}

// GetStrings returns the metadata associated with the key as a slice of strings
// An empty slice will return false
func (r DownloadRequest) GetStrings(key string) ([]string, bool) {
	values, ok := r.DownloadMetadata.Extra[key]
	values = utils.Filter(values, func(s string) bool {
		return len(s) > 0
	})
	return values, ok && len(values) > 0
}

// GetString returns the metadata associated with the key as a string,
// an empty string is returned if not present
func (r DownloadRequest) GetString(key string, fallback ...string) (string, bool) {
	values, ok := r.GetStrings(key)
	if !ok {
		if len(fallback) > 0 {
			return fallback[0], true
		}
		return "", false
	}

	return values[0], true
}

func (r DownloadRequest) GetStringOrDefault(key string, fallback string) string {
	s, ok := r.GetString(key)
	if !ok {
		return fallback
	}
	return s
}

// GetInt returns the metadata associated with the key as an int,
// zero is returned if the value is not present or if conversion failed
func (r DownloadRequest) GetInt(key string, fallback ...int) (int, error) {
	val, ok := r.GetString(key)
	if !ok {
		if len(fallback) > 0 {
			return fallback[0], nil
		}
		return 0, nil
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		if len(fallback) > 0 {
			return fallback[0], nil
		}
		return 0, err
	}
	return i, nil
}

// GetBool returns the metadata associated with the key as a bool,
// returns true if the value is equal to "true" while ignoring case
func (r DownloadRequest) GetBool(key string, fallback ...bool) bool {
	val, ok := r.GetString(key)
	if !ok {
		if len(fallback) > 0 {
			return fallback[0]
		}
		return false
	}
	return strings.ToLower(val) == "true"
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
	Email       string `json:"email" validate:"required"`
	Token       string `json:"token" validate:"required"`
	ApiKey      string `json:"apiKey,omitempty"`
	Permissions int    `json:"permissions" validate:"gte=0"`
}
type UpdateUserRequest struct {
	Name  string `json:"username" validate:"required"`
	Email string `json:"email"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"oldPassword" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required"`
}

type SwapPageRequest struct {
	Id1 uint `json:"id1" validate:"required"`
	Id2 uint `json:"id2" validate:"required,diff=Id1"`
}

type ResetPasswordRequest struct {
	Key      string `json:"key" validate:"required"`
	Password string `json:"password" validate:"required"`
}
