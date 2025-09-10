package main

// cty-conditions-doc: parse +cty tags and render Camunda-styled HTML using component generators.

// NOTE: plenty of this hacks package is AI generated, I didn't feel like this was a big deal seeing as this is
// not an important part of our code but rather a small hack project to generate additional doc comments

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	hrend "github.com/sourcehawk/cty-generator-addons/internal/html/renderers"
	tp "github.com/sourcehawk/cty-generator-addons/internal/tag_parser"
	lhs "github.com/sourcehawk/cty-generator-addons/internal/tag_parser/line_handlers"
	tps "github.com/sourcehawk/cty-generator-addons/internal/tag_parser/tag_parsers"
	"golang.org/x/net/html"
)

// -------- Domain types --------

type ReasonDoc struct {
	Name        string // string literal value (e.g. "CreationError")
	ConstName   string // Go const identifier
	Description string // from comments
}

type ConditionDoc struct {
	Name        string // string literal value (e.g. "Ready", "EncryptionReady")
	ConstName   string // Go const identifier
	Description string
	Reasons     []ReasonDoc
}

type CRD struct {
	Name       string
	Conditions []ConditionDoc
}

func main() {
	path := flag.String("path", ".", "root directory to scan (recursively)")
	title := flag.String("title", "Conditions Reference", "Section title")
	injectPath := flag.String("inject-into", "index.html", "CTY index.html to modify in-place (append to last .content)")

	flag.Parse()

	fp := tp.NewFileTagParser(
		[]tp.DocTagParser{tps.ConditionTagParser{}, tps.ReasonTagParser{}},
		lhs.GoLineCommentMatcher{},
		lhs.GoLineCommentTrimmer{},
		lhs.GoLineCommentBulletPointMatcher{},
	)

	// Walk repo and collect all tag results.
	var all []*tp.DocTagResult
	err := filepath.WalkDir(*path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := filepath.Base(p)
			if base == "vendor" || strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go") {
			res, perr := fp.ParseTags(p)
			if perr != nil {
				return perr
			}
			all = append(all, res...)
		}
		return nil
	})
	if err != nil {
		failf("walk error: %v", err)
	}

	// Aggregate into CRD -> Conditions -> Reasons
	crds := buildCRDConditionsFromResults(all)

	// Build the component tree and render
	section := hrend.NewSectionNode(*title)
	for _, crd := range crds {
		crdNode := hrend.NewCRDNode(crd.Name)

		for _, cond := range crd.Conditions {
			condNode := hrend.NewConditionNode(cond.Name, cond.Description)

			for _, r := range cond.Reasons {
				condNode.AddChild(hrend.NewReasonNode(r.Name, r.Description))
			}
			crdNode.AddChild(condNode)
		}

		section.AddChild(crdNode)
	}

	htmlOut, err := section.Generate()
	if err != nil {
		failf("render error: %v", err)
	}

	if *injectPath != "" {
		baseBytes, err := os.ReadFile(*injectPath)
		if err != nil {
			failf("read %s: %v", *injectPath, err)
		}

		merged, err := injectIntoContentString(string(baseBytes), string(htmlOut))
		if err != nil {
			failf("inject: %v")
		}

		if err := os.WriteFile(*injectPath, []byte(merged), 0o644); err != nil {
			failf("write %s: %v", *injectPath, err)
		}
		return
	}
}

// -------- Aggregation --------

func buildCRDConditionsFromResults(results []*tp.DocTagResult) []CRD {
	// crd -> condName -> *ConditionDoc
	crdMap := map[string]map[string]*ConditionDoc{}

	getCRDSet := func(name string) map[string]*ConditionDoc {
		if crdMap[name] == nil {
			crdMap[name] = map[string]*ConditionDoc{}
		}
		return crdMap[name]
	}

	for _, r := range results {
		switch r.Type {
		case tps.DocTagCondition:
			crdName := r.TagValues["crd"]
			condName := r.Variable["value"]
			constID := r.Variable["const"]
			if crdName == "" || condName == "" {
				continue
			}
			crdSet := getCRDSet(crdName)
			if crdSet[condName] == nil {
				crdSet[condName] = &ConditionDoc{
					Name:        condName,
					ConstName:   constID,
					Description: strings.TrimSpace(r.Comment),
				}
			} else if crdSet[condName].Description == "" && strings.TrimSpace(r.Comment) != "" {
				crdSet[condName].Description = strings.TrimSpace(r.Comment)
			}

		case tps.DocTagReason:
			crdName := r.TagValues["crd"]
			condName := r.TagValues["condition"]
			reasonName := r.Variable["value"]
			constID := r.Variable["const"]
			if crdName == "" || condName == "" || reasonName == "" {
				continue
			}
			crdSet := getCRDSet(crdName)
			if crdSet[condName] == nil {
				// placeholder condition if declared later/elsewhere
				crdSet[condName] = &ConditionDoc{Name: condName}
			}
			addReasonUnique(&crdSet[condName].Reasons, ReasonDoc{
				Name:        reasonName,
				ConstName:   constID,
				Description: strings.TrimSpace(r.Comment),
			})
		}
	}

	// materialize & sort
	var crds []CRD
	for crdName, set := range crdMap {
		var conds []ConditionDoc
		for _, c := range set {
			sort.Slice(c.Reasons, func(i, j int) bool { return c.Reasons[i].Name < c.Reasons[j].Name })
			conds = append(conds, *c)
		}
		sort.Slice(conds, func(i, j int) bool { return conds[i].Name < conds[j].Name })
		crds = append(crds, CRD{Name: crdName, Conditions: conds})
	}
	sort.Slice(crds, func(i, j int) bool { return crds[i].Name < crds[j].Name })
	return crds
}

func addReasonUnique(slice *[]ReasonDoc, r ReasonDoc) {
	for _, ex := range *slice {
		if ex.Name == r.Name || ex.ConstName == r.ConstName {
			return
		}
	}
	*slice = append(*slice, r)
}

func failf(f string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, f+"\n", a...)
	os.Exit(1)
}

// injectIntoContentString appends fragment HTML to the last <div class="content"> in base.
// It returns the full updated HTML as a string.
func injectIntoContentString(base, fragment string) (string, error) {
	doc, err := html.Parse(strings.NewReader(base))
	if err != nil {
		return "", err
	}

	// Find all <div class="content">; append to the last one.
	var contents []*html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" && hasClass(n, "content") {
			contents = append(contents, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	if len(contents) == 0 {
		return "", errors.New(`no <div class="content"> found`)
	}
	target := contents[len(contents)-1]

	// Parse the fragment in the context of the target and append
	nodes, err := html.ParseFragment(strings.NewReader(fragment), target)
	if err != nil {
		return "", err
	}
	for _, n := range nodes {
		target.AppendChild(n)
	}

	var out bytes.Buffer
	if err := html.Render(&out, doc); err != nil {
		return "", err
	}
	return out.String(), nil
}

func hasClass(n *html.Node, class string) bool {
	for _, a := range n.Attr {
		if a.Key != "class" {
			continue
		}
		for _, c := range strings.Fields(a.Val) {
			if c == class {
				return true
			}
		}
	}
	return false
}
