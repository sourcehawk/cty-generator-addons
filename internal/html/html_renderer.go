package html

import (
	"bytes"
	"html"
	"html/template"
	"strings"
)

type Generator interface {
	Generate() (template.HTML, error)
	AddChild(child Generator)
}

type BaseHTMLGenerator struct {
	Template *template.Template
	Children []Generator
}

func (g *BaseHTMLGenerator) RenderChildren() ([]string, error) {
	var parts []string

	for _, child := range g.Children {
		h, err := child.Generate()
		if err != nil {
			return nil, err
		}
		parts = append(parts, string(h))
	}

	return parts, nil
}

func (g *BaseHTMLGenerator) AddChild(child Generator) {
	g.Children = append(g.Children, child)
}

// ExecTemplate a helper used by nodes to execute their local template.
func (g *BaseHTMLGenerator) ExecTemplate(name string, data any) (template.HTML, error) {
	var buf bytes.Buffer
	if name == "" {
		if err := g.Template.Execute(&buf, data); err != nil {
			return "", err
		}
	} else {
		if err := g.Template.ExecuteTemplate(&buf, name, data); err != nil {
			return "", err
		}
	}
	return template.HTML(buf.String()), nil
}

// Converts plain text (with lines starting `-`, `•`, or `*`) into safe <p>/<ul>/<li> HTML.
// Empty lines create paragraph breaks. Consecutive non-bullet lines become a single paragraph.
func formatCommentHTML(s string) string {
	if s == "" {
		return ""
	}
	lines := strings.Split(s, "\n")

	var out []string
	var para []string
	var bullets []string

	isBullet := func(t string) bool {
		// treat indented bullets as bullets too
		t = strings.TrimLeft(t, " \t")
		return strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* ") || strings.HasPrefix(t, "• ")
	}

	flushPara := func() {
		if len(para) == 0 {
			return
		}
		// join with a single space so wrapped comments become one paragraph
		text := html.EscapeString(strings.Join(para, " "))
		out = append(out, "<p>"+text+"</p>")
		para = nil
	}

	flushBullets := func() {
		if len(bullets) == 0 {
			return
		}
		out = append(out, "<ul>")
		for _, b := range bullets {
			// remove leading marker once more (handles mixed indents)
			bt := strings.TrimSpace(b)
			bt = strings.TrimLeft(bt, " \t")
			bt = strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(bt, "- "), "• "), "* ")
			out = append(out, "<li>"+html.EscapeString(bt)+"</li>")
		}
		out = append(out, "</ul>")
		bullets = nil
	}

	for _, ln := range lines {
		t := strings.TrimRightFunc(ln, func(r rune) bool { return r == ' ' || r == '\t' })
		ts := strings.TrimSpace(t)

		switch {
		case ts == "":
			// Empty comment line => paragraph break
			flushBullets()
			flushPara()

		case isBullet(t):
			// Bulleted item
			flushPara()
			bullets = append(bullets, t)

		default:
			// Paragraph text
			flushBullets()
			para = append(para, ts)
		}
	}

	flushBullets()
	flushPara()
	return strings.Join(out, "\n")
}

// Make the helper available to templates (return template.HTML so it doesn't get escaped again)
var tmplFuncs = template.FuncMap{
	"formatComment": func(s string) template.HTML {
		return template.HTML(formatCommentHTML(s))
	},
}

func MustParseTemplate(localName, src string) *template.Template {
	return template.Must(template.New(localName).Funcs(tmplFuncs).Parse(src))
}
