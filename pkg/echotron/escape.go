package echotron

import "strings"

var markdownEscaper = strings.NewReplacer(
	`_`, `\\_`,
	`*`, `\\*`,
	`[`, `\\[`,
	`]`, `\\]`,
	`(`, `\(`,
	`)`, `\)`,
	`~`, `\\~`,
	"`", string([]byte{0x5C, 0x5C, 0x60}),
	`>`, `\\>`,
	`#`, `\\#`,
	`+`, `\\+`,
	`-`, `\\-`,
	`=`, `\\=`,
	`|`, `\\|`,
	`{`, `\\{`,
	`}`, `\\}`,
	`.`, `\\.`,
	`!`, `\\!`,
)

func EscapeMarkdownMessage(text string) string {
	return markdownEscaper.Replace(text)
}

var htmlEscaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
)

func EscapeHTMLMessage(text string) string {
	return htmlEscaper.Replace(text)
}
