package renderers

import (
	"html/template"
	"strings"

	hr "github.com/sourcehawk/cty-generator-addons/internal/html"
)

const crdTemplate = `
<div class="accordion-item">
  <button class="accordion-button collapsed" type="button" onclick="toggleAccordion(this)">
    <div style="width: 100%;">
      <div class="property-info">
        <span class="property-name">{{ .Name }}</span>
        <span class="property-type property-required">Condition Options</span>
      </div>
      <div class="property-description">Condition types for the {{ .Name }} resource.</div>
    </div>
  </button>
  <div class="collapse">
    <div class="accordion-body">
      <div class="accordion" id="conditions-{{ .SafeID }}">
        {{ if .HasChildren }}{{ .Children }}{{ else }}<p class="muted">No conditions documented for this resource.</p>{{ end }}
      </div>
    </div>
  </div>
</div>`

type CRDNode struct {
	hr.BaseHTMLGenerator

	Name string
}

func NewCRDNode(name string) *CRDNode {
	return &CRDNode{
		BaseHTMLGenerator: hr.BaseHTMLGenerator{
			Template: hr.MustParseTemplate("crd", crdTemplate),
		},
		Name: name,
	}
}

func (n *CRDNode) Generate() (template.HTML, error) {
	parts, err := n.RenderChildren()
	if err != nil {
		return "", err
	}
	safeID := strings.ToLower(strings.ReplaceAll(n.Name, " ", "-"))
	data := map[string]any{
		"Name":        n.Name,
		"HasChildren": len(parts) > 0,
		"SafeID":      safeID,
		"Children":    template.HTML(strings.Join(parts, "")),
	}
	return n.ExecTemplate("", data)
}
