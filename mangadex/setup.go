package mangadex

import (
	"encoding/json"
	"io"
	"net/http"
)

func Init() error {
	err := loadTags()
	if err != nil {
		return err
	}

	return nil
}

func loadTags() error {
	tagUrl := URL + "/manga/tag"

	resp, err := http.Get(tagUrl)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var tagResponse TagResponse
	err = json.Unmarshal(body, &tagResponse)
	if err != nil {
		return err
	}

	for _, tag := range tagResponse.Data {
		enName, ok := tag.Attributes.Name["en"]
		if !ok {
			continue
		}
		tags.Set(enName, tag.Id)
	}
	return nil
}
