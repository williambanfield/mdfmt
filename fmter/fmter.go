package fmter

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

var Decorators = map[ast.NodeKind]Decorator{
	ast.KindEmphasis: {
		Pre:  []byte("**"),
		Post: []byte("**"),
	},
	ast.KindCodeSpan: {
		Pre:  []byte{'`'},
		Post: []byte{'`'},
	},
	ast.KindBlockquote: {
		Pre: []byte("> "),
	},
}

type Decorator struct {
	Pre, Post []byte
}

// CodeFormatter defines an interfaces for providing custom formatting for
// fenced code blocks.
type CodeFormatter interface {
	// Format receives the code as a list of bytes and is expected to return
	// a formatted equivalent of the same code.
	Format([]byte) ([]byte, error)
}

// Option interface for defining functional options for the renderer.
type Option interface {
	SetRendererOption(*Renderer)
}

type Renderer struct {
	codeFormatters map[string]CodeFormatter
	maxWidth       int
}

// NewRenderer returns a goldmark node renderer that pretty-prints the
// input markdown, applying basic style fixes to produce a consist format
// throughout the file.
func NewRenderer(opts ...Option) renderer.NodeRenderer {
	r := &Renderer{
		codeFormatters: make(map[string]CodeFormatter),
		maxWidth:       80,
	}
	for _, opt := range opts {
		opt.SetRendererOption(r)
	}
	return r
}

type withCodeFencerFormatter struct {
	l     string
	fmter CodeFormatter
}

func (w withCodeFencerFormatter) SetRendererOption(r *Renderer) {
	r.codeFormatters[w.l] = w.fmter
}

// WithCodeFenceFormatter is a functional option for specifying a code fence
// formatter. The supplied fmter will be used in code fences that are detected
// to be written in the supplied language.
func WithCodeFenceFormatter(language string, fmter CodeFormatter) Option {
	return withCodeFencerFormatter{
		l:     language,
		fmter: fmter,
	}
}

type withMaxCharacterWidth int

func (w withMaxCharacterWidth) SetRendererOption(r *Renderer) {
	r.maxWidth = int(w)
}

// WithMaxCharacterWidth is a functional option for specifying the max character width
// to limit each line within a paragraph to.
func WithMaxCharacterWidth(w int) Option {
	return withMaxCharacterWidth(w)
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

	// table

	reg.Register(extast.KindTable, r.renderTable)
	reg.Register(extast.KindTableHeader, r.renderTableHeader)
	reg.Register(extast.KindTableRow, r.renderTableRow)
	reg.Register(extast.KindTableCell, r.renderTableCell)
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
		// TODO(williambanfield): this may not be a top-level paragraph but should
		// still respect the max width of the document. Ensure to encorporate any
		// offset that is already present in the line(s) in the maxWidth calculation.
		split := maxWidth(lines, r.maxWidth)
		for i := 0; i < len(split); i++ {
			w.Write(split[i])
			w.WriteByte('\n')
		}
	} else {
		// if this is a top-level paragraph, let handle the newline here, otherwise
		// let the parent node insert the newlines.
		if n.Parent().Kind() == ast.KindDocument {
			w.WriteByte('\n')
		}
	}
	return ast.WalkSkipChildren, nil
}

// TODO: Write a test
// TODO: change from byte length to character length! https://pkg.go.dev/unicode/utf8#RuneCount

// maxWidth takes in a paragraph of text with line breaks and converts it to a
// paragraph where every line contains at least one word and is at most w characters wide,
// granted the first word is not greater than w characters.
func maxWidth(s []byte, w int) [][]byte {
	var res [][]byte
	if len(s) == 0 {
		return res
	}
	sr := bytes.ReplaceAll(s, []byte{'\n'}, []byte{' '})
	inds := spaceRegexp.FindAllIndex(sr, -1)

	// Append an additional position at the end of the list so that the last word
	// in the text will be included. There is not necessarily a space at the end of
	// the text, so the regular expression we used may only find the space before the last word.
	inds = append([][]int{{0}}, inds...)

	// Prepend the first position in the list so that that the first word can be selected.
	// This is necessary because the first space character occurs after the first 'word'.
	// The loop below starts at the first position in the list of indices so without prepending
	// the list []int{{0}}, the first word will be omitted.
	inds = append(inds, []int{len(sr)})

	lineStart := 0
	for lineStart < len(inds)-1 { // loop over lines
		lineEnd := lineStart + 1

		// A single additional offset is added to the lineStart index in the calculation to account
		// for the leading whitespace. This whitespace is going to be deleted, but
		// was the index given by the regular expression, so it should be ignored
		// for character count calculations.
		spOff := 1
		if sr[inds[lineStart][0]] != ' ' {
			spOff = 0
		}

		// loop over words, continually trying to add the next one. If adding the
		// word overflows maxWidth, put the line break after the current word and
		// start the next line.
		for lineEnd < len(inds)-1 &&
			inds[lineEnd+1][0]-(inds[lineStart][0]+spOff) < w {
			lineEnd++
		}
		//TODO(williambanfield): preserve hard line breaks.
		line := bytes.Trim(sr[inds[lineStart][0]:inds[lineEnd][0]], " ")
		res = append(res, line)
		lineStart = lineEnd
	}
	return res
}

func (r *Renderer) renderDocument(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTextBlock(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		lines := []byte{}
		l := n.Lines().Len()
		for i := 0; i < l; i++ {
			l := n.Lines().At(i)
			lines = append(lines, l.Value(s)...)
		}
		// TODO(williambanfield): this may not be a top-level textblock but should
		// still respect the max width of the document. Ensure to encorporate any
		// offset that is already present in the line(s) in the maxWidth calculation.
		split := maxWidth(lines, r.maxWidth)
		for i := 0; i < len(split); i++ {
			w.Write(split[i])
			w.WriteByte('\n')
		}
	} else {
		// if this is a top-level paragraph, let handle the newline here, otherwise
		// let the parent node insert the newlines.
		if n.Parent().Kind() == ast.KindDocument {
			w.WriteByte('\n')
		}
	}
	return ast.WalkSkipChildren, nil
}
func (r *Renderer) renderHeading(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		h := n.(*ast.Heading)
		//TODO(williambanfield): preserve header style. Headers can be prefixed with # or underlined.
		// See: https://www.markdownguide.org/basic-syntax/#alternate-syntax
		w.Write(bytes.Repeat([]byte{'#'}, h.Level))
		w.WriteByte(' ')
	} else {
		w.WriteByte('\n')
		w.WriteByte('\n')
	}
	return ast.WalkContinue, nil
}
func (r *Renderer) renderBlockquote(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.Write(Decorators[n.Kind()].Pre)
	} else {
		w.Write(Decorators[n.Kind()].Post)
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
		} else {
			w.Write(lines)
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
	if entering {
		return ast.WalkContinue, nil
	}

	// If a list follows this node and there was no extra newline in the original document,
	// do not add a newline. goldmark considers lists of different indentation
	// level to be separate lists. For example:
	//
	// * Item 1
	//   * Item 2
	// * Item 3
	// * Item 4
	//
	// Is 3 different lists from goldmark's perspective, with the list containing Item 2 being a child list of the list containing Item 1.
	//
	// TODO(williambanfield): Determine if a newline followed the node in the original text.
	if n.NextSibling() != nil && n.NextSibling().Kind() != ast.KindList {
		w.WriteByte('\n')
	}
	return ast.WalkContinue, nil
}
func (r *Renderer) renderListItem(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	li := n.(*ast.ListItem)
	l := li.Parent().(*ast.List)
	if entering {
		w.Write(bytes.Repeat([]byte{' '}, li.Offset-2))
		if l.IsOrdered() {
			// TODO(williambanfield): Preserve list numbering.
			fmt.Fprintf(w, "%d%s ", l.Start, string(l.Marker))
		} else {
			w.WriteByte(l.Marker)
			w.WriteByte(' ')
		}
	}
	// Currently, list items receive a newline by way of containing a 'textblock' so a newline is not needed here
	// to handle the !entering case.
	return ast.WalkContinue, nil
}
func (r *Renderer) renderThematicBreak(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("---") //TODO(williambanfield): preserve original thematic break characters used.
	} else {
		w.WriteByte('\n')
	}
	return ast.WalkContinue, nil
}
func (r *Renderer) renderAutoLink(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}
func (r *Renderer) renderCodeSpan(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.Write(Decorators[n.Kind()].Pre)
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			segment := c.(*ast.Text).Segment
			value := segment.Value(s)
			if bytes.HasSuffix(value, []byte("\n")) {
				w.Write(value[:len(value)-1])
				w.WriteByte(' ')
			} else {
				w.Write(value)
			}
		}
		return ast.WalkSkipChildren, nil
	} else {
		w.Write(Decorators[n.Kind()].Post)
	}
	return ast.WalkContinue, nil
}
func (r *Renderer) renderEmphasis(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.Write(Decorators[n.Kind()].Pre)
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			segment := c.(*ast.Text).Segment
			value := segment.Value(s)
			if bytes.HasSuffix(value, []byte("\n")) {
				w.Write(value[:len(value)-1])
				w.WriteByte(' ')
			} else {
				w.Write(value)
			}
		}
		return ast.WalkSkipChildren, nil
	} else {
		w.Write(Decorators[n.Kind()].Post)
	}
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
	if !entering {
		return ast.WalkContinue, nil
	}
	t := n.(*ast.Text)
	w.Write(n.Text(s))
	if t.HardLineBreak() || t.SoftLineBreak() {
		_ = w.WriteByte('\n')
	}
	return ast.WalkContinue, nil
}
func (r *Renderer) renderString(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTable(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	// Calculate the max-width column for every column in the table.
	if !entering {
		w.WriteByte('\n')
		return ast.WalkContinue, nil
	}
	gcl := make([]int, n.FirstChild().ChildCount())
	for r := n.FirstChild(); r != nil; r = r.NextSibling() {
		cx := 0
		for c := r.FirstChild(); c != nil; c = c.NextSibling() {
			var l int
			lines := c.Lines()
			for i := 0; i < lines.Len(); i++ {
				li := lines.At(i)
				l += li.Len()
			}
			if l > gcl[cx] {
				gcl[cx] = l
			}
			cx++
		}
	}
	for r := n.FirstChild(); r != nil; r = r.NextSibling() {
		r.SetAttributeString("column-widths", gcl)
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTableHeader(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	a, ok := n.AttributeString("column-widths")
	if !ok {
		panic("missing attribute for column width")
	}
	gcl := a.([]int)
	if entering {
		cx := 0
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			c.SetAttributeString("column-width", gcl[cx])
			cx++
		}
	} else {
		w.WriteByte('|')
		w.WriteByte('\n')
		cx := 0
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			w.WriteByte('|')
			w.Write(bytes.Repeat([]byte{'-'}, gcl[cx]+2))
			cx++
		}
		w.WriteByte('|')
		w.WriteByte('\n')
	}
	return ast.WalkContinue, nil
}
func (r *Renderer) renderTableRow(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	a, ok := n.AttributeString("column-widths")
	if !ok {
		panic("missing attribute for column width")
	}
	gcl := a.([]int)
	if entering {
		cx := 0
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			c.SetAttributeString("column-width", gcl[cx])
			cx++
		}
	} else {
		w.WriteByte('|')
		w.WriteByte('\n')
	}
	return ast.WalkContinue, nil
}
func (r *Renderer) renderTableCell(w util.BufWriter, s []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteByte('|') //TODO(williambanfield): respect the originally used table column separator.
		w.WriteByte(' ')
	} else {
		a, ok := n.AttributeString("column-width")
		if !ok {
			panic("missing attribute for column width")
		}
		wd := a.(int)
		var tl int
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			li := lines.At(i)
			tl += li.Len()
		}
		pl := wd - tl
		w.Write(bytes.Repeat([]byte{' '}, pl+1))
	}
	return ast.WalkContinue, nil
}
