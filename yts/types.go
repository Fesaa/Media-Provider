package yts

type SearchResult struct {
	Status         string     `json:"status"`
	Status_message string     `json:"status_message"`
	Data           SearchData `json:"data,omitempty"`
}

type SearchData struct {
	MovieCount int     `json:"movie_count"`
	Limit      int     `json:"limit"`
	Page       int     `json:"page_number"`
	Movies     []Movie `json:"movies"`
}

type Movie struct {
	ID               int       `json:"id"`
	Url              string    `json:"url"`
	Imdb_code        string    `json:"imbd_code"`
	Title            string    `json:"title"`
	TitleEnglish     string    `json:"title_english"`
	TitleLong        string    `json:"title_long"`
	Slug             string    `json:"slug"`
	Year             int       `json:"year"`
	Rating           float32   `json:"rating"`
	Genres           []string  `json:"genres"`
	Summary          string    `json:"summary"`
	DescriptionFull  string    `json:"description_full"`
	Lang             string    `json:"lang"`
	BackGroundImage  string    `json:"background_image"`
	SmallCoverImage  string    `json:"small_cover_image"`
	MediumCoverImage string    `json:"medium_cover_image"`
	LargeCoverImage  string    `json:"large_cover_image"`
	State            string    `json:"state"`
	Torrents         []Torrent `json:"torrents"`
}

type Torrents struct {
	Torrents         []Torrent `json:"torrents"`
	DateUploaded     string    `json:"date_uploaded"`
	DateUploadedUnix int       `json:"date_uploaded_unix"`
}

type Torrent struct {
	Url              string `json:"url"`
	Hash             string `json:"hash"`
	Quality          string `json:"quality"`
	Type             string `json:"type"`
	Seeds            int    `json:"seeds"`
	Peers            int    `json:"peers"`
	Size             string `json:"size"`
	DateUploaded     string `json:"date_uploaded"`
	DateUploadedUnix int    `json:"date_uploaded_unix"`
}
