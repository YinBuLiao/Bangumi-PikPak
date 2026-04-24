package web

import (
	"context"
	"testing"

	"bangumi-pikpak/internal/bangumi"
	"bangumi-pikpak/internal/pikpak"
)

type fakeDrive struct {
	files map[string][]pikpak.RemoteFile
}

func (f fakeDrive) Login() error                                      { return nil }
func (f fakeDrive) List(parentID string) ([]pikpak.RemoteFile, error) { return f.files[parentID], nil }
func (f fakeDrive) DownloadURL(id string) (string, error)             { return "https://example.test/" + id, nil }

func TestBuildLibraryGroupsBangumiEpisodesAndPlayableFiles(t *testing.T) {
	drive := fakeDrive{files: map[string][]pikpak.RemoteFile{
		"root": {{ID: "b1", Name: "测试番剧", Kind: pikpak.KindFolder}},
		"b1":   {{ID: "e3", Name: "第03集", Kind: pikpak.KindFolder}},
		"e3":   {{ID: "v1", Name: "episode03.mp4", Kind: "drive#file", MimeType: "video/mp4", ThumbnailLink: "https://example.test/thumb.jpg", Size: 123}},
	}}
	lib, err := BuildLibrary(context.Background(), drive, "root", t.TempDir())
	if err != nil {
		t.Fatalf("BuildLibrary returned error: %v", err)
	}
	if len(lib.Bangumi) != 1 || lib.Bangumi[0].Title != "测试番剧" {
		t.Fatalf("bangumi mismatch: %#v", lib)
	}
	ep := lib.Bangumi[0].Episodes[0]
	if ep.Label != "第03集" || ep.Files[0].StreamURL != "/api/stream?id=v1" || lib.Bangumi[0].CoverURL != "https://example.test/thumb.jpg" {
		t.Fatalf("episode/file mismatch: %#v", lib.Bangumi[0])
	}
}

func TestBuildLibraryKeepsEmptyPikPakBangumiFolders(t *testing.T) {
	drive := fakeDrive{files: map[string][]pikpak.RemoteFile{
		"root": {{ID: "b1", Name: "还在离线中的番剧", Kind: pikpak.KindFolder}},
		"b1":   {},
	}}
	lib, err := BuildLibrary(context.Background(), drive, "root", t.TempDir())
	if err != nil {
		t.Fatalf("BuildLibrary returned error: %v", err)
	}
	if len(lib.Bangumi) != 1 || lib.Bangumi[0].Title != "还在离线中的番剧" {
		t.Fatalf("expected empty remote folder to be visible, got %#v", lib)
	}
}

type fakeCoverResolver struct{}

func (fakeCoverResolver) SearchMetadata(title string) (bangumi.Metadata, error) {
	if title == "测试番剧" {
		return bangumi.Metadata{CoverURL: "https://lain.bgm.tv/pic/cover/l/test.jpg", Summary: "Bangumi.tv 简介"}, nil
	}
	return bangumi.Metadata{}, nil
}

func TestEnrichLibraryCoversUsesBangumiTitle(t *testing.T) {
	lib := LibraryResponse{Bangumi: []Bangumi{{Title: "测试番剧"}}}
	err := EnrichLibraryCovers(context.Background(), &lib, fakeCoverResolver{}, t.TempDir())
	if err != nil {
		t.Fatalf("EnrichLibraryCovers returned error: %v", err)
	}
	if lib.Bangumi[0].CoverURL != "https://lain.bgm.tv/pic/cover/l/test.jpg" || lib.Bangumi[0].Summary != "Bangumi.tv 简介" {
		t.Fatalf("cover mismatch: %#v", lib.Bangumi[0])
	}
}
