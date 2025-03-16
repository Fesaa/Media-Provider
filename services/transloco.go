package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"os"
	"path/filepath"
	"strings"
)

var (
	DefaultLanguage = config.OrDefault(os.Getenv("LANGUAGE"), "en")

	ErrLanguageNotFound = errors.New("language not found")
	ErrKeyNotFound      = errors.New("key not found")
)

type TranslocoService interface {
	// GetTranslation calls GetTranslationLang with lang = DefaultLanguage
	GetTranslation(key string, params ...any) string
	// GetTranslationLang returns the formatted string if found, otherwise the key
	GetTranslationLang(lang, key string, params ...any) string
	// TryTranslationLang returns the formatted string if found, otherwise an appropriate error
	TryTranslationLang(lang, key string, params ...any) (string, error)

	// GetLanguages returns a list of all loaded languages
	GetLanguages() []string
}

type translocoService struct {
	languages map[string]map[string]string
	log       zerolog.Logger
	fs        afero.Afero
}

func TranslocoServiceProvider(log zerolog.Logger, fs afero.Afero) (TranslocoService, error) {
	transloco := &translocoService{
		languages: make(map[string]map[string]string),
		log:       log.With().Str("handler", "transloco-service").Logger(),
		fs:        fs,
	}
	if err := transloco.loadLanguages(); err != nil {
		return nil, err
	}
	return transloco, nil
}

func (t *translocoService) loadLanguages() error {
	files, err := filepath.Glob("./I18N/*.json")
	if err != nil {
		return err
	}

	t.log.Trace().Int("files", len(files)).Msg("loading languages")
	for _, file := range files {
		language := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))

		fileContent, err := t.fs.ReadFile(file)
		if err != nil {
			return err
		}

		var mappings map[string]string
		if err := json.Unmarshal(fileContent, &mappings); err != nil {
			return err
		}

		t.languages[language] = mappings
		t.log.Debug().Str("language", language).Int("keys", len(mappings)).Msg("loaded language")
	}
	return nil
}

func (t *translocoService) GetLanguages() []string {
	return utils.Keys(t.languages)
}

func (t *translocoService) GetTranslation(key string, params ...any) string {
	return t.GetTranslationLang(DefaultLanguage, key, params...)
}

func (t *translocoService) GetTranslationLang(lang, key string, params ...any) string {
	tl, err := t.TryTranslationLang(lang, key, params...)
	if err != nil {
		return key
	}
	return tl
}

func (t *translocoService) TryTranslationLang(lang, key string, params ...any) (string, error) {
	mappings, ok := t.languages[lang]
	if !ok {
		t.log.Warn().Str("lang", lang).Msg("language not found. Unexpected? Report this!")
		return "", ErrLanguageNotFound
	}

	mapping, ok := mappings[key]
	if !ok {
		if lang != DefaultLanguage {
			t.log.Trace().Str("lang", lang).Str("key", key).Msg("key not found in language, falling back")
			return t.TryTranslationLang(DefaultLanguage, key, params...)
		}

		t.log.Warn().Str("lang", lang).Str("key", key).Msg("key not found. Unexpected? Report this!")
		return "", ErrKeyNotFound
	}

	if len(params) == 0 {
		return mapping, nil
	}

	return fmt.Sprintf(mapping, params...), nil
}
