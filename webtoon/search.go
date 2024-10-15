package webtoon

import (
	"encoding/json"
	"fmt"
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/Fesaa/Media-Provider/wisewolf"
	"io"
	"net/url"
	"regexp"
	"strings"
)

const (
	DOMAIN       = "https://www.webtoons.com"
	BASE_URL     = "https://www.webtoons.com/en/"
	SEARCH_URL   = BASE_URL + "search/immediate?keyword=%s"
	IMAGE_PREFIX = "https://webtoon-phinf.pstatic.net/"
	EPISODE_LIST = DOMAIN + "/episodeList?titleNo=%s"
)

var (
	rg = regexp.MustCompile("[^a-zA-Z0-9 ]+")
)

func Search(options SearchOptions) ([]SearchData, error) {
	resp, err := wisewolf.Client.Get(searchUrl(options.Query))
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response Response
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}

	return utils.Map(response.Result.SearchedList, func(s SearchData) SearchData {
		s.Genre = strings.ToLower(s.Genre)
		return s
	}), nil
}

func constructProxyImg(imageUrl string) string {
	if strings.HasPrefix(imageUrl, IMAGE_PREFIX) {
		imageUrl = strings.TrimPrefix(imageUrl, IMAGE_PREFIX)
	}
	parts := strings.Split(imageUrl, "/")
	if len(parts) != 4 {
		return ""
	}
	date := parts[1]
	id := parts[2]
	fileName := func() string {
		s := parts[3]
		if strings.HasSuffix(s, "?type=q90") {
			return strings.TrimSuffix(s, "?type=q90")
		}
		return s
	}()

	return fmt.Sprintf("proxy/webtoon/covers/%s/%s/%s", date, id, fileName)
}

func searchUrl(keyword string) string {
	keyword = rg.ReplaceAllString(keyword, "")
	return fmt.Sprintf(SEARCH_URL, url.QueryEscape(keyword))
}

func (s *SearchData) Url() string {
	return fmt.Sprintf(BASE_URL+"%s/%s/list?title_no=%d", s.Genre, url.PathEscape(s.Name), s.Id)
}

func (s *SearchData) ProxiedImage() string {
	return constructProxyImg(s.ThumbnailMobile)
}

func (s *SearchData) ComicInfoRating() comicinfo.AgeRating {
	if s.Rating {
		return comicinfo.AgeRatingMaturePlus17
	}
	return comicinfo.AgeRatingEveryone
}
