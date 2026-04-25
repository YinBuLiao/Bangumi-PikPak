package web

import (
	"regexp"
	"strings"
	"unicode"
)

var bracketTokenPattern = regexp.MustCompile(`[\[【]([^\]】]{1,160})[\]】]`)
var leadingBracketPattern = regexp.MustCompile(`^\s*[\[【]([^\]】]{1,160})[\]】]\s*`)
var trailingParenPattern = regexp.MustCompile(`\s*[\(（]([^\)）]{1,180})[\)）]\s*$`)
var releaseRangeSuffixPattern = regexp.MustCompile(`(?i)\s*(?:第\s*)?\d{1,3}\s*[-~～_]\s*\d{1,3}\s*(?:集|话|話|合集)?\s*$`)
var releaseEpisodeSuffixPattern = regexp.MustCompile(`(?i)\s*(?:第\s*)?\d{1,3}\s*(?:集|话|話)\s*$`)
var anniversarySuffixPattern = regexp.MustCompile(`\s*\d+\s*周年\s*$`)
var plainQualityTailPattern = regexp.MustCompile(`(?i)\s+(?:BDRip|BluRay|WebRip|WEB-DL|HEVC|AVC|H\.?26[45]|x26[45]|720p|1080p|2160p|4k)\b.*$`)

// ExtractBangumiTitle turns noisy storage/release folder names into a Bangumi
// search-friendly title. It intentionally keeps both Chinese and romanized
// title text when they appear together, but removes subgroup, episode range,
// codec, quality and subtitle tags.
func ExtractBangumiTitle(raw string) string {
	original := normalizeTitleSpaces(trimKnownTitleExtension(strings.TrimSpace(raw)))
	if original == "" {
		return strings.TrimSpace(raw)
	}
	cleaned := original
	for {
		match := leadingBracketPattern.FindStringSubmatchIndex(cleaned)
		if len(match) == 0 {
			break
		}
		token := cleaned[match[2]:match[3]]
		if !isLikelyGroupToken(token) {
			break
		}
		cleaned = strings.TrimSpace(cleaned[match[1]:])
	}
	cleaned = unwrapOrRemoveBracketTokens(cleaned)
	cleaned = stripTechnicalParenTail(cleaned)
	cleaned = stripReleaseSuffix(cleaned)
	cleaned = preferChineseSlashTitle(cleaned)
	cleaned = normalizeTitleSpaces(cleaned)
	if cleaned == "" {
		return original
	}
	return cleaned
}

func trimKnownTitleExtension(s string) string {
	lower := strings.ToLower(s)
	for _, ext := range []string{".torrent", ".mp4", ".mkv", ".webm", ".m4v", ".mov", ".avi", ".flv"} {
		if strings.HasSuffix(lower, ext) {
			return strings.TrimSpace(s[:len(s)-len(ext)])
		}
	}
	return s
}

func unwrapOrRemoveBracketTokens(s string) string {
	var out strings.Builder
	last := 0
	matches := bracketTokenPattern.FindAllStringSubmatchIndex(s, -1)
	for _, match := range matches {
		out.WriteString(s[last:match[0]])
		token := strings.TrimSpace(s[match[2]:match[3]])
		if !isLikelyMetadataToken(token) && !isLikelyGroupToken(token) {
			out.WriteString(" ")
			out.WriteString(token)
			out.WriteString(" ")
		}
		last = match[1]
	}
	out.WriteString(s[last:])
	return normalizeTitleSpaces(out.String())
}

func stripTechnicalParenTail(s string) string {
	for {
		match := trailingParenPattern.FindStringSubmatchIndex(s)
		if len(match) == 0 {
			return s
		}
		token := s[match[2]:match[3]]
		if !isLikelyMetadataToken(token) {
			return s
		}
		s = strings.TrimSpace(s[:match[0]])
	}
}

func stripReleaseSuffix(s string) string {
	s = releaseRangeSuffixPattern.ReplaceAllString(s, "")
	s = releaseEpisodeSuffixPattern.ReplaceAllString(s, "")
	s = anniversarySuffixPattern.ReplaceAllString(s, "")
	s = plainQualityTailPattern.ReplaceAllString(s, "")
	return strings.Trim(strings.TrimSpace(s), "-_·,，.。 ")
}

func preferChineseSlashTitle(s string) string {
	parts := strings.Split(s, "/")
	if len(parts) <= 1 {
		return s
	}
	first := strings.TrimSpace(parts[0])
	if first != "" && containsCJK(first) {
		return first
	}
	return s
}

func isLikelyMetadataToken(token string) bool {
	t := strings.TrimSpace(token)
	if t == "" {
		return true
	}
	compact := strings.ToLower(strings.NewReplacer(" ", "", "_", "", "-", "", ".", "").Replace(t))
	if regexp.MustCompile(`^\d{1,3}([~～]\d{1,3})?(合集|集|话|話)?$`).MatchString(compact) {
		return true
	}
	if regexp.MustCompile(`^\d{3,4}p$`).MatchString(compact) {
		return true
	}
	switch compact {
	case "op", "ed", "sp", "ova", "oad", "pv", "cm", "ncop", "nced":
		return true
	}
	keywords := []string{
		"bdrip", "bluray", "webrip", "webdl", "hevc", "avc", "h264", "h265", "x264", "x265",
		"aac", "flac", "mp4", "mkv", "ass", "chs", "cht", "gb", "big5", "yuv420", "10bit",
		"简", "繁", "字幕", "内封", "外挂", "合集", "完结", "batch", "complete",
	}
	for _, keyword := range keywords {
		if strings.Contains(compact, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func isLikelyGroupToken(token string) bool {
	t := strings.ToLower(strings.TrimSpace(token))
	if t == "" {
		return false
	}
	groupKeywords := []string{
		"字幕组", "发布组", "压制组", "字幕", "发布", "压制", "搬运", "小队",
		"raws", "raw", "fans", "fansub", "subs", "sub", "yyddm", "yydm",
		"桜都", "喵萌", "动漫国", "猎户", "异域", "风之圣殿",
	}
	for _, keyword := range groupKeywords {
		if strings.Contains(t, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func containsCJK(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) || (r >= 0x3040 && r <= 0x30ff) || (r >= 0xac00 && r <= 0xd7af) {
			return true
		}
	}
	return false
}

func normalizeTitleSpaces(s string) string {
	s = strings.NewReplacer("＿", " ", "_", " ", "　", " ").Replace(s)
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}
