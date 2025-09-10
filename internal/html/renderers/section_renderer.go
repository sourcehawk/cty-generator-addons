package renderers

import (
	"html/template"
	"strings"

	hr "github.com/sourcehawk/cty-generator-addons/internal/html"
)

const sectionTemplate = `
<div class="card">
  <div class="card-header">
    <span class="icon icon-cog"></span>
    <div>
      <strong>{{ .Title }}</strong>
      <div style="font-size: 0.9rem; opacity: 0.8;">Condition types &amp; reasons</div>
    </div>
  </div>

  <div class="card-body">
    <div class="accordion" id="conditions-root">
      {{ if .HasChildren }}{{ .Children }}{{ else }}<p class="muted">No conditions found.</p>{{ end }}
    </div>
  </div>
</div>`

type SectionNode struct {
	hr.BaseHTMLGenerator
	Title string
}

func NewSectionNode(title string) *SectionNode {
	return &SectionNode{
		BaseHTMLGenerator: hr.BaseHTMLGenerator{
			Template: hr.MustParseTemplate("section", sectionTemplate),
		},
		Title: title,
	}
}

func (n *SectionNode) Generate() (template.HTML, error) {
	parts, err := n.RenderChildren()
	if err != nil {
		return "", err
	}
	data := map[string]any{
		"Title":       n.Title,
		"HasChildren": len(parts) > 0,
		"Children":    template.HTML(strings.Join(parts, "")),
	}
	return n.ExecTemplate("", data)
}
