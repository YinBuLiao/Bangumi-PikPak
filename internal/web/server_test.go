package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandlerServesVueDistIndexForSPARoutes(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<div id="app"></div><script type="module" src="/assets/app.js"></script>`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "app.js"), []byte(`console.log("vite")`), 0o600); err != nil {
		t.Fatal(err)
	}

	handler := Server{StaticDir: dir}.Handler()
	for _, path := range []string{"/", "/library/test"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s status=%d body=%s", path, rr.Code, rr.Body.String())
		}
		if !strings.Contains(rr.Body.String(), `type="module"`) {
			t.Fatalf("%s did not serve Vue index: %s", path, rr.Body.String())
		}
	}
}

func TestHandlerServesVueDistAsset(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(`index`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "app.js"), []byte(`console.log("vite")`), 0o600); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
	rr := httptest.NewRecorder()
	Server{StaticDir: dir}.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if strings.TrimSpace(rr.Body.String()) != `console.log("vite")` {
		t.Fatalf("asset body mismatch: %s", rr.Body.String())
	}
}

func TestCJKFallbackKeywordsIncludesShortSearchableTerm(t *testing.T) {
	got := cjkFallbackKeywords("海豹宝宝")
	if len(got) == 0 {
		t.Fatal("expected fallback keywords")
	}
	found := false
	for _, keyword := range got {
		if keyword == "海豹" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 海豹 fallback, got %#v", got)
	}
}
