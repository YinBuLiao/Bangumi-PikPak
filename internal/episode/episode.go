package episode

import (
	"fmt"
	"regexp"
	"strconv"
)

var episodePatterns = []*regexp.Regexp{
	regexp.MustCompile(`\[(\d{1,3})\]`),
	regexp.MustCompile(`(?i)(?:^|\s)-\s*(\d{1,3})(?:\s|\(|\[|$)`),
	regexp.MustCompile(`第\s*(\d{1,3})\s*[话話集]`),
	regexp.MustCompile(`(?i)episode\s*(\d{1,3})`),
}

func LabelFromTitle(title string) (string, bool) {
	for _, pattern := range episodePatterns {
		match := pattern.FindStringSubmatch(title)
		if len(match) < 2 {
			continue
		}
		n, err := strconv.Atoi(match[1])
		if err != nil {
			continue
		}
		return fmt.Sprintf("第%02d集", n), true
	}
	return "", false
}
