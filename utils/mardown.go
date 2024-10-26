package utils

import (
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
)

var (
	p        *parser.Parser
	renderer *html.Renderer
)

func init() {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p = parser.NewWithExtensions(extensions)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer = html.NewRenderer(opts)
}

func MdToSafeHtml(mdString string) string {
	mdBytes := []byte(mdString)
	unsafeHtml := markdown.ToHTML(mdBytes, p, renderer)
	safeHtml := bluemonday.UGCPolicy().SanitizeBytes(unsafeHtml)
	return string(safeHtml)
}

func SanitizeHtml(htmlString string) string {
	return bluemonday.UGCPolicy().Sanitize(htmlString)
}
