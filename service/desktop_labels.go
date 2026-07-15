package service

import "strconv"

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
