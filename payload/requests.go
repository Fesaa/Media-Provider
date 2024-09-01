package payload

import "github.com/Fesaa/Media-Provider/config"

type SearchRequest struct {
	Provider  []config.Provider   `json:"provider"`
	Query     string              `json:"query"`
	Modifiers map[string][]string `json:"modifiers,omitempty"`
}

type DownloadRequest struct {
	Provider  config.Provider `json:"provider"`
	Id        string          `json:"id"`
	BaseDir   string          `json:"dir"`
	TempTitle string          `json:"title"`
}

type StopRequest struct {
	Provider    config.Provider `json:"provider"`
	Id          string          `json:"id"`
	DeleteFiles bool            `json:"delete"`
}

type ListDirsRequest struct {
	Dir       string `json:"dir"`
	ShowFiles bool   `json:"files"`
}

type LoginRequest struct {
	Password string `json:"password"`
	Remember bool   `json:"remember,omitempty"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type UpdatePasswordRequest struct {
	Password string `json:"password"`
}
