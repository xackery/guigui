// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/hajimehoshi/guigui"
)

type ScrollablePanel struct {
	guigui.DefaultWidgetBehavior

	content            *guigui.Widget
	scollOverlayWidget *guigui.Widget
	border             *guigui.Widget

	contentWidth  int
	contentHeight int
}

func (s *ScrollablePanel) SetContent(content *guigui.Widget) {
	s.content = content
}

func (s *ScrollablePanel) SetContentSize(width, height int) {
	s.contentWidth = width
	s.contentHeight = height
}

func (s *ScrollablePanel) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	if s.scollOverlayWidget == nil {
		var so ScrollOverlay
		s.scollOverlayWidget = guigui.NewWidget(&so)
	}

	if s.content != nil {
		offsetX, offsetY := s.scollOverlayWidget.Behavior().(*ScrollOverlay).Offset()
		b := widget.Bounds()
		b.Max.X = max(b.Max.X, b.Min.X+s.contentWidth)
		b.Max.Y = max(b.Max.Y, b.Min.Y+s.contentHeight)
		b = b.Add(image.Pt(int(offsetX), int(offsetY)))
		appender.AppendChildWidget(s.content, b)
	}

	appender.AppendChildWidget(s.scollOverlayWidget, widget.Bounds())

	if s.border == nil {
		b := scrollablePanelBorder{
			scrollOverlay: s.scollOverlayWidget.Behavior().(*ScrollOverlay),
		}
		s.border = guigui.NewWidget(&b)
	}
	appender.AppendChildWidget(s.border, widget.Bounds())
}

func (s *ScrollablePanel) Update(context *guigui.Context, widget *guigui.Widget) error {
	so := s.scollOverlayWidget.Behavior().(*ScrollOverlay)
	so.SetContentSize(s.contentWidth, s.contentHeight)

	return nil
}

type scrollablePanelBorder struct {
	guigui.DefaultWidgetBehavior

	scrollOverlay *ScrollOverlay
}

func (s *scrollablePanelBorder) Draw(context *guigui.Context, widget *guigui.Widget, dst *ebiten.Image) {
	// Render borders.
	strokeWidth := float32(1 * context.Scale())
	x0 := float32(widget.Bounds().Min.X)
	x1 := float32(widget.Bounds().Max.X)
	y0 := float32(widget.Bounds().Min.Y)
	y1 := float32(widget.Bounds().Max.Y)
	offsetX, offsetY := s.scrollOverlay.Offset()
	if offsetX < 0 {
		vector.StrokeLine(dst, x0+strokeWidth/2, y0, x0+strokeWidth/2, y1, strokeWidth, Color(context.ColorMode(), ColorTypeBase, 0.85), false)
	}
	if offsetY < 0 {
		vector.StrokeLine(dst, x0, y0+strokeWidth/2, x1, y0+strokeWidth/2, strokeWidth, Color(context.ColorMode(), ColorTypeBase, 0.85), false)
	}
}
