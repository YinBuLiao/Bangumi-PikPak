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

func TestParseEpisodeMetadataExtractsTitleAndCover(t *testing.T) {
	html := `<div class="bangumi-poster" style="background-image: url('/images/Bangumi/202604/d9305f09.jpg?width=400&height=560&format=webp');"></div>
<p class="bangumi-title"><a href="/Home/Bangumi/3903">关于邻家的天使大人 第二季</a></p>`
	meta, err := ParseEpisodeMetadata(strings.NewReader(html), "https://mikanani.me")
	if err != nil {
		t.Fatalf("ParseEpisodeMetadata returned error: %v", err)
	}
	if meta.Title != "关于邻家的天使大人 第二季" {
		t.Fatalf("title mismatch: %q", meta.Title)
	}
	if meta.CoverURL != "https://mikanani.me/images/Bangumi/202604/d9305f09.jpg?width=400&height=560&format=webp" {
		t.Fatalf("cover mismatch: %q", meta.CoverURL)
	}
}

func TestParseSearchResultsExtractsTorrentRows(t *testing.T) {
	html := `<table><tbody>
<tr class="js-search-results-row">
  <td><input class="js-episode-select" data-magnet="magnet:?xt=urn:btih:abc"/></td>
  <td><a href="/Home/Episode/abc" class="magnet-link-wrap">[字幕组] 测试番剧 - 03 [1080P]</a></td>
  <td>407 MB</td>
  <td>2026/04/23 12:17</td>
  <td><a href="/Download/20260423/abc.torrent">下载</a></td>
</tr>
</tbody></table>`
	results, err := ParseSearchResults(strings.NewReader(html), "https://mikanani.me")
	if err != nil {
		t.Fatalf("ParseSearchResults returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected one result, got %d", len(results))
	}
	got := results[0]
	if got.Title != "[字幕组] 测试番剧 - 03 [1080P]" || got.Link != "https://mikanani.me/Home/Episode/abc" {
		t.Fatalf("result mismatch: %#v", got)
	}
	if got.TorrentURL != "https://mikanani.me/Download/20260423/abc.torrent" || got.EpisodeLabel != "第03集" {
		t.Fatalf("torrent/episode mismatch: %#v", got)
	}
}

func TestParseSearchResultsReturnsEmptySliceWhenNoRows(t *testing.T) {
	results, err := ParseSearchResults(strings.NewReader(`<html><body>没有找到结果</body></html>`), "https://mikanani.me")
	if err != nil {
		t.Fatalf("ParseSearchResults returned error: %v", err)
	}
	if results == nil || len(results) != 0 {
		t.Fatalf("expected non-nil empty slice, got %#v", results)
	}
}

func TestParseSchedule(t *testing.T) {
	html := `<div class="sk-bangumi" data-dayofweek="5">
		<div id="data-row-5" class="row">星期五</div>
		<div class="an-box"><ul>
			<li>
				<span data-src="/images/Bangumi/202604/d9305f09.jpg?width=400&height=400&format=webp" class="js-expand_bangumi b-lazy" data-bangumiid="3903"></span>
				<div class="an-info"><div class="an-info-group">
					<div class="date-text">2026/04/23 更新</div>
					<a href="/Home/Bangumi/3903" class="an-text" title="测试新番 第二季">测试新番 第二季</a>
				</div></div>
			</li>
		</ul></div>
	</div>`
	got, err := ParseSchedule(strings.NewReader(html), "https://mikanani.me")
	if err != nil {
		t.Fatalf("ParseSchedule returned error: %v", err)
	}
	if len(got.Days) != 1 || len(got.Items) != 1 {
		t.Fatalf("schedule size mismatch: %#v", got)
	}
	item := got.Items[0]
	if item.ID != 3903 || item.Title != "测试新番 第二季" || item.Weekday != 5 || item.DayLabel != "星期五" || item.PageURL != "https://mikanani.me/Home/Bangumi/3903" {
		t.Fatalf("item mismatch: %#v", item)
	}
	if item.CoverURL == "" || item.Updated != "2026/04/23 更新" {
		t.Fatalf("metadata mismatch: %#v", item)
	}
}
