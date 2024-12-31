package utils

import (
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
	"strings"
)

var (
	renderer *html.Renderer
)

func init() {
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer = html.NewRenderer(opts)
}

func MdToSafeHtml(mdString string) string {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)

	mdBytes := []byte(mdString)

	// TODO: Have it not add <p> tags. These cause it to be a little too big for Kavita most of the time.
	// Instead of replacing it ourselves
	unsafeHtml := string(markdown.ToHTML(mdBytes, p, renderer))

	unsafeHtml = strings.ReplaceAll(unsafeHtml, "<p>", "")
	unsafeHtml = strings.ReplaceAll(unsafeHtml, "</p>", "")
	return SanitizeHtml(unsafeHtml)
}

func SanitizeHtml(htmlString string) string {
	return bluemonday.UGCPolicy().Sanitize(htmlString)
}
