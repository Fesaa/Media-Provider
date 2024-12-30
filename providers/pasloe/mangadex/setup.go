package mangadex

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog"
	"io"
	"net/http"
)

func Init(httpClient *http.Client, log zerolog.Logger) {
	err := loadTags(httpClient)
	if err != nil {
		log.Warn().Err(err).Msg("failed to load tags, filtering won't work")
	}
}

func loadTags(httpClient *http.Client) error {
	tagUrl := URL + "/manga/tag"

	resp, err := httpClient.Get(tagUrl)
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
