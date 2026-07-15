package service

import "testing"

func TestParseDesktopLabels(t *testing.T) {
	t.Run("无 enable 返回 nil", func(t *testing.T) {
		if ParseDesktopLabels(map[string]string{"nimoos.title": "x"}) != nil {
			t.Fatal("expected nil without nimoos.enable")
		}
		if ParseDesktopLabels(map[string]string{"nimoos.enable": "false"}) != nil {
			t.Fatal("expected nil with enable=false")
		}
		if ParseDesktopLabels(nil) != nil {
			t.Fatal("expected nil with nil labels")
		}
	})

	t.Run("全套解析", func(t *testing.T) {
		m := ParseDesktopLabels(map[string]string{
			"nimoos.enable": "true", "nimoos.title": "我的下载器",
			"nimoos.icon": "/icon.png", "nimoos.scheme": "https",
			"nimoos.port": "8080", "nimoos.index": "/ui",
			"nimoos.widget.path": "/widget", "nimoos.widget.w": "4", "nimoos.widget.h": "3",
		})
		if m == nil {
			t.Fatal("expected meta")
		}
		if m.Title != "我的下载器" || m.Icon != "/icon.png" || m.Scheme != "https" ||
			m.Port != "8080" || m.Index != "/ui" || m.WidgetPath != "/widget" ||
			m.WidgetW != 4 || m.WidgetH != 3 {
			t.Fatalf("bad parse: %+v", m)
		}
	})

	t.Run("缺省值", func(t *testing.T) {
		m := ParseDesktopLabels(map[string]string{"nimoos.enable": "true"})
		if m == nil {
			t.Fatal("expected meta")
		}
		if m.Scheme != "http" || m.Index != "/" {
			t.Fatalf("bad defaults: %+v", m)
		}
		if m.WidgetPath != "" || m.WidgetW != 0 || m.WidgetH != 0 {
			t.Fatalf("widget fields should be zero: %+v", m)
		}
	})

	t.Run("非法 w/h 置 0 照传", func(t *testing.T) {
		m := ParseDesktopLabels(map[string]string{
			"nimoos.enable": "true", "nimoos.widget.path": "/w",
			"nimoos.widget.w": "abc", "nimoos.widget.h": "-1",
		})
		if m.WidgetW != 0 {
			t.Fatalf("invalid w should be 0, got %d", m.WidgetW)
		}
		if m.WidgetH != -1 {
			t.Fatalf("h=-1 parses to -1 (前端夹紧), got %d", m.WidgetH)
		}
	})
}
