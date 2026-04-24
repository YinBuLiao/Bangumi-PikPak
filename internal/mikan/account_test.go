package mikan

import (
	"strings"
	"testing"
)

func TestParseBangumiCandidate(t *testing.T) {
	html := `<p class="bangumi-title"><a href="/Home/Bangumi/3903#subgroups">测试番剧</a></p><div class="bangumi-poster" style="background-image:url('/images/Bangumi/3903.jpg')"></div>`
	got, err := ParseBangumiCandidate(strings.NewReader(html), "https://mikanani.me")
	if err != nil {
		t.Fatalf("ParseBangumiCandidate returned error: %v", err)
	}
	if got.ID != 3903 || got.Title != "测试番剧" || got.PageURL != "https://mikanani.me/Home/Bangumi/3903#subgroups" || got.CoverURL == "" {
		t.Fatalf("candidate mismatch: %#v", got)
	}
}

func TestParseRequestVerificationToken(t *testing.T) {
	token, err := parseRequestVerificationToken(strings.NewReader(`<input name="__RequestVerificationToken" value="abc123">`))
	if err != nil {
		t.Fatalf("parseRequestVerificationToken returned error: %v", err)
	}
	if token != "abc123" {
		t.Fatalf("token mismatch: %q", token)
	}
}

func TestParseBangumiSubjectID(t *testing.T) {
	got, err := ParseBangumiSubjectID(strings.NewReader(`<a href="https://bgm.tv/subject/458684">Bangumi番组计划链接</a>`))
	if err != nil {
		t.Fatalf("ParseBangumiSubjectID returned error: %v", err)
	}
	if got != 458684 {
		t.Fatalf("subject id mismatch: %d", got)
	}
}
