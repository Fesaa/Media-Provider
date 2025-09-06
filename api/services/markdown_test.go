package services

import (
	"testing"

	"github.com/rs/zerolog"
)

// Note that we are only testing the configuration we've done on top of gomarkdown, we are assuming the lib works

func TestMdToSafeHtmlMarkdownWithLinks(t *testing.T) {
	ms := MarkdownServiceProvider(zerolog.Nop())
	input := `[Media-Provider](https://github.com/Fesaa/Media-Provider/)`
	expected := `<a href="https://github.com/Fesaa/Media-Provider/" target="_blank" rel="nofollow noopener">Media-Provider</a>`
	output := ms.MdToSafeHtml(input)

	if output != expected {
		t.Errorf("Expected: %q, got: %q", expected, output)
	}
}

func TestMdToSafeHtmlNoParagraphTags(t *testing.T) {
	ms := MarkdownServiceProvider(zerolog.Nop())
	input := "This is a paragraph.\n\nAnother paragraph."
	expected := "This is a paragraph.\n\nAnother paragraph."
	output := ms.MdToSafeHtml(input)

	if output != expected {
		t.Errorf("Expected: %q, got: %q", expected, output)
	}
}
