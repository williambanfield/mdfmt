package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/williambanfield/marker/mdfmt"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

func main() {
	f, err := os.Open("./README.md")
	if err != nil {
		panic(err)
	}
	src, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	buf := bytes.Buffer{}
	r := text.NewReader(src)
	p := newParser()
	n := p.Parse(r)
	re := newRenderer()
	re.Render(os.Stdout, src, n)
	fmt.Println(string(buf.Bytes()))
}

func newParser() parser.Parser {
	p := goldmark.DefaultParser()
	//	p.AddOptions(parser.WithParagraphTransformers(util.Prioritized(extension.NewTableParagraphTransformer(), 200)))
	return p
}

func newRenderer() renderer.Renderer {
	return renderer.NewRenderer(
		renderer.WithNodeRenderers(
			util.Prioritized(mdfmt.NewRenderer(), 1000),
			//util.Prioritized(html.NewRenderer(), 1000),
			//			util.Prioritized(extension.NewTableHTMLRenderer(), 2000),
		),
	)
}
