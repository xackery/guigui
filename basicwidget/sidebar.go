// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/guigui"
)

type Sidebar struct {
	guigui.DefaultWidgetBehavior

	scrollablePanelWidget *guigui.Widget
}

func (s *Sidebar) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	if s.scrollablePanelWidget == nil {
		var sp ScrollablePanel
		s.scrollablePanelWidget = guigui.NewWidget(&sp)
	}
	appender.AppendChildWidgetWithBounds(s.scrollablePanelWidget, widget.Bounds())
}

func (s *Sidebar) SetContent(context *guigui.Context, f func(context *guigui.Context, widget *guigui.Widget, childAppender *ScrollablePanelChildWidgetAppender)) {
	s.scrollablePanelWidget.Behavior().(*ScrollablePanel).SetContent(f)
}

func (s *Sidebar) Draw(context *guigui.Context, widget *guigui.Widget, dst *ebiten.Image) {
	dst.Fill(Color(context.ColorMode(), ColorTypeBase, 0.875))
	b := widget.Bounds()
	b.Min.X = b.Max.X - int(1*context.Scale())
	dst.SubImage(b).(*ebiten.Image).Fill(Color(context.ColorMode(), ColorTypeBase, 0.85))
}
