package bangumi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.bgm.tv"
const defaultUserAgent = "YinBuLiao/Bangumi-PikPak/1.0 (https://github.com/YinBuLiao/Bangumi-PikPak)"

type Client struct {
	HTTPClient *http.Client
	BaseURL    string
	UserAgent  string
}

type Subject struct {
	ID         int        `json:"id"`
	Type       int        `json:"type"`
	Name       string     `json:"name"`
	NameCN     string     `json:"name_cn"`
	Summary    string     `json:"summary"`
	Date       string     `json:"date"`
	AirDate    string     `json:"air_date"`
	AirWeekday int        `json:"air_weekday"`
	Rank       int        `json:"rank"`
	Platform   string     `json:"platform"`
	Rating     Rating     `json:"rating"`
	Collection Collection `json:"collection"`
	Images     struct {
		Large  string `json:"large"`
		Common string `json:"common"`
		Medium string `json:"medium"`
		Small  string `json:"small"`
		Grid   string `json:"grid"`
	} `json:"images"`
}

type Rating struct {
	Total int     `json:"total"`
	Score float64 `json:"score"`
}

type Collection struct {
	Wish    int `json:"wish"`
	Collect int `json:"collect"`
	Doing   int `json:"doing"`
	OnHold  int `json:"on_hold"`
	Dropped int `json:"dropped"`
}

type Metadata struct {
	ID       int    `json:"id,omitempty"`
	Title    string `json:"title,omitempty"`
	CoverURL string `json:"cover_url,omitempty"`
	Summary  string `json:"summary,omitempty"`
}

type DiscoverSubject struct {
	ID         int     `json:"id"`
	Title      string  `json:"title"`
	Name       string  `json:"name,omitempty"`
	CoverURL   string  `json:"cover_url,omitempty"`
	Summary    string  `json:"summary,omitempty"`
	Score      float64 `json:"score,omitempty"`
	Rank       int     `json:"rank,omitempty"`
	AirDate    string  `json:"air_date,omitempty"`
	AirWeekday int     `json:"air_weekday,omitempty"`
	Doing      int     `json:"doing,omitempty"`
	Collection int     `json:"collection,omitempty"`
}

type subjectSearchResponse struct {
	Data  []Subject `json:"data"`
	Total int       `json:"total"`
}

func ParseSubjectSearchResponse(raw []byte) ([]Subject, error) {
	var resp subjectSearchResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse bangumi search response: %w", err)
	}
	return resp.Data, nil
}

func (s Subject) CoverURL() string {
	for _, candidate := range []string{s.Images.Large, s.Images.Common, s.Images.Medium, s.Images.Grid, s.Images.Small} {
		if strings.TrimSpace(candidate) != "" {
			return candidate
		}
	}
	return ""
}

func (c Client) SearchCover(keyword string) (string, error) {
	meta, err := c.SearchMetadata(keyword)
	if err != nil {
		return "", err
	}
	return meta.CoverURL, nil
}

func (c Client) SearchMetadata(keyword string) (Metadata, error) {
	subjects, err := c.SearchSubjects(keyword, 5)
	if err != nil {
		return Metadata{}, err
	}
	for _, subject := range subjects {
		if subject.Type != 2 {
			continue
		}
		detail, err := c.GetSubject(subject.ID)
		if err != nil {
			return Metadata{}, err
		}
		if detail.ID == 0 {
			detail = subject
		}
		return subjectMetadata(detail), nil
	}
	return Metadata{}, nil
}

func (c Client) GetSubject(id int) (Subject, error) {
	if id <= 0 {
		return Subject{}, nil
	}
	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		base = defaultBaseURL
	}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v0/subjects/%d", base, id), nil)
	if err != nil {
		return Subject{}, fmt.Errorf("create bangumi subject request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent())

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 12 * time.Second}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return Subject{}, fmt.Errorf("fetch bangumi subject: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return Subject{}, fmt.Errorf("read bangumi subject response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Subject{}, fmt.Errorf("fetch bangumi subject: status %s: %s", resp.Status, strings.TrimSpace(string(raw)))
	}
	var subject Subject
	if err := json.Unmarshal(raw, &subject); err != nil {
		return Subject{}, fmt.Errorf("parse bangumi subject response: %w", err)
	}
	return subject, nil
}

func (c Client) SearchSubjects(keyword string, limit int) ([]Subject, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 5
	}
	body := map[string]any{
		"keyword": keyword,
		"filter": map[string]any{
			"type": []int{2},
		},
	}
	rawBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal bangumi search request: %w", err)
	}
	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		base = defaultBaseURL
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/v0/search/subjects?limit=%d&offset=0", base, limit), bytes.NewReader(rawBody))
	if err != nil {
		return nil, fmt.Errorf("create bangumi search request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent())

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 12 * time.Second}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search bangumi subject: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read bangumi search response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("search bangumi subject: status %s: %s", resp.Status, strings.TrimSpace(string(raw)))
	}
	return ParseSubjectSearchResponse(raw)
}

func (c Client) userAgent() string {
	if strings.TrimSpace(c.UserAgent) != "" {
		return c.UserAgent
	}
	return defaultUserAgent
}

func subjectMetadata(subject Subject) Metadata {
	title := subject.NameCN
	if strings.TrimSpace(title) == "" {
		title = subject.Name
	}
	return Metadata{
		ID:       subject.ID,
		Title:    title,
		CoverURL: subject.CoverURL(),
		Summary:  strings.TrimSpace(subject.Summary),
	}
}

type calendarDay struct {
	Weekday map[string]any `json:"weekday"`
	Items   []Subject      `json:"items"`
}

func (c Client) Calendar() ([]Subject, error) {
	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		base = defaultBaseURL
	}
	req, err := http.NewRequest(http.MethodGet, base+"/calendar", nil)
	if err != nil {
		return nil, fmt.Errorf("create bangumi calendar request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent())
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 12 * time.Second}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch bangumi calendar: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read bangumi calendar response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch bangumi calendar: status %s: %s", resp.Status, strings.TrimSpace(string(raw)))
	}
	var days []calendarDay
	if err := json.Unmarshal(raw, &days); err != nil {
		return nil, fmt.Errorf("parse bangumi calendar response: %w", err)
	}
	out := make([]Subject, 0)
	for _, day := range days {
		out = append(out, day.Items...)
	}
	return out, nil
}

func (c Client) Discover(section, tag string, limit int) ([]DiscoverSubject, error) {
	return c.DiscoverPage(section, tag, limit, 0)
}

func (c Client) DiscoverPage(section, tag string, limit, offset int) ([]DiscoverSubject, error) {
	if limit <= 0 || limit > 50 {
		limit = 24
	}
	if offset < 0 {
		offset = 0
	}
	section = strings.ToLower(strings.TrimSpace(section))
	tag = strings.TrimSpace(tag)
	var subjects []Subject
	var err error
	switch section {
	case "watching", "calendar", "follow":
		subjects, err = c.Calendar()
	case "browse", "category":
		if tag == "" {
			tag = "动画"
		}
		subjects, err = c.SearchSubjectsAdvanced(tag, limit, offset, "match", []string{tag})
		if err != nil {
			subjects, err = c.SearchSubjectsAdvanced(tag, limit, offset, "match", nil)
		}
	default:
		subjects, err = c.SearchSubjectsAdvanced(tag, limit, offset, "rank", nil)
		if err != nil {
			keyword := tag
			if keyword == "" {
				keyword = "动画"
			}
			subjects, err = c.SearchSubjectsAdvanced(keyword, limit, offset, "match", nil)
		}
	}
	if err != nil {
		return nil, err
	}
	if section == "watching" || section == "calendar" || section == "follow" {
		sortSubjects(subjects, func(a, b Subject) bool { return a.Collection.Doing > b.Collection.Doing })
	}
	out := make([]DiscoverSubject, 0, len(subjects))
	for _, subject := range subjects {
		if subject.Type != 0 && subject.Type != 2 {
			continue
		}
		out = append(out, subjectDiscover(subject))
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (c Client) SearchSubjectsAdvanced(keyword string, limit, offset int, sortName string, tags []string) ([]Subject, error) {
	if limit <= 0 {
		limit = 24
	}
	body := map[string]any{
		"keyword": strings.TrimSpace(keyword),
		"filter":  map[string]any{"type": []int{2}},
	}
	if strings.TrimSpace(sortName) != "" {
		body["sort"] = strings.TrimSpace(sortName)
	}
	if len(tags) > 0 {
		body["filter"].(map[string]any)["tag"] = tags
	}
	rawBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal bangumi search request: %w", err)
	}
	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		base = defaultBaseURL
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/v0/search/subjects?limit=%d&offset=%d", base, limit, offset), bytes.NewReader(rawBody))
	if err != nil {
		return nil, fmt.Errorf("create bangumi search request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent())
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 12 * time.Second}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search bangumi subject: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read bangumi search response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("search bangumi subject: status %s: %s", resp.Status, strings.TrimSpace(string(raw)))
	}
	return ParseSubjectSearchResponse(raw)
}

func sortSubjects(items []Subject, less func(a, b Subject) bool) {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if less(items[j], items[i]) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

func subjectDiscover(subject Subject) DiscoverSubject {
	title := subject.NameCN
	if strings.TrimSpace(title) == "" {
		title = subject.Name
	}
	airDate := subject.AirDate
	if airDate == "" {
		airDate = subject.Date
	}
	return DiscoverSubject{ID: subject.ID, Title: title, Name: subject.Name, CoverURL: subject.CoverURL(), Summary: strings.TrimSpace(subject.Summary), Score: subject.Rating.Score, Rank: subject.Rank, AirDate: airDate, AirWeekday: subject.AirWeekday, Doing: subject.Collection.Doing, Collection: subject.Collection.Collect}
}
