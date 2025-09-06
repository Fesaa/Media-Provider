package utils

import (
	"net/url"
	"path"
)

func Ext(uri string, defaultExt ...string) string {
	def := OrDefault(defaultExt, "jpg")

	parsedURL := MustReturn(url.Parse(uri))
	parsedURL.RawQuery = ""
	parsedURL.Fragment = ""

	ext := path.Ext(parsedURL.String())
	if ext == "" {
		return def
	}

	return ext
}
