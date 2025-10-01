package common

import (
	"testing"
)

func TestCore_IsContent(t *testing.T) {
	type testCase struct {
		name     string
		diskName string
		want     bool
		chapter  string
		volume   string
	}
	tests := []testCase{
		{
			name:     "Valid Chapter Format",
			diskName: "My Manga Ch. 0012.cbz",
			want:     true,
			volume:   "",
			chapter:  "12",
		},
		{
			name:     "Valid Volume Format",
			diskName: "My Manga Vol. 05.cbz",
			want:     true,
			volume:   "5",
		},
		{
			name:     "Valid OneShot Format (new)",
			diskName: "My Manga Oneshot Title (OneShot).cbz",
			want:     true,
		},
		{
			name:     "Valid OneShot Format (old)",
			diskName: "My Manga OneShot Oneshot Title.cbz",
			want:     true,
		},
		{
			name:     "Invalid Format - no match",
			diskName: "Random_File_Name.zip",
			want:     false,
		},
		{
			name:     "Invalid Format - wrong extension",
			diskName: "My Manga Ch. 0012.pdf",
			want:     false,
		},
		{
			name:     "Valid format with Volume",
			diskName: "My Manga Vol. 5 Ch. 0007.cbz",
			want:     true,
			volume:   "5",
			chapter:  "7",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, got := IsCbz(tt.diskName)

			if got != tt.want {
				t.Errorf("IsContent() = %v, want %v", got, tt.want)
			}

			if content.Volume != tt.volume {
				t.Errorf("IsContent() = %v,\n want %v", content.Volume, tt.volume)
			}

			if content.Chapter != tt.chapter {
				t.Errorf("IsContent() = %v,\n want %v", content.Chapter, tt.chapter)
			}
		})
	}
}
