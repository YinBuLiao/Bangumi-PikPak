package mikan

import (
	"strings"
	"testing"
)

func TestParseTitle(t *testing.T) {
	html := `<html><body><p class="bangumi-title">  进击的巨人 最终季  </p></body></html>`
	title, err := ParseTitle(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseTitle returned error: %v", err)
	}
	if title != "进击的巨人 最终季" {
		t.Fatalf("title mismatch: %q", title)
	}
}

func TestParseTitleMissingSelector(t *testing.T) {
	_, err := ParseTitle(strings.NewReader(`<html></html>`))
	if err == nil {
		t.Fatal("expected error")
	}
}
