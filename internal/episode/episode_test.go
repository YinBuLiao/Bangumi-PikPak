package episode

import "testing"

func TestLabelFromTitleBracketEpisode(t *testing.T) {
	label, ok := LabelFromTitle("[桜都字幕组] 关于邻家的天使大人 第二季 [03][1080P][简体内嵌]")
	if !ok {
		t.Fatal("expected episode to be parsed")
	}
	if label != "第03集" {
		t.Fatalf("label mismatch: %q", label)
	}
}

func TestLabelFromTitleDashEpisode(t *testing.T) {
	label, ok := LabelFromTitle("[黒ネズミたち] Otonari no Tenshi-sama 2 - 03 (ABEMA 1920x1080 AVC AAC MKV)")
	if !ok {
		t.Fatal("expected episode to be parsed")
	}
	if label != "第03集" {
		t.Fatalf("label mismatch: %q", label)
	}
}

func TestLabelFromTitleChineseEpisode(t *testing.T) {
	label, ok := LabelFromTitle("某番 第12话 1080P")
	if !ok {
		t.Fatal("expected episode to be parsed")
	}
	if label != "第12集" {
		t.Fatalf("label mismatch: %q", label)
	}
}

func TestLabelFromTitleMissingEpisode(t *testing.T) {
	_, ok := LabelFromTitle("无集数标题")
	if ok {
		t.Fatal("expected no episode")
	}
}
