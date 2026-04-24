package bangumi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseSubjectSearchResponsePrefersLargeCover(t *testing.T) {
	raw := []byte(`{"data":[{"id":1,"type":2,"name":"Otonari","name_cn":"邻家的天使","images":{"large":"https://lain.bgm.tv/pic/cover/l/a.jpg","common":"https://lain.bgm.tv/pic/cover/c/a.jpg"}}],"total":1}`)
	subjects, err := ParseSubjectSearchResponse(raw)
	if err != nil {
		t.Fatalf("ParseSubjectSearchResponse returned error: %v", err)
	}
	if len(subjects) != 1 {
		t.Fatalf("expected one subject, got %d", len(subjects))
	}
	if subjects[0].CoverURL() != "https://lain.bgm.tv/pic/cover/l/a.jpg" {
		t.Fatalf("cover mismatch: %#v", subjects[0])
	}
}

func TestClientSearchCoverSendsAnimeFilterAndUserAgent(t *testing.T) {
	var body string
	var ua string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v0/search/subjects":
			b := make([]byte, r.ContentLength)
			_, _ = r.Body.Read(b)
			body = string(b)
			_, _ = w.Write([]byte(`{"data":[{"id":2,"type":2,"name":"Test","name_cn":"测试","images":{"common":"https://lain.bgm.tv/pic/cover/c/search.jpg"}}],"total":1}`))
		case "/v0/subjects/2":
			_, _ = w.Write([]byte(`{"id":2,"type":2,"name":"Test","name_cn":"测试","images":{"common":"https://lain.bgm.tv/pic/cover/c/test.jpg"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	c := Client{HTTPClient: srv.Client(), BaseURL: srv.URL, UserAgent: "tester/bangumi-pikpak"}
	cover, err := c.SearchCover("测试番剧")
	if err != nil {
		t.Fatalf("SearchCover returned error: %v", err)
	}
	if cover != "https://lain.bgm.tv/pic/cover/c/test.jpg" {
		t.Fatalf("cover mismatch: %q", cover)
	}
	if ua != "tester/bangumi-pikpak" {
		t.Fatalf("user-agent mismatch: %q", ua)
	}
	if body == "" || !containsAll(body, "测试番剧", "\"type\":[2]") {
		t.Fatalf("request body mismatch: %s", body)
	}
}

func TestClientSearchMetadataFetchesSubjectDetailSummary(t *testing.T) {
	var detailCalled bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v0/search/subjects":
			_, _ = w.Write([]byte(`{"data":[{"id":7,"type":2,"name":"Oshi no Ko","name_cn":"我推的孩子","images":{"common":"https://lain.bgm.tv/pic/cover/c/search.jpg"}}],"total":1}`))
		case "/v0/subjects/7":
			detailCalled = true
			_, _ = w.Write([]byte(`{"id":7,"type":2,"name":"Oshi no Ko","name_cn":"我推的孩子","summary":"少女偶像与转生医生的故事。","images":{"large":"https://lain.bgm.tv/pic/cover/l/detail.jpg"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := Client{HTTPClient: srv.Client(), BaseURL: srv.URL, UserAgent: "tester/bangumi-pikpak"}
	meta, err := c.SearchMetadata("我推的孩子")
	if err != nil {
		t.Fatalf("SearchMetadata returned error: %v", err)
	}
	if !detailCalled {
		t.Fatal("expected subject detail endpoint to be called")
	}
	if meta.Title != "我推的孩子" || meta.CoverURL != "https://lain.bgm.tv/pic/cover/l/detail.jpg" || meta.Summary != "少女偶像与转生医生的故事。" {
		t.Fatalf("metadata mismatch: %#v", meta)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(s, part) {
			return false
		}
	}
	return true
}
