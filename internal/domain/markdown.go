package domain

import "context"

type MarkdownHTMLConverter interface {
	ConvertToHTML(ctx context.Context, markdown string) (string, error)
}
