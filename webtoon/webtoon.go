package webtoon

import "github.com/Fesaa/Media-Provider/payload"

func newWebToon(req payload.DownloadRequest, client Client) WebToon {
	return &webtoon{}
}

type webtoon struct {
}

func (w *webtoon) Title() string {
	//TODO implement me
	panic("implement me")
}

func (w *webtoon) Id() string {
	//TODO implement me
	panic("implement me")
}

func (w *webtoon) GetBaseDir() string {
	//TODO implement me
	panic("implement me")
}

func (w *webtoon) Cancel() {
	//TODO implement me
	panic("implement me")
}

func (w *webtoon) WaitForInfoAndDownload() {
	//TODO implement me
	panic("implement me")
}

func (w *webtoon) GetInfo() payload.InfoStat {
	//TODO implement me
	panic("implement me")
}

func (w *webtoon) GetDownloadDir() string {
	//TODO implement me
	panic("implement me")
}

func (w *webtoon) GetPrevChapters() []string {
	//TODO implement me
	panic("implement me")
}
