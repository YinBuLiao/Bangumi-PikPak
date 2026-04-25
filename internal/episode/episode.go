package episode

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var rangePatterns = []*regexp.Regexp{
	regexp.MustCompile(`\[(\d{1,3})\s*[-~～_]\s*(\d{1,3})\]`),
}

var episodePatterns = []*regexp.Regexp{
	regexp.MustCompile(`\[(\d{1,3})\]`),
	regexp.MustCompile(`(?i)(?:^|\s)-\s*(\d{1,3})(?:\s|\(|\[|$)`),
	regexp.MustCompile(`第\s*(\d{1,3})\s*[话話集]`),
	regexp.MustCompile(`(?i)episode\s*(\d{1,3})`),
}

var bdResolutionPattern = regexp.MustCompile(`(?i)\bbd\s+\d{3,4}x\d{3,4}`)

func LabelFromTitle(title string) (string, bool) {
	for _, pattern := range rangePatterns {
		match := pattern.FindStringSubmatch(title)
		if len(match) < 3 {
			continue
		}
		start, err1 := strconv.Atoi(match[1])
		end, err2 := strconv.Atoi(match[2])
		if err1 != nil || err2 != nil || start <= 0 || end <= 0 || end < start {
			continue
		}
		return fmt.Sprintf("第%02d-%02d集", start, end), true
	}
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
	if looksLikeBatch(title) {
		return "合集", true
	}
	return "", false
}

func looksLikeBatch(title string) bool {
	normalized := strings.ToLower(strings.TrimSpace(title))
	if normalized == "" {
		return false
	}
	keywords := []string{
		"合集", "全集", "全卷", "complete", "batch", "bdrip", "blu-ray", "bluray",
	}
	for _, keyword := range keywords {
		if strings.Contains(normalized, keyword) {
			return true
		}
	}
	return bdResolutionPattern.FindString(normalized) != ""
}
