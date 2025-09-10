package tag_parsers

import (
	"fmt"
	"regexp"

	tp "github.com/sourcehawk/cty-generator-addons/internal/tag_parser"
)

const DocTagCondition tp.DocTagType = "condition"

// // +cty:condition:for=ZeebeCluster
var reCondTag = regexp.MustCompile(
	`^\s*//.*\+cty:condition:for\s*=\s*(?P<crd>\S+)`,
)

// ConditionTagParser parses lines like: // +cty:condition:for=ZeebeCluster
type ConditionTagParser struct{}

func (ConditionTagParser) Matches(line string) bool {
	return reCondTag.MatchString(line)
}

func (ConditionTagParser) ParseTag(tagLine string) (map[string]string, error) {
	m := reCondTag.FindStringSubmatch(tagLine)
	if m == nil {
		return nil, fmt.Errorf("missing value for +cty:condition:for")
	}
	crd := m[reCondTag.SubexpIndex("crd")]
	if crd == "" {
		return nil, fmt.Errorf("could not capture CRD from +cty:condition:for")
	}
	return map[string]string{"crd": crd}, nil
}

func (ConditionTagParser) ParseVariable(varLine string) (map[string]string, error) {
	constName := firstIdent(varLine)
	if constName == "" {
		return nil, fmt.Errorf("could not parse const name from: %q", varLine)
	}
	lit, ok := firstStringLiteral(varLine)
	if !ok {
		// fallback to identifier as the Name
		return map[string]string{
			"const": constName,
			"value": constName,
		}, nil
	}
	return map[string]string{
		"const": constName,
		"value": lit,
	}, nil
}

func (ConditionTagParser) Type() tp.DocTagType { return DocTagCondition }
