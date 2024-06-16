package mangadex

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

func Init(c MangadexConfig) {
	err := loadTags()
	if err != nil {
		slog.Warn("Failed to load tags, tag filtering won't work", "err", err)
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

	defer resp.Body.Close()
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
