// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/guigui"
)

type Sidebar struct {
	guigui.DefaultWidgetBehavior

	scrollablePanelWidget *guigui.Widget

	widthMinusDefault  int
	heightMinusDefault int
}

func (s *Sidebar) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	if s.scrollablePanelWidget == nil {
		s.scrollablePanelWidget = guigui.NewWidget(&ScrollablePanel{})
	}
	w, h := s.Size(context, widget)
	s.scrollablePanelWidget.Behavior().(*ScrollablePanel).SetSize(context, w, h)
	appender.AppendChildWidget(s.scrollablePanelWidget, widget.Position())
}

func (s *Sidebar) SetContent(context *guigui.Context, f func(context *guigui.Context, widget *guigui.Widget, childAppender *ScrollablePanelChildWidgetAppender)) {
	if s.scrollablePanelWidget == nil {
		s.scrollablePanelWidget = guigui.NewWidget(&ScrollablePanel{})
	}
	s.scrollablePanelWidget.Behavior().(*ScrollablePanel).SetContent(f)
}

func (s *Sidebar) Draw(context *guigui.Context, widget *guigui.Widget, dst *ebiten.Image) {
	dst.Fill(Color(context.ColorMode(), ColorTypeBase, 0.875))
	b := s.bounds(context, widget)
	b.Min.X = b.Max.X - int(1*context.Scale())
	dst.SubImage(b).(*ebiten.Image).Fill(Color(context.ColorMode(), ColorTypeBase, 0.85))
}

func defaultSidebarWidth(context *guigui.Context) (int, int) {
	return 6 * UnitSize(context), 6 * UnitSize(context)
}

func (s *Sidebar) Size(context *guigui.Context, widget *guigui.Widget) (int, int) {
	dw, dh := defaultSidebarWidth(context)
	return s.widthMinusDefault + dw, s.heightMinusDefault + dh
}

func (s *Sidebar) SetSize(context *guigui.Context, width, height int) {
	dw, dh := defaultSidebarWidth(context)
	s.widthMinusDefault = width - dw
	s.heightMinusDefault = height - dh
}

func (s *Sidebar) bounds(context *guigui.Context, widget *guigui.Widget) image.Rectangle {
	p := widget.Position()
	w, h := s.Size(context, widget)
	return image.Rectangle{
		Min: p,
		Max: p.Add(image.Pt(w, h)),
	}
}
