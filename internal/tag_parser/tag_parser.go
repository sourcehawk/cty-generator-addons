package tag_parser

import (
	"fmt"
	"os"
	"strings"
)

type DocTagType string

type DocTagParser interface {
	// Matches returns true if this parsers tag matches the line in the file
	Matches(line string) bool
	// ParseTag parses the desired values from a documentation tag
	ParseTag(tagLine string) (map[string]string, error)
	// ParseVariable parses the desired values from the variable that has the documentation tag
	ParseVariable(varLine string) (map[string]string, error)
	// Type Returns this DocTags DocTagType
	Type() DocTagType
}

type LineMatcher interface {
	Matches(line string) bool
}

type LineTrimmer interface {
	Trim(line string) string
}

type DocTagResult struct {
	Comment   string
	Variable  map[string]string
	TagValues map[string]string
	Type      DocTagType
}

type FileDocTagParser struct {
	parsers        []DocTagParser
	commentMatcher LineMatcher
	commentTrimmer LineTrimmer
	bulletMatcher  LineMatcher
}

func NewFileTagParser(
	parsers []DocTagParser,
	commentMatcher LineMatcher, commentTrimmer LineTrimmer, bulletMatcher LineMatcher,
) *FileDocTagParser {
	return &FileDocTagParser{
		parsers:        parsers,
		commentMatcher: commentMatcher,
		commentTrimmer: commentTrimmer,
		bulletMatcher:  bulletMatcher,
	}
}

func (ftp *FileDocTagParser) ParseTags(filename string) ([]*DocTagResult, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	var results []*DocTagResult

	for i, line := range lines {
		for _, parser := range ftp.parsers {
			if parser.Matches(line) {
				tagValues, err := parser.ParseTag(line)
				if err != nil {
					return nil, fmt.Errorf("%s:%d\n\t%w", filename, i+1, err)
				}
				variableLine, err := getFirstNonCommentLineAfter(i, lines, ftp.commentMatcher)
				if err != nil {
					return nil, fmt.Errorf("%s:%d\n\t%w", filename, i+1, err)
				}
				variableValues, err := parser.ParseVariable(variableLine)
				if err != nil {
					return nil, fmt.Errorf("%s:%d\n\t%w", filename, i+1, err)
				}

				results = append(results, &DocTagResult{
					Comment: strings.Join(
						getSurroundingCommentLines(
							i, lines,
							ftp.commentMatcher, ftp.commentTrimmer, ftp.bulletMatcher,
						), "\n",
					),
					Variable:  variableValues,
					TagValues: tagValues,
					Type:      parser.Type(),
				})
			}
		}
	}

	return results, nil
}

func getSurroundingCommentLines(
	aroundIndex int, lines []string,
	commentMatcher LineMatcher, commentTrimmer LineTrimmer, bulletMatcher LineMatcher,
) []string {
	// 1) collect comment lines before/after FROM THE FULL FILE
	before := getCommentLinesBefore(aroundIndex, lines, commentMatcher, commentTrimmer)
	after := getCommentLinesAfter(aroundIndex, lines, commentMatcher, commentTrimmer)

	// (optional) put "before" back into natural top→down order
	for i, j := 0, len(before)-1; i < j; i, j = i+1, j-1 {
		before[i], before[j] = before[j], before[i]
	}

	commentLines := append(before, after...)

	// 2) group non-bullet lines into paragraphs, pass bullet lines through
	var joined []string
	var current strings.Builder

	flush := func() {
		if current.Len() == 0 {
			return
		}
		joined = append(joined, current.String())
		current.Reset()
	}

	for _, line := range commentLines {
		if bulletMatcher.Matches(line) {
			flush()
			joined = append(joined, line)
			continue
		}
		if current.Len() > 0 {
			current.WriteByte('\n') // keep line breaks inside paragraph
		}
		current.WriteString(line)
	}

	flush() // 3) make sure the trailing paragraph is included
	return joined
}

func getCommentLinesBefore(
	index int, lines []string,
	commentMatcher LineMatcher, commentTrimmer LineTrimmer,
) []string {
	var commentLines []string

	if index <= 0 {
		return commentLines
	}

	start := index
	for {
		start--
		if start < 0 {
			break
		}

		line := lines[start]
		line = strings.TrimSpace(line)

		if !commentMatcher.Matches(line) {
			break
		}
		commentLines = append(commentLines, commentTrimmer.Trim(line))
	}
	
	return commentLines
}

func getCommentLinesAfter(
	index int, lines []string,
	commentMatcher LineMatcher, commentTrimmer LineTrimmer,
) []string {
	var commentLines []string

	if index >= len(lines)-1 {
		return commentLines
	}

	start := index
	for {
		start++
		if start >= len(lines) {
			break
		}

		line := lines[start]
		line = strings.TrimSpace(line)

		if !commentMatcher.Matches(line) {
			break
		}
		commentLines = append(commentLines, commentTrimmer.Trim(line))
	}

	return commentLines
}

func getFirstNonCommentLineAfter(index int, lines []string, commentMatcher LineMatcher) (string, error) {
	if index >= len(lines)-1 {
		return "", fmt.Errorf("could not find a variable declaration after tag declaration at line %d", index+1)
	}

	start := index
	for {
		start++
		if start >= len(lines) {
			break
		}

		line := lines[start]
		line = strings.TrimSpace(line)

		if !commentMatcher.Matches(line) {
			return line, nil
		}
	}

	return "", fmt.Errorf("could not find a variable declaration after tag declaration at line %d", index+1)
}
