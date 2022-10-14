package mdfmt

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type Renderer struct {
}

func NewRenderer() renderer.NodeRenderer {
	return &Renderer{}
}

// RendererFuncs registers NodeRendererFuncs to given NodeRendererFuncRegisterer.
func (r *Renderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindTextBlock, r.renderTextBlock)
	reg.Register(ast.KindParagraph, r.renderParagraph)
	reg.Register(ast.KindDocument, r.renderDocument)
	reg.Register(ast.KindHeading, r.renderHeading)
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
	reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
	reg.Register(ast.KindHTMLBlock, r.renderHTMLBlock)
	reg.Register(ast.KindList, r.renderList)
	reg.Register(ast.KindListItem, r.renderListItem)
	reg.Register(ast.KindThematicBreak, r.renderThematicBreak)

	// inlines

	reg.Register(ast.KindAutoLink, r.renderAutoLink)
	reg.Register(ast.KindCodeSpan, r.renderCodeSpan)
	reg.Register(ast.KindEmphasis, r.renderEmphasis)
	reg.Register(ast.KindImage, r.renderImage)
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindRawHTML, r.renderRawHTML)
	reg.Register(ast.KindText, r.renderText)
	reg.Register(ast.KindString, r.renderString)
}

var wordBoundaryRegexp = regexp.MustCompile("\b")

func (r *Renderer) renderParagraph(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	fmt.Println("Paragraph")
	if entering {
		lines := []byte{}
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			segment := c.(*ast.Text).Segment
			value := segment.Value(s)
			lines = append(lines, value...)
		}
		for _, l := range maxWidth(lines, 80) {
			fmt.Println(string(l))
		}
	}
	return ast.WalkSkipChildren, nil
}

func maxWidth(s []byte, w int) [][]byte {
	n := bytes.ReplaceAll(s, []byte("\n"), []byte(" "))
	var res [][]byte
	for i := 0; i < len(s); {
		li := i
		ri := i + w
		if ri > len(s) {
			ri = len(s)
			res = append(res, n[li:ri])
			break
		}
		j := ri
		for ; j > li && n[j-1] != byte(' '); j-- {
		}
		if j > li {
			ri = j
		}
		line := n[li:ri]
		res = append(res, line)
		i += len(line)
	}
	return res
}

func (r *Renderer) renderDocument(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	fmt.Println("Document")
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTextBlock(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	fmt.Println("TextBlock")
	return ast.WalkSkipChildren, nil
}
func (r *Renderer) renderHeading(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	fmt.Println("Heading")
	return ast.WalkContinue, nil
}
func (r *Renderer) renderBlockquote(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	fmt.Println("Blockquote")
	return ast.WalkContinue, nil
}
func (r *Renderer) renderCodeBlock(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	fmt.Println("CodeBlock")
	return ast.WalkContinue, nil
}
func (r *Renderer) renderFencedCodeBlock(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	fmt.Println("FencedCodeBlock")
	return ast.WalkContinue, nil
}
func (r *Renderer) renderHTMLBlock(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	fmt.Println("HTMLBlock")
	return ast.WalkContinue, nil
}
func (r *Renderer) renderList(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	fmt.Println("List")
	return ast.WalkContinue, nil
}
func (r *Renderer) renderListItem(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	fmt.Println("ListItem")
	return ast.WalkContinue, nil
}
func (r *Renderer) renderThematicBreak(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	fmt.Println("ThematicBreak")
	return ast.WalkContinue, nil
}
func (r *Renderer) renderAutoLink(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	fmt.Println("AutoLink")
	return ast.WalkContinue, nil
}
func (r *Renderer) renderCodeSpan(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	fmt.Println("CodeSpan")
	return ast.WalkContinue, nil
}
func (r *Renderer) renderEmphasis(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	fmt.Println("Emphasis")
	return ast.WalkContinue, nil
}
func (r *Renderer) renderImage(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	fmt.Println("Image")
	return ast.WalkContinue, nil
}
func (r *Renderer) renderLink(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	fmt.Println("Link")
	return ast.WalkContinue, nil
}
func (r *Renderer) renderRawHTML(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	fmt.Println("RawHTML")
	return ast.WalkContinue, nil
}
func (r *Renderer) renderText(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	fmt.Println("Text")
	return ast.WalkContinue, nil
}
func (r *Renderer) renderString(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	fmt.Println("String")
	return ast.WalkContinue, nil
}
