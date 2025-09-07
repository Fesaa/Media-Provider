package webtoon

import (
	"testing"

	"github.com/Fesaa/Media-Provider/internal/comicinfo"
)

func TestSearchData_Url(t *testing.T) {
	d := SearchData{
		Id:              WebToonID,
		Name:            WebToonName,
		ReadCount:       "0",
		ThumbnailMobile: "",
		AuthorNameList:  []string{"TIKKLIL", "Rebecca Sullivan"},
		Genre:           "romance",
		Rating:          false,
	}

	want := "https://www.webtoons.com/en/romance/Night%20Owls%20&%20Summer%20Skies/list?title_no=4747"
	if got := d.Url(); got != want {
		t.Errorf("SearchData.Url() = %v, want %v", got, want)
	}
}

func TestSearchData_ProxiedImage(t *testing.T) {
	d := SearchData{
		ThumbnailMobile: "https://webtoon-phinf.pstatic.net/20220920_231/1663613908982aRtbh_JPEG/5NightOwls26SummerSkies_mobile_thumbnail.jpg?type=q90",
	}

	want := "proxy/webtoon/covers/20220920_231/1663613908982aRtbh_JPEG/5NightOwls26SummerSkies_mobile_thumbnail.jpg"
	if got := d.ProxiedImage(); got != want {
		t.Errorf("SearchData.ProxiedImage() = %v, want %v", got, want)
	}
}

func TestSearchData_ProxiedImageInvalid(t *testing.T) {
	d := SearchData{}

	if got := d.ProxiedImage(); got != "" {
		t.Errorf("SearchData.ProxiedImage() = %v, want %v", got, "")
	}
}

func TestSearchData_ComicInfoRating(t *testing.T) {
	type test struct {
		name   string
		rating bool
		want   comicinfo.AgeRating
	}

	tests := []test{
		{
			name:   "true",
			rating: true,
			want:   comicinfo.AgeRatingMaturePlus17,
		},
		{
			name:   "false",
			rating: false,
			want:   comicinfo.AgeRatingEveryone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := SearchData{
				Rating: tt.rating,
			}
			if got := d.ComicInfoRating(); got != tt.want {
				t.Errorf("SearchData.ComicInfoRating() = %v, want %v", got, tt.want)
			}
		})
	}
}
