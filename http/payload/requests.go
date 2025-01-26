package payload

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"strconv"
	"strings"
)

type SearchRequest struct {
	Provider  []models.Provider   `json:"provider" validate:"required,min=1,dive,provider"`
	Query     string              `json:"query"`
	Modifiers map[string][]string `json:"modifiers,omitempty" validate:"dive,keys,required,endkeys,dive,required"`
}

type DownloadRequest struct {
	Provider         models.Provider     `json:"provider" validate:"required,provider"`
	Id               string              `json:"id" validate:"required"`
	BaseDir          string              `json:"dir" validate:"required"`
	TempTitle        string              `json:"title" validate:"required"`
	DownloadMetadata map[string][]string `json:"downloadMetadata,omitempty"`
}

// IncludesMetadataSlice returns true if the request includes metadata for all passed keys, false otherwise
func (r DownloadRequest) IncludesMetadataSlice(keys []string) bool {
	return r.IncludesMetadata(keys...)
}

// IncludesMetadata returns true if the request includes metadata for all passed keys, false otherwise
func (r DownloadRequest) IncludesMetadata(keys ...string) bool {
	for _, key := range keys {
		if _, ok := r.DownloadMetadata[key]; !ok {
			return false
		}
	}
	return true
}

// GetStrings returns the metadata associated with the key as a slice of strings
// An empty slice will return false
func (r DownloadRequest) GetStrings(key string) ([]string, bool) {
	values, ok := r.DownloadMetadata[key]
	return values, ok && len(values) > 0
}

// GetString returns the metadata associated with the key as a string,
// an empty string is returned if not present
func (r DownloadRequest) GetString(key string) (string, bool) {
	values, ok := r.GetStrings(key)
	if !ok {
		return "", false
	}

	return values[0], true
}

// GetInt returns the metadata associated with the key as an int,
// zero is returned if the value is not present or if conversion failed
func (r DownloadRequest) GetInt(key string) (int, error) {
	val, ok := r.GetString(key)
	if !ok {
		return 0, nil
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}
	return i, nil
}

// GetBool returns the metadata associated with the key as a bool,
// returns true if the value is equal to "true" while ignoring case
func (r DownloadRequest) GetBool(key string) bool {
	val, ok := r.GetString(key)
	if !ok {
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
	Dir       string `json:"dir" validate:"required"`
	ShowFiles bool   `json:"files" validate:"required"`
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
