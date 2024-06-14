package config

// Not really the right package, will see later where

type Info struct {
	Provider  Provider `json:"Provider"`
	InfoHash  string   `json:"InfoHash"`
	Name      string   `json:"Name"`
	Size      string   `json:"Size"`
	Progress  int64    `json:"Progress"`
	Completed int64    `json:"Completed"`
	Speed     string   `json:"Speed"`
}
