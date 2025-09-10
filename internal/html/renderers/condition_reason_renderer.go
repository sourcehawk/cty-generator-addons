package renderers

import (
	"html/template"

	hr "github.com/sourcehawk/cty-generator-addons/internal/html"
)

const reasonTemplate = `
<div class="accordion-item-static">
  <div class="property-info">
    <span class="property-name">{{ .Name }}</span>
	<span class="property-type property-required">Reason Type</span>
    <span class="property-type">string</span>
  </div>
  {{ if .Description }}<div class="property-description">{{ formatComment .Description }}</div>{{ end }}
</div>`

type ReasonNode struct {
	hr.BaseHTMLGenerator

	Name        string
	Description string
}

func NewReasonNode(name, description string) *ReasonNode {
	return &ReasonNode{
		BaseHTMLGenerator: hr.BaseHTMLGenerator{
			Template: hr.MustParseTemplate("reason", reasonTemplate),
		},
		Name:        name,
		Description: description,
	}
}

func (n *ReasonNode) Generate() (template.HTML, error) {
	data := map[string]any{
		"Name":        n.Name,
		"Description": n.Description,
	}
	return n.ExecTemplate("", data)
}
