package rss

import (
	"strings"
	"testing"
)

func TestParseEntries(t *testing.T) {
	feed := `<?xml version="1.0" encoding="UTF-8"?>
	<rss version="2.0"><channel><title>Mikan</title><item>
	<title>[Group] Test 01</title>
	<link>https://mikanani.me/Home/Episode/abc</link>
	<pubDate>Fri, 24 Apr 2026 12:34:56 +0800</pubDate>
	<enclosure url="https://mikanani.me/Download/test.torrent" type="application/x-bittorrent" />
	</item></channel></rss>`
	entries, err := Parse(strings.NewReader(feed))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entry count mismatch: %d", len(entries))
	}
	got := entries[0]
	if got.Title != "[Group] Test 01" || got.Link != "https://mikanani.me/Home/Episode/abc" || got.TorrentURL != "https://mikanani.me/Download/test.torrent" || got.PublishedDate != "2026-04-24" {
		t.Fatalf("entry mismatch: %#v", got)
	}
}
