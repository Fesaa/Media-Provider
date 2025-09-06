package services

import (
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
	"github.com/rs/zerolog"
)

type MarkdownService interface {
	MdToSafeHtml(string) string
	SanitizeHtml(string) string
}

type markdownService struct {
	renderer *html.Renderer
	log      zerolog.Logger
}

func MarkdownServiceProvider(log zerolog.Logger) MarkdownService {
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	return &markdownService{
		renderer: renderer,
		log:      log.With().Str("handler", "markdown-service").Logger(),
	}
}

func (m *markdownService) MdToSafeHtml(mdString string) string {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)

	mdBytes := []byte(mdString)

	unsafeHtml := string(markdown.ToHTML(mdBytes, p, m.renderer))
	unsafeHtml = strings.ReplaceAll(unsafeHtml, "<p>", "")
	unsafeHtml = strings.ReplaceAll(unsafeHtml, "</p>", "")
	return strings.TrimSuffix(m.SanitizeHtml(unsafeHtml), "\n")
}

func (m *markdownService) SanitizeHtml(htmlString string) string {
	return bluemonday.UGCPolicy().
		AllowAttrs("target").OnElements("a").
		Sanitize(htmlString)
}
