package tag_parsers

import (
	"fmt"
	"regexp"

	tp "github.com/sourcehawk/cty-generator-addons/internal/tag_parser"
)

const DocTagReason tp.DocTagType = "reason"

// // +cty:reason:for=ZeebeCluster/EncryptionReady
// allows optional spaces around '=' and '/'
var reReasonTag = regexp.MustCompile(
	`^\s*//.*\+cty:reason:for\s*=\s*(?P<crd>[^\s/]+)\s*/\s*(?P<condition>\S+)`,
)

// ReasonTagParser parses lines like: // +cty:reason:for=ZeebeCluster/EncryptionReady
type ReasonTagParser struct{}

func (ReasonTagParser) Matches(line string) bool {
	return reReasonTag.MatchString(line)
}

func (ReasonTagParser) ParseTag(tagLine string) (map[string]string, error) {
	m := reReasonTag.FindStringSubmatch(tagLine)
	if m == nil {
		return nil, fmt.Errorf("missing value for +cty:reason:for")
	}
	crd := m[reReasonTag.SubexpIndex("crd")]
	cond := m[reReasonTag.SubexpIndex("condition")]
	if crd == "" || cond == "" {
		return nil, fmt.Errorf("invalid +cty:reason:for, expected <CRD>/<Condition>")
	}
	return map[string]string{"crd": crd, "condition": cond}, nil
}

func (ReasonTagParser) ParseVariable(varLine string) (map[string]string, error) {
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

func (ReasonTagParser) Type() tp.DocTagType { return DocTagReason }
