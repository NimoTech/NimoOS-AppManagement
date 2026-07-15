package service

import (
	"strconv"

	"github.com/NimoTech/NimoOS-AppManagement/model"
)

// DesktopLabelMeta 是 nimoos.* 桌面接入 label 的解析结果。
// 规范见 NimoOS-UI/docs/superpowers/specs/2026-07-15-desktop-app-recognition-design.md §3。
type DesktopLabelMeta struct {
	Title      string
	Icon       string
	Scheme     string
	Port       string
	Index      string
	WidgetPath string
	WidgetW    int // 0 = 未声明或非法,前端夹紧
	WidgetH    int
}

// ParseDesktopLabels 只在 labels["nimoos.enable"] == "true" 时返回非 nil。
func ParseDesktopLabels(labels map[string]string) *DesktopLabelMeta {
	if labels["nimoos.enable"] != "true" {
		return nil
	}
	m := &DesktopLabelMeta{
		Title:      labels["nimoos.title"],
		Icon:       labels["nimoos.icon"],
		Scheme:     labels["nimoos.scheme"],
		Port:       labels["nimoos.port"],
		Index:      labels["nimoos.index"],
		WidgetPath: labels["nimoos.widget.path"],
	}
	if m.Scheme == "" {
		m.Scheme = "http"
	}
	if m.Index == "" {
		m.Index = "/"
	}
	m.WidgetW, _ = strconv.Atoi(labels["nimoos.widget.w"])
	m.WidgetH, _ = strconv.Atoi(labels["nimoos.widget.h"])
	return m
}

// ApplyDesktopMeta 把 nimoos.* label 应用到 MyAppList;无 nimoos.enable=true 则不动。
// 注意:不覆盖 app.Name(容器名是前端桌面的稳定 key),标题放 DesktopTitle。
func ApplyDesktopMeta(app *model.MyAppList, labels map[string]string) {
	dm := ParseDesktopLabels(labels)
	if dm == nil {
		return
	}
	app.Desktop = true
	app.DesktopTitle = dm.Title
	app.Icon = dm.Icon
	app.Port = dm.Port
	app.Index = dm.Index
	app.Protocol = dm.Scheme
	app.WidgetPath = dm.WidgetPath
	app.WidgetW = dm.WidgetW
	app.WidgetH = dm.WidgetH
}
