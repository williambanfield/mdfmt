package mdfmt

import (
	"bytes"
	"regexp"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type Renderer struct {
	codeFormatters map[string]CodeFormatter
}

// CodeFormatter defines an interfaces for providing custom formatting for
// fenced code blocks.
type CodeFormatter interface {
	// Format receives the code as a list of bytes and is expected to return
	// a formatted equivalent of the same code.
	Format([]byte) ([]byte, error)

	// Languages returns the list of languages this formatter should be used to
	// format. If the list is empty, it will not be used to format any languages.
	Languages() []string
}

func NewRenderer(cfs []CodeFormatter) renderer.NodeRenderer {
	cm := make(map[string]CodeFormatter)
	for _, c := range cfs {
		for _, l := range c.Languages() {
			cm[l] = c
		}
	}
	return &Renderer{
		codeFormatters: cm,
	}
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

var spaceRegexp = regexp.MustCompile(` `)

func (r *Renderer) renderParagraph(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		lines := []byte{}

		l := n.Lines().Len()
		for i := 0; i < l; i++ {
			l := n.Lines().At(i)
			lines = append(lines, l.Value(s)...)
		}
		split := maxWidth(lines, 80)
		for i := 0; i < len(split); i++ {
			w.Write(split[i])
			w.WriteByte('\n')
		}
		w.WriteByte('\n')
	}
	return ast.WalkSkipChildren, nil
}

// TODO: Write a test
// TODO: change from byte length to character length! https://pkg.go.dev/unicode/utf8#RuneCount

// maxWidth takes in a paragraph of text with line breaks and converts it to a
// paragraph where every line contains at least one word and is at most w characters wide,
// granted the first word is not greater than w characters.
func maxWidth(s []byte, w int) [][]byte {
	inds := spaceRegexp.FindAllIndex(s, -1)

	// Prepend the first position in the list so that that the first word can be selected.
	// This is necessary because the first space character occurs after the first 'word'.
	// The loop below starts at the first position in the list of indices so without prepending
	// the list []int{{0}}, the first word will be omitted.
	inds = append([][]int{{0}}, inds...)

	var res [][]byte
	lineStart := 0
	for lineStart < len(inds)-1 { // loop over lines
		lineEnd := lineStart + 1
		for lineEnd < len(inds)-1 &&
			inds[lineEnd+1][0]-inds[lineStart][0] < w { // loop over words, continually trying to add the next one!
			lineEnd++
		}

		//TODO(williambanfield): preserve hard line breaks.
		line := bytes.Trim(s[inds[lineStart][0]:inds[lineEnd][0]], " ")
		res = append(res, line)
		lineStart = lineEnd
	}
	return res
}

func (r *Renderer) renderDocument(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTextBlock(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}
func (r *Renderer) renderHeading(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		h := n.(*ast.Heading)
		w.Write(bytes.Repeat([]byte{'#'}, h.Level))
		w.WriteByte(' ')
	} else {
		w.WriteByte('\n')
	}
	return ast.WalkContinue, nil
}
func (r *Renderer) renderBlockquote(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("> ")
	} else {
		w.WriteByte('\n')
	}
	return ast.WalkContinue, nil
}
func (r *Renderer) renderCodeBlock(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}
func (r *Renderer) renderFencedCodeBlock(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	fcb := n.(*ast.FencedCodeBlock)
	if entering {
		_, _ = w.WriteString("```")
		ln := fcb.Language(s)
		if ln != nil {
			_, _ = w.Write(ln)
		}
		w.WriteByte('\n')
		l := n.Lines().Len()
		var lines []byte
		for i := 0; i < l; i++ {
			line := n.Lines().At(i)
			lines = append(lines, line.Value(s)...)
		}
		if fmter, ok := r.codeFormatters[string(ln)]; ok {
			res, err := fmter.Format(lines)
			if err != nil {
				return ast.WalkStop, err
			}
			w.Write(res)
		}
	} else {
		_, _ = w.WriteString("```\n")
	}
	return ast.WalkContinue, nil
}
func (r *Renderer) renderHTMLBlock(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}
func (r *Renderer) renderList(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}
func (r *Renderer) renderListItem(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}
func (r *Renderer) renderThematicBreak(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}
func (r *Renderer) renderAutoLink(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}
func (r *Renderer) renderCodeSpan(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}
func (r *Renderer) renderEmphasis(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}
func (r *Renderer) renderImage(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}
func (r *Renderer) renderLink(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}
func (r *Renderer) renderRawHTML(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}
func (r *Renderer) renderText(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.Write(n.Text(s))
	}
	return ast.WalkContinue, nil
}
func (r *Renderer) renderString(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}
