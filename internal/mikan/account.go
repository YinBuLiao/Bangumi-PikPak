package mikan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type BangumiCandidate struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	CoverURL string `json:"cover_url,omitempty"`
	PageURL  string `json:"page_url"`
}

func NewSession(base *http.Client) *http.Client {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}
	if base != nil {
		client.Transport = base.Transport
		client.Timeout = base.Timeout
		client.CheckRedirect = base.CheckRedirect
	}
	return client
}

func Login(client *http.Client, username, password string) error {
	username = strings.TrimSpace(username)
	if username == "" || strings.TrimSpace(password) == "" {
		return fmt.Errorf("mikan username/password is required")
	}
	loginURL := defaultBaseURL + "/Account/Login"
	resp, err := client.Get(loginURL)
	if err != nil {
		return fmt.Errorf("fetch mikan login page: %w", err)
	}
	token, err := parseRequestVerificationToken(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return err
	}
	form := url.Values{}
	form.Set("UserName", username)
	form.Set("Password", password)
	form.Set("RememberMe", "true")
	form.Set("__RequestVerificationToken", token)
	req, err := http.NewRequest(http.MethodPost, loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", loginURL)
	req.Header.Set("User-Agent", "Bangumi-PikPak/1.0")
	resp, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("post mikan login: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("mikan login: status %s", resp.Status)
	}
	body := string(raw)
	if strings.Contains(body, `name="Password"`) && strings.Contains(body, `name="UserName"`) {
		return fmt.Errorf("mikan login failed: check username/password")
	}
	return nil
}

func SubscribeBangumi(client *http.Client, bangumiID, language int) error {
	if bangumiID <= 0 {
		return fmt.Errorf("mikan bangumi id is required")
	}
	if language < 0 || language > 2 {
		language = 0
	}
	payload := map[string]any{"BangumiID": bangumiID, "SubtitleGroupID": nil, "Language": language}
	raw, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, defaultBaseURL+"/Home/SubscribeBangumi", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", fmt.Sprintf("%s/Home/Bangumi/%d", defaultBaseURL, bangumiID))
	req.Header.Set("User-Agent", "Bangumi-PikPak/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("subscribe mikan bangumi: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("subscribe mikan bangumi: status %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func FindBangumiCandidate(client *http.Client, title string) (BangumiCandidate, error) {
	results, err := Search(client, title, 8)
	if err != nil {
		return BangumiCandidate{}, err
	}
	for _, result := range results {
		candidate, err := FetchBangumiCandidateFromEpisode(client, result.Link)
		if err == nil && candidate.ID > 0 {
			return candidate, nil
		}
	}
	return BangumiCandidate{}, fmt.Errorf("mikan bangumi not found for %q", title)
}

func FetchBangumiCandidateFromEpisode(client *http.Client, pageURL string) (BangumiCandidate, error) {
	resp, err := client.Get(pageURL)
	if err != nil {
		return BangumiCandidate{}, fmt.Errorf("fetch mikan episode page: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return BangumiCandidate{}, fmt.Errorf("fetch mikan episode page: status %s", resp.Status)
	}
	return ParseBangumiCandidate(resp.Body, defaultBaseURL)
}

func FetchBangumiSubjectID(client *http.Client, pageURL string) (int, error) {
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Get(pageURL)
	if err != nil {
		return 0, fmt.Errorf("fetch mikan bangumi page: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("fetch mikan bangumi page: status %s", resp.Status)
	}
	return ParseBangumiSubjectID(resp.Body)
}

func ParseBangumiSubjectID(r io.Reader) (int, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return 0, fmt.Errorf("parse mikan bangumi page: %w", err)
	}
	subjectID := 0
	doc.Find(`a[href*="bgm.tv/subject/"], a[href*="bangumi.tv/subject/"]`).EachWithBreak(func(_ int, a *goquery.Selection) bool {
		href, ok := a.Attr("href")
		if !ok {
			return true
		}
		id := bangumiSubjectID(href)
		if id <= 0 {
			return true
		}
		subjectID = id
		return false
	})
	if subjectID <= 0 {
		return 0, fmt.Errorf("bangumi.tv subject link not found")
	}
	return subjectID, nil
}

func ParseBangumiCandidate(r io.Reader, base string) (BangumiCandidate, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return BangumiCandidate{}, fmt.Errorf("parse mikan episode page: %w", err)
	}
	var out BangumiCandidate
	doc.Find(`a[href*="/Home/Bangumi/"]`).EachWithBreak(func(_ int, a *goquery.Selection) bool {
		href, ok := a.Attr("href")
		if !ok {
			return true
		}
		id := mikanBangumiID(href)
		if id <= 0 {
			return true
		}
		out.ID = id
		out.Title = strings.TrimSpace(a.Text())
		out.PageURL = absoluteURL(base, href)
		return false
	})
	if out.ID <= 0 {
		return out, fmt.Errorf("mikan bangumi link not found")
	}
	if out.Title == "" {
		out.Title = strings.TrimSpace(doc.Find(titleSelector).First().Text())
	}
	if style, ok := doc.Find(".bangumi-poster").First().Attr("style"); ok {
		out.CoverURL = absoluteURL(base, backgroundImageURL(style))
	}
	return out, nil
}

func parseRequestVerificationToken(r io.Reader) (string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", fmt.Errorf("parse mikan login page: %w", err)
	}
	token, _ := doc.Find(`input[name="__RequestVerificationToken"]`).First().Attr("value")
	if strings.TrimSpace(token) == "" {
		return "", fmt.Errorf("mikan login token not found")
	}
	return token, nil
}

func mikanBangumiID(href string) int {
	re := regexp.MustCompile(`/Home/Bangumi/(\d+)`)
	m := re.FindStringSubmatch(href)
	if len(m) != 2 {
		return 0
	}
	id, _ := strconv.Atoi(m[1])
	return id
}

func bangumiSubjectID(href string) int {
	re := regexp.MustCompile(`(?:bgm\.tv|bangumi\.tv)/subject/(\d+)`)
	m := re.FindStringSubmatch(href)
	if len(m) != 2 {
		return 0
	}
	id, _ := strconv.Atoi(m[1])
	return id
}
