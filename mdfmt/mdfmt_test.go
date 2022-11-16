package mdfmt_test

import (
	"bytes"
	"testing"

	"github.com/williambanfield/marker/mdfmt"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

func TestParagraphMaxWidth(t *testing.T) {
	cases := []struct {
		name     string
		maxWidth int
		input    []byte
		expected []byte
	}{
		{
			name:     "width max respected",
			maxWidth: 10,
			input:    []byte("a b c d e f g h i j\n"),
			expected: []byte("a b c d e\nf g h i j\n"),
		},
		{
			name:     "long unbroken strings ignore max width",
			maxWidth: 10,
			input:    []byte("abcdefghijhij a b c d e\n"),
			expected: []byte("abcdefghijhij\na b c d e\n"),
		},
		{
			name:     "unbroken strings begin newline",
			maxWidth: 10,
			input:    []byte("a b c defg i jk\n"),
			expected: []byte("a b c\ndefg i jk\n"),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := goldmark.DefaultParser()
			p.AddOptions(parser.WithParagraphTransformers(util.Prioritized(extension.NewTableParagraphTransformer(), 200)))

			pb := p.Parse(text.NewReader(c.input))
			mr := mdfmt.NewRenderer(
				mdfmt.WithMaxCharacterWidth(c.maxWidth),
			)
			re := renderer.NewRenderer(
				renderer.WithNodeRenderers(
					util.Prioritized(mr, 1000),
				),
			)
			b := &bytes.Buffer{}
			re.Render(b, c.input, pb)
			if !bytes.Equal(b.Bytes(), c.expected) {
				t.Errorf("rendered output does not match expected.\nOutput:\n%s\nExpected:\n%s", b.Bytes(), c.expected)
			}
		})
	}
}
