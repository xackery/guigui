// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"image"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/hajimehoshi/guigui"
)

type widgetWithBounds struct {
	widget   *guigui.Widget
	position image.Point
}

type ScrollablePanel struct {
	guigui.DefaultWidgetBehavior

	setContentFunc     func(context *guigui.Context, widget *guigui.Widget, childAppender *ScrollablePanelChildWidgetAppender)
	childWidgets       []widgetWithBounds
	scollOverlayWidget *guigui.Widget
	border             *guigui.Widget

	paddingX int
	paddingY int
}

type ScrollablePanelChildWidgetAppender struct {
	scrollablePanel *ScrollablePanel
}

func (s *ScrollablePanelChildWidgetAppender) AppendChildWidget(widget *guigui.Widget, x, y int) {
	s.scrollablePanel.childWidgets = append(s.scrollablePanel.childWidgets, widgetWithBounds{
		widget:   widget,
		position: image.Pt(x, y),
	})
}

func (s *ScrollablePanel) SetContent(f func(context *guigui.Context, widget *guigui.Widget, childAppender *ScrollablePanelChildWidgetAppender)) {
	s.setContentFunc = f

}

func (s *ScrollablePanel) SetPadding(paddingX, paddingY int) {
	s.paddingX = paddingX
	s.paddingY = paddingY
}

func (s *ScrollablePanel) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	s.childWidgets = slices.Delete(s.childWidgets, 0, len(s.childWidgets))
	if s.setContentFunc != nil {
		s.setContentFunc(context, widget, &ScrollablePanelChildWidgetAppender{
			scrollablePanel: s,
		})
	}

	if s.scollOverlayWidget == nil {
		var so ScrollOverlay
		s.scollOverlayWidget = guigui.NewWidget(&so)
	}

	offsetX, offsetY := s.scollOverlayWidget.Behavior().(*ScrollOverlay).Offset()
	for _, childWidget := range s.childWidgets {
		x, y := childWidget.position.X, childWidget.position.Y
		w, h := childWidget.widget.Size(context)
		b := image.Rect(x, y, x+w, y+h)
		b = b.Add(image.Pt(int(offsetX), int(offsetY)))
		appender.AppendChildWidget(childWidget.widget, b)
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
	b := widget.Bounds()
	var w, h int
	for _, childWidget := range s.childWidgets {
		cx, cy := childWidget.position.X, childWidget.position.Y
		cw, ch := childWidget.widget.Size(context)
		w = max(w, cx+cw-b.Min.X+s.paddingX)
		h = max(h, cy+ch-b.Min.Y+s.paddingY)
	}
	w = max(w, b.Dx())
	h = max(h, b.Dy())
	so := s.scollOverlayWidget.Behavior().(*ScrollOverlay)
	so.SetContentSize(w, h)
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
