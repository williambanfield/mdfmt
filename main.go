package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"

	"github.com/williambanfield/mdfmt/fmter"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

func main() {
	var input io.Reader
	if len(os.Args) == 1 {
		input = os.Stdin
	} else {
		var err error
		input, err = os.Open(os.Args[1])
		if err != nil {
			panic(err)
		}
	}
	src, err := io.ReadAll(input)
	if err != nil {
		panic(err)
	}
	r := text.NewReader(src)
	p := newParser()
	n := p.Parse(r)
	re := newRenderer()
	re.Render(os.Stdout, src, n)
}

func newParser() parser.Parser {
	p := goldmark.DefaultParser()
	p.AddOptions(parser.WithParagraphTransformers(util.Prioritized(extension.NewTableParagraphTransformer(), 200)))
	return p
}

func newRenderer() renderer.Renderer {
	var opts []fmter.Option
	gp, err := exec.LookPath("gofmt")
	if err == nil {
		gofmt := gofmter{
			path: gp,
		}
		opts = append(opts, fmter.WithCodeFenceFormatter("go", gofmt))
	}
	mdf := fmter.NewRenderer(opts...)

	return renderer.NewRenderer(
		renderer.WithNodeRenderers(
			util.Prioritized(mdf, 1000),
		),
	)
}

type gofmter struct {
	path string
}

func (g gofmter) Format(b []byte) ([]byte, error) {
	c := exec.Command(g.path)
	c.Stdin = bytes.NewReader(b)
	buf := &bytes.Buffer{}
	c.Stdout = buf
	err := c.Run()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
