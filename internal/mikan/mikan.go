package mikan

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const titleSelector = "p.bangumi-title"

func ParseTitle(r io.Reader) (string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", fmt.Errorf("parse mikan html: %w", err)
	}
	title := strings.TrimSpace(doc.Find(titleSelector).First().Text())
	if title == "" {
		return "", fmt.Errorf("mikan title selector %q not found", titleSelector)
	}
	return title, nil
}

func FetchTitle(client *http.Client, url string) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("fetch mikan page: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("fetch mikan page: status %s", resp.Status)
	}
	return ParseTitle(resp.Body)
}
