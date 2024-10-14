package webtoon

type SearchOptions struct {
	Query string
}

type Data struct {
	Id       string
	Name     string
	Author   string
	Genre    string
	ImageUrl string
	Url      string
}
