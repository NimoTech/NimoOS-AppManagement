package service

import (
	"testing"

	"github.com/NimoTech/NimoOS-AppManagement/model"
)

func TestApplyDesktopMeta(t *testing.T) {
	t.Run("enable 容器填充元数据", func(t *testing.T) {
		app := model.MyAppList{Name: "my-container", ID: "abc123"}
		ApplyDesktopMeta(&app, map[string]string{
			"nimoos.enable": "true", "nimoos.title": "下载器",
			"nimoos.icon": "/icon.png", "nimoos.port": "8080",
			"nimoos.widget.path": "/widget", "nimoos.widget.w": "3", "nimoos.widget.h": "2",
			"nimoos.widget.minw": "3", "nimoos.widget.maxw": "4",
		})
		if !app.Desktop || app.DesktopTitle != "下载器" || app.Icon != "/icon.png" ||
			app.Port != "8080" || app.Index != "/" || app.Protocol != "http" ||
			app.WidgetPath != "/widget" || app.WidgetW != 3 || app.WidgetH != 2 ||
			app.WidgetMinW != 3 || app.WidgetMaxW != 4 || app.WidgetMinH != 0 || app.WidgetMaxH != 0 {
			t.Fatalf("bad apply: %+v", app)
		}
		if app.Name != "my-container" {
			t.Fatal("Name(容器名,前端 key)不得被 title 覆盖")
		}
	})

	t.Run("无 enable 一字不动", func(t *testing.T) {
		app := model.MyAppList{Name: "n", ID: "i", Icon: ""}
		before := app
		ApplyDesktopMeta(&app, map[string]string{"nimoos.title": "x"})
		if app != before {
			t.Fatalf("must be untouched: %+v", app)
		}
	})
}

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

	t.Run("自定义尺寸范围四 label", func(t *testing.T) {
		m := ParseDesktopLabels(map[string]string{
			"nimoos.enable":      "true",
			"nimoos.widget.path": "/widget",
			"nimoos.widget.minw": "3", "nimoos.widget.minh": "2",
			"nimoos.widget.maxw": "4", "nimoos.widget.maxh": "3",
		})
		if m.WidgetMinW != 3 || m.WidgetMinH != 2 || m.WidgetMaxW != 4 || m.WidgetMaxH != 3 {
			t.Fatalf("bad range: %+v", m)
		}
	})

	t.Run("范围 label 缺省为 0(不带字段)", func(t *testing.T) {
		m := ParseDesktopLabels(map[string]string{
			"nimoos.enable": "true", "nimoos.widget.path": "/widget",
		})
		if m.WidgetMinW != 0 || m.WidgetMinH != 0 || m.WidgetMaxW != 0 || m.WidgetMaxH != 0 {
			t.Fatalf("expected all-zero range: %+v", m)
		}
	})

	t.Run("resize=false 糖:min=max=声明的 w/h", func(t *testing.T) {
		m := ParseDesktopLabels(map[string]string{
			"nimoos.enable": "true", "nimoos.widget.path": "/widget",
			"nimoos.widget.w": "4", "nimoos.widget.h": "3",
			"nimoos.widget.resize": "false",
		})
		if m.WidgetMinW != 4 || m.WidgetMaxW != 4 || m.WidgetMinH != 3 || m.WidgetMaxH != 3 {
			t.Fatalf("bad sugar: %+v", m)
		}
	})

	t.Run("resize=false 糖:w/h 未声明按默认 2×2 锁死", func(t *testing.T) {
		m := ParseDesktopLabels(map[string]string{
			"nimoos.enable": "true", "nimoos.widget.path": "/widget",
			"nimoos.widget.resize": "false",
		})
		if m.WidgetMinW != 2 || m.WidgetMaxW != 2 || m.WidgetMinH != 2 || m.WidgetMaxH != 2 {
			t.Fatalf("bad sugar default: %+v", m)
		}
	})

	t.Run("显式 min/max label 优先于 resize=false", func(t *testing.T) {
		m := ParseDesktopLabels(map[string]string{
			"nimoos.enable": "true", "nimoos.widget.path": "/widget",
			"nimoos.widget.w": "3",
			"nimoos.widget.minw":   "2",
			"nimoos.widget.resize": "false",
		})
		// minw 显式给 2,其余未给的按糖补:maxw=w=3,minh=maxh=2
		if m.WidgetMinW != 2 || m.WidgetMaxW != 3 || m.WidgetMinH != 2 || m.WidgetMaxH != 2 {
			t.Fatalf("explicit label must win: %+v", m)
		}
	})
}
