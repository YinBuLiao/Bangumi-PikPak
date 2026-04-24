package rss

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

type Entry struct {
	Title         string
	Link          string
	TorrentURL    string
	PublishedDate string
}

func Parse(r io.Reader) ([]Entry, error) {
	feed, err := gofeed.NewParser().Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parse rss: %w", err)
	}
	entries := make([]Entry, 0, len(feed.Items))
	for _, item := range feed.Items {
		var torrent string
		for _, enclosure := range item.Enclosures {
			if enclosure.URL != "" {
				torrent = enclosure.URL
				break
			}
		}
		published := ""
		if item.PublishedParsed != nil {
			published = item.PublishedParsed.Format("2006-01-02")
		} else if item.Published != "" {
			if parsed, err := time.Parse(time.RFC3339, item.Published); err == nil {
				published = parsed.Format("2006-01-02")
			} else {
				published = item.Published
			}
		}
		entries = append(entries, Entry{Title: item.Title, Link: item.Link, TorrentURL: torrent, PublishedDate: published})
	}
	return entries, nil
}

func Fetch(client *http.Client, url string) ([]Entry, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch rss: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch rss: status %s", resp.Status)
	}
	return Parse(resp.Body)
}
