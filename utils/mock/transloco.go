package mock

type Transloco struct{}

func (t Transloco) GetTranslation(key string, params ...any) string {
	return key
}

func (t Transloco) GetTranslationLang(lang, key string, params ...any) string {
	return key
}

func (t Transloco) TryTranslationLang(lang, key string, params ...any) (string, error) {
	return key, nil
}

func (t Transloco) GetLanguages() []string {
	return []string{}
}
