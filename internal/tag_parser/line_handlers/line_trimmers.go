package line_handlers

import "strings"

type GoLineCommentTrimmer struct{}

func (GoLineCommentTrimmer) Trim(line string) string {
	s := strings.TrimSpace(line)
	s = strings.TrimPrefix(s, "//")
	return strings.TrimSpace(s)
}
