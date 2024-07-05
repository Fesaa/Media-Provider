package mangadex

import (
	"encoding/json"
	"fmt"
	"github.com/Fesaa/Media-Provider/log"
	"io"
	"net/http"
)

func Init(c Config) {
	err := loadTags()
	if err != nil {
		log.Warn("failed to load tags, filtering won't work", "err", err)
	}

	m = newClient(c)
}

func loadTags() error {
	tagUrl := URL + "/manga/tag"

	resp, err := http.Get(tagUrl)
	if err != nil {
		return fmt.Errorf("loadTags Get: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("loadTags status: %s", resp.Status)
	}

	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			log.Warn("failed to close body", "error", err)
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("loadTags readAll: %s", err)
	}

	var tagResponse TagResponse
	err = json.Unmarshal(body, &tagResponse)
	if err != nil {
		return fmt.Errorf("loadTags unmarshal: %s", err)
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
