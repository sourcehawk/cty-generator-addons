package line_handlers

import "strings"

type GoLineCommentMatcher struct{}

func (GoLineCommentMatcher) Matches(line string) bool {
	s := strings.TrimSpace(line)
	return strings.HasPrefix(s, "//")
}

type GoLineCommentBulletPointMatcher struct{}

func (GoLineCommentBulletPointMatcher) Matches(line string) bool {
	s := strings.TrimSpace(line)
	return strings.HasPrefix(s, "//  -") ||
		strings.HasPrefix(s, "// -") ||
		strings.HasPrefix(s, "//  *") ||
		strings.HasPrefix(s, "// *")
}
