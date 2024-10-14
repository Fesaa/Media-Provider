package webtoon

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/url"
	"strings"
)

const (
	BASE_URL     = "https://www.webtoons.com/en/"
	SEARCH_URL   = BASE_URL + "search?keyword=%s"
	IMAGE_PREFIX = "https://webtoon-phinf.pstatic.net/"
)

func Search(options SearchOptions) ([]Data, error) {
	doc, err := wrapInDoc(searchUrl(options.Query))
	if err != nil {
		return nil, err
	}

	webtoons := doc.Find(".card_lst li")
	if webtoons.Length() == 0 {
		return []Data{}, nil
	}

	return goquery.Map(webtoons, constructWebToonFromNode), nil
}

func constructWebToonFromNode(_ int, s *goquery.Selection) Data {
	pageUrl := s.Find("a").AttrOr("href", "")
	img := s.Find("img").First().AttrOr("src", "")
	info := s.Find(".info")
	subj := info.Find(".subj").Text()
	author := info.Find(".author").Text()
	genre := s.Find(".genre").Text()
	d := Data{
		Id:       extractId(pageUrl),
		Name:     subj,
		Author:   author,
		ImageUrl: constructProxyImg(img),
		Genre:    genre,
		Url:      pageUrl,
	}
	return d
}

func constructProxyImg(imageUrl string) string {
	if !strings.HasPrefix(imageUrl, IMAGE_PREFIX) {
		return ""
	}
	parts := strings.Split(strings.TrimPrefix(imageUrl, IMAGE_PREFIX), "/")
	if len(parts) != 3 {
		return ""
	}
	date := parts[0]
	id := parts[1]
	fileName := func() string {
		s := parts[2]
		if strings.HasSuffix(s, "?type=q90") {
			return strings.TrimSuffix(s, "?type=q90")
		}
		return s
	}()

	return fmt.Sprintf("proxy/webtoon/covers/%s/%s/%s", date, id, fileName)
}

func extractId(u string) string {
	if strings.HasPrefix(u, BASE_URL) {
		return strings.TrimPrefix(u, BASE_URL)
	}
	return ""
}

func searchUrl(keyword string) string {
	return fmt.Sprintf(SEARCH_URL, url.QueryEscape(keyword))
}
