package renderers

import (
	"html/template"
	"strings"

	hr "github.com/sourcehawk/cty-generator-addons/internal/html"
)

const conditionTemplate = `
<div class="accordion-item">
  <button class="accordion-button collapsed" type="button" onclick="toggleAccordion(this)">
    <div style="width: 100%;">
      <div class="property-info">
        <span class="property-name">{{ .Name }}</span>
        <span class="property-type property-required">Condition Type</span>
		<span class="property-type">string</span>
      </div>
      {{ if .Description }}<div class="property-description">{{ formatComment .Description }}</div>{{ end }}
    </div>
  </button>
  <div class="collapse">
    <div class="accordion-body">
      <h4 class="d-flex align-items-center gap-2 mb-4">
        Reasons
      </h4>
      <p>
		<span class="icon icon-info"></span>
        Possible reasons for the condition
	  </p>
      <div class="accordion" id="reasons-{{ .SafeID }}">
        {{ if .HasChildren }}{{ .Children }}{{ else }}<p class="muted">No specific reasons documented.</p>{{ end }}
      </div>
    </div>
  </div>
</div>`

type ConditionNode struct {
	hr.BaseHTMLGenerator

	Name        string
	Description string
}

func NewConditionNode(name, description string) *ConditionNode {
	return &ConditionNode{
		BaseHTMLGenerator: hr.BaseHTMLGenerator{
			Template: hr.MustParseTemplate("condition", conditionTemplate),
		},
		Name:        name,
		Description: description,
	}
}

func (n *ConditionNode) Generate() (template.HTML, error) {
	parts, err := n.RenderChildren()
	if err != nil {
		return "", err
	}
	safeID := strings.ToLower(strings.ReplaceAll(n.Name, " ", "-"))
	data := map[string]any{
		"Name":        n.Name,
		"Description": n.Description,
		"SafeID":      safeID,
		"HasChildren": len(parts) > 0,
		"Children":    template.HTML(strings.Join(parts, "")),
	}
	return n.ExecTemplate("", data)
}
