package adapters

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type MarkdownHTMLConverter struct{}

func NewMarkdownHTMLConverter() *MarkdownHTMLConverter {
	return &MarkdownHTMLConverter{}
}

func (c *MarkdownHTMLConverter) ConvertToHTML(ctx context.Context, input string) (string, error) {
	// create markdown parser with extensions
	extensions := parser.HardLineBreak | parser.NoEmptyLineBeforeBlock | parser.NoIntraEmphasis |
		parser.FencedCode | parser.Strikethrough | parser.SpaceHeadings | parser.BackslashLineBreak
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(input))

	r := newTelegramHTMLRenderer()
	var buf bytes.Buffer
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		return r.RenderNode(&buf, node, entering)
	})
	return buf.String(), nil
}

func newTelegramHTMLRenderer() *telegramHTMLRenderer {
	return &telegramHTMLRenderer{}
}

type telegramHTMLRenderer struct{}

func (r *telegramHTMLRenderer) RenderNode(w io.Writer, node ast.Node, entering bool) ast.WalkStatus {
	switch n := node.(type) {
	case *ast.Paragraph:
		if entering {
			r.writeCR(w)
		}
	case *ast.Text:
		html.EscapeHTML(w, n.Literal)
	case *ast.Strong, *ast.Heading:
		r.outOneOf(w, entering, "<b>", "</b> ")
	case *ast.Emph:
		r.outOneOf(w, entering, "<i>", "</i> ")
	case *ast.Del:
		r.outOneOf(w, entering, "<s>", "</s> ")
	case *ast.Code:
		io.WriteString(w, "<code>")
		html.EscapeHTML(w, n.Literal)
		io.WriteString(w, "</code> ")
	case *ast.Link:
		if entering {
			io.WriteString(w, "<a href=\"")
			html.EscLink(w, n.Destination)
			io.WriteString(w, "\">")
		} else {
			io.WriteString(w, "</a> ")
		}
	case *ast.BlockQuote:
		r.writeCR(w)
		r.outOneOf(w, entering, "<blockquote>", "</blockquote> ")
	case *ast.HorizontalRule:
		if entering {
			r.writeCR(w)
			io.WriteString(w, "------")
			r.writeCR(w)
		}
	case *ast.CodeBlock:
		r.writeCR(w)
		r.writeCodeBlock(w, n)
		r.writeCR(w)
	case *ast.List:
		// No-op on entering and exiting
	case *ast.ListItem:
		r.writeListItem(w, n, entering)
	case *ast.HTMLSpan:
		r.outOneOf(w, entering, "<span>", "</span> ")
	case *ast.Document:
		// No-op on entering and exiting
	case *ast.HTMLBlock:
		if entering {
			io.WriteString(w, "<code>")
			html.EscapeHTML(w, n.Literal)
		} else {
			io.WriteString(w, "</code> ")
		}
	default:
		return ast.SkipChildren
	}

	return ast.GoToNext
}

// outOneOf writes first or second depending on outFirst
func (r *telegramHTMLRenderer) outOneOf(w io.Writer, outFirst bool, first, second string) {
	if outFirst {
		io.WriteString(w, first)
	} else {
		io.WriteString(w, second)
	}
}

// writeListItem writes ast.ListItem node
func (r *telegramHTMLRenderer) writeListItem(w io.Writer, listItem *ast.ListItem, entering bool) {
	if entering {
		r.writeCR(w)
		r.listItemEnter(w, listItem)
	}
}

func (r *telegramHTMLRenderer) listItemEnter(w io.Writer, listItem *ast.ListItem) {
	tab := " - "
	list, ok := listItem.GetParent().(*ast.List)
	if !ok {
		return
	}
	if list.Start > 0 {
		for i, child := range list.GetChildren() {
			if child == listItem {
				tab = fmt.Sprintf("%d. ", list.Start+i)
				break
			}
		}
	}
	if isNestedList(list) {
		tab = "  " + tab
	}
	io.WriteString(w, tab)
}

func isNestedList(list *ast.List) bool {
	if parent := list.GetParent(); parent != nil {
		if _, ok := parent.(*ast.List); ok {
			return true
		}
	}
	return false
}

func (r *telegramHTMLRenderer) writeCodeBlock(w io.Writer, n *ast.CodeBlock) {
	if len(n.Info) > 0 {
		fmt.Fprintf(w, "<pre><code class=\"language-%s\">", bytes.TrimSpace(n.Info))
	} else {
		io.WriteString(w, "<pre><code>")
	}
	html.EscapeHTML(w, n.Literal)
	io.WriteString(w, "</code></pre>")
}

func (r *telegramHTMLRenderer) writeCR(w io.Writer) {
	io.WriteString(w, "\n")
}
