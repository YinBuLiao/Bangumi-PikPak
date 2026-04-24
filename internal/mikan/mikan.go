package mikan

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"bangumi-pikpak/internal/episode"

	"github.com/PuerkitoBio/goquery"
)

const titleSelector = "p.bangumi-title"
const defaultBaseURL = "https://mikanani.me"

var timeNow = time.Now

type EpisodeMetadata struct {
	Title    string `json:"title"`
	CoverURL string `json:"cover_url,omitempty"`
}

type SearchResult struct {
	Title        string `json:"title"`
	BangumiTitle string `json:"bangumi_title,omitempty"`
	CoverURL     string `json:"cover_url,omitempty"`
	Summary      string `json:"summary,omitempty"`
	Link         string `json:"link"`
	TorrentURL   string `json:"torrent_url"`
	Magnet       string `json:"magnet,omitempty"`
	Size         string `json:"size,omitempty"`
	Updated      string `json:"updated,omitempty"`
	EpisodeLabel string `json:"episode_label,omitempty"`
}

type ScheduleResponse struct {
	Year   int            `json:"year"`
	Season string         `json:"season"`
	Days   []ScheduleDay  `json:"days"`
	Items  []ScheduleItem `json:"items"`
}

type ScheduleDay struct {
	Weekday int            `json:"weekday"`
	Label   string         `json:"label"`
	Items   []ScheduleItem `json:"items"`
}

type ScheduleItem struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	CoverURL  string `json:"cover_url,omitempty"`
	CoverFrom string `json:"cover_from,omitempty"`
	PageURL   string `json:"page_url,omitempty"`
	Updated   string `json:"updated,omitempty"`
	Weekday   int    `json:"weekday"`
	DayLabel  string `json:"day_label"`
}

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
	meta, err := FetchEpisodeMetadata(client, url)
	if err != nil {
		return "", err
	}
	return meta.Title, nil
}

func ParseEpisodeMetadata(r io.Reader, base string) (EpisodeMetadata, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return EpisodeMetadata{}, fmt.Errorf("parse mikan html: %w", err)
	}
	title := strings.TrimSpace(doc.Find(titleSelector).First().Text())
	if title == "" {
		return EpisodeMetadata{}, fmt.Errorf("mikan title selector %q not found", titleSelector)
	}
	cover := ""
	style, ok := doc.Find(".bangumi-poster").First().Attr("style")
	if ok {
		cover = backgroundImageURL(style)
	}
	if cover != "" {
		cover = absoluteURL(base, cover)
	}
	return EpisodeMetadata{Title: title, CoverURL: cover}, nil
}

func FetchEpisodeMetadata(client *http.Client, pageURL string) (EpisodeMetadata, error) {
	resp, err := client.Get(pageURL)
	if err != nil {
		return EpisodeMetadata{}, fmt.Errorf("fetch mikan page: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return EpisodeMetadata{}, fmt.Errorf("fetch mikan page: status %s", resp.Status)
	}
	base := defaultBaseURL
	if parsed, err := url.Parse(pageURL); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		base = parsed.Scheme + "://" + parsed.Host
	}
	return ParseEpisodeMetadata(resp.Body, base)
}

func Search(client *http.Client, keyword string, limit int) ([]SearchResult, error) {
	base := defaultBaseURL
	searchURL := base + "/Home/Search?searchstr=" + url.QueryEscape(keyword)
	resp, err := client.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("search mikan: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("search mikan: status %s", resp.Status)
	}
	results, err := ParseSearchResults(resp.Body, base)
	if err != nil {
		return nil, err
	}
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

func FetchSchedule(client *http.Client, year int, season string) (ScheduleResponse, error) {
	if client == nil {
		client = http.DefaultClient
	}
	if year <= 0 || strings.TrimSpace(season) == "" {
		year, season = CurrentSeason()
	}
	endpoint := defaultBaseURL + "/Home/BangumiCoverFlowByDayOfWeek?year=" + url.QueryEscape(strconv.Itoa(year)) + "&seasonStr=" + url.QueryEscape(season)
	resp, err := client.Get(endpoint)
	if err != nil {
		return ScheduleResponse{}, fmt.Errorf("fetch mikan schedule: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ScheduleResponse{}, fmt.Errorf("fetch mikan schedule: status %s", resp.Status)
	}
	schedule, err := ParseSchedule(resp.Body, defaultBaseURL)
	if err != nil {
		return ScheduleResponse{}, err
	}
	schedule.Year = year
	schedule.Season = season
	return schedule, nil
}

func ParseSchedule(r io.Reader, base string) (ScheduleResponse, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return ScheduleResponse{}, fmt.Errorf("parse mikan schedule html: %w", err)
	}
	resp := ScheduleResponse{Days: make([]ScheduleDay, 0), Items: make([]ScheduleItem, 0)}
	doc.Find(".sk-bangumi").Each(func(_ int, block *goquery.Selection) {
		weekday, _ := strconv.Atoi(strings.TrimSpace(block.AttrOr("data-dayofweek", "0")))
		label := compactSpace(block.ChildrenFiltered(".row").First().Text())
		if label == "" {
			label = weekdayLabel(weekday)
		}
		day := ScheduleDay{Weekday: weekday, Label: label, Items: make([]ScheduleItem, 0)}
		block.Find("li").Each(func(_ int, li *goquery.Selection) {
			link := li.Find("a.an-text").First()
			title := strings.TrimSpace(link.AttrOr("title", ""))
			if title == "" {
				title = compactSpace(link.Text())
			}
			if title == "" {
				return
			}
			href := strings.TrimSpace(link.AttrOr("href", ""))
			id := 0
			if span := li.Find(".js-expand_bangumi").First(); span.Length() > 0 {
				id, _ = strconv.Atoi(strings.TrimSpace(span.AttrOr("data-bangumiid", "0")))
			}
			if id == 0 {
				id = mikanBangumiID(href)
			}
			cover := ""
			if span := li.Find(".js-expand_bangumi").First(); span.Length() > 0 {
				cover = strings.TrimSpace(span.AttrOr("data-src", ""))
			}
			item := ScheduleItem{
				ID:        id,
				Title:     title,
				CoverURL:  absoluteURL(base, cover),
				CoverFrom: "mikan",
				PageURL:   absoluteURL(base, href),
				Updated:   compactSpace(li.Find(".date-text").First().Text()),
				Weekday:   weekday,
				DayLabel:  label,
			}
			day.Items = append(day.Items, item)
			resp.Items = append(resp.Items, item)
		})
		if len(day.Items) > 0 {
			resp.Days = append(resp.Days, day)
		}
	})
	return resp, nil
}

func CurrentSeason() (int, string) {
	now := timeNow()
	month := int(now.Month())
	switch {
	case month >= 1 && month <= 3:
		return now.Year(), "冬"
	case month >= 4 && month <= 6:
		return now.Year(), "春"
	case month >= 7 && month <= 9:
		return now.Year(), "夏"
	default:
		return now.Year(), "秋"
	}
}

func ParseSearchResults(r io.Reader, base string) ([]SearchResult, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("parse mikan search html: %w", err)
	}
	results := make([]SearchResult, 0)
	doc.Find("tr.js-search-results-row").Each(func(_ int, row *goquery.Selection) {
		titleLink := row.Find("a.magnet-link-wrap").First()
		title := strings.TrimSpace(titleLink.Text())
		link, _ := titleLink.Attr("href")
		magnet, _ := row.Find("input.js-episode-select").First().Attr("data-magnet")
		torrentURL := ""
		row.Find(`a[href*="/Download/"]`).EachWithBreak(func(_ int, a *goquery.Selection) bool {
			href, ok := a.Attr("href")
			if ok {
				torrentURL = href
				return false
			}
			return true
		})
		if title == "" || torrentURL == "" {
			return
		}
		cells := row.Find("td")
		size := strings.TrimSpace(cells.Eq(2).Text())
		updated := strings.TrimSpace(cells.Eq(3).Text())
		label, _ := episode.LabelFromTitle(title)
		results = append(results, SearchResult{
			Title:        title,
			Link:         absoluteURL(base, link),
			TorrentURL:   absoluteURL(base, torrentURL),
			Magnet:       strings.TrimSpace(magnet),
			Size:         size,
			Updated:      updated,
			EpisodeLabel: label,
		})
	})
	return results, nil
}

func absoluteURL(base, raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err == nil && u.IsAbs() {
		return raw
	}
	b, err := url.Parse(base)
	if err != nil {
		return raw
	}
	return b.ResolveReference(u).String()
}

func backgroundImageURL(style string) string {
	lower := strings.ToLower(style)
	idx := strings.Index(lower, "url(")
	if idx < 0 {
		return ""
	}
	rest := style[idx+4:]
	end := strings.Index(rest, ")")
	if end < 0 {
		return ""
	}
	raw := strings.TrimSpace(rest[:end])
	return strings.Trim(raw, `"'`)
}

func compactSpace(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

func weekdayLabel(day int) string {
	switch day {
	case 1:
		return "星期一"
	case 2:
		return "星期二"
	case 3:
		return "星期三"
	case 4:
		return "星期四"
	case 5:
		return "星期五"
	case 6:
		return "星期六"
	case 7:
		return "星期日"
	default:
		return "其他"
	}
}
