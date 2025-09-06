package services

import (
	"slices"
	"testing"

	"github.com/rs/zerolog"
	"github.com/spf13/afero"
)

func tempTransloco(t *testing.T) TranslocoService {
	t.Helper()
	transloco, err := TranslocoServiceProvider(zerolog.Nop(), afero.Afero{Fs: afero.NewMemMapFs()})
	if err != nil {
		t.Fatalf("error creating transloco %v", err)
	}

	transloco.(*translocoService).languages = map[string]map[string]string{
		"en": {
			"key1":      "value1",
			"key2":      "value2",
			"key3":      "value3",
			"empty-key": "",
			"formatted": "Hi! %s",
		},
		"es": {
			"key1": "value1-in-es",
		},
	}

	return transloco
}

func Test_translocoService_GetLanguages(t1 *testing.T) {
	transloco := tempTransloco(t1)
	languages := transloco.GetLanguages()
	if len(languages) != 2 {
		t1.Errorf("expected 2 languages, got %v", len(languages))
	}

	want := []string{"en", "es"}
	for _, l := range languages {
		if !slices.Contains(want, l) {
			t1.Errorf("expected %v, got %v", want, l)
		}
	}
}

func Test_translocoService_GetTranslation(t1 *testing.T) {
	type args struct {
		key    string
		params []any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal key",
			args: args{
				key: "key1",
			},
			want: "value1",
		},
		{
			name: "empty key",
			args: args{
				key: "empty-key",
			},
			want: "",
		},
		{
			name: "invalid key",
			args: args{
				key: "this-key-doesnt-exist",
			},
			want: "this-key-doesnt-exist",
		},
		{
			name: "formatted key",
			args: args{
				key:    "formatted",
				params: []any{"Bye!"},
			},
			want: "Hi! Bye!",
		},
	}

	t := tempTransloco(t1)

	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			if got := t.GetTranslation(tt.args.key, tt.args.params...); got != tt.want {
				t1.Errorf("GetTranslation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_translocoService_GetTranslationLang(t1 *testing.T) {
	type args struct {
		lang   string
		key    string
		params []any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no-fallback",
			args: args{
				lang: "es",
				key:  "key1",
			},
			want: "value1-in-es",
		},
		{
			name: "fallback",
			args: args{
				lang: "es",
				key:  "key2",
			},
			want: "value2",
		},
	}

	t := tempTransloco(t1)
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			if got := t.GetTranslationLang(tt.args.lang, tt.args.key, tt.args.params...); got != tt.want {
				t1.Errorf("GetTranslationLang() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_translocoService_TryTranslationLang(t1 *testing.T) {
	type args struct {
		lang   string
		key    string
		params []any
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "unknown lang",
			args: args{
				lang: "unknown",
				key:  "key1",
			},
			wantErr: true,
		},
		{
			name: "unknown key",
			args: args{
				lang: "es",
				key:  "unknown-key",
			},
			wantErr: true,
		},
		{
			name: "no-failure",
			args: args{
				lang: "en",
				key:  "key1",
			},
			want:    "value1",
			wantErr: false,
		},
	}

	t := tempTransloco(t1)
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			got, err := t.TryTranslationLang(tt.args.lang, tt.args.key, tt.args.params...)
			if (err != nil) != tt.wantErr {
				t1.Errorf("TryTranslationLang() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t1.Errorf("TryTranslationLang() got = %v, want %v", got, tt.want)
			}
		})
	}
}
