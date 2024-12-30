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

func (w *widgetWithBounds) bounds(context *guigui.Context) image.Rectangle {
	return image.Rectangle{
		Min: w.position,
		Max: w.position.Add(image.Pt(w.widget.Size(context))),
	}
}

type ScrollablePanel struct {
	guigui.DefaultWidgetBehavior

	setContentFunc     func(context *guigui.Context, widget *guigui.Widget, childAppender *ScrollablePanelChildWidgetAppender)
	childWidgets       []widgetWithBounds
	scollOverlayWidget *guigui.Widget
	border             *guigui.Widget

	paddingX           int
	paddingY           int
	widthMinusDefault  int
	heightMinusDefault int
}

type ScrollablePanelChildWidgetAppender struct {
	context         *guigui.Context
	scrollablePanel *ScrollablePanel
}

func (s *ScrollablePanelChildWidgetAppender) AppendChildWidget(widget *guigui.Widget, position image.Point) {
	s.scrollablePanel.childWidgets = append(s.scrollablePanel.childWidgets, widgetWithBounds{
		widget:   widget,
		position: position,
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
			context:         context,
			scrollablePanel: s,
		})
	}

	if s.scollOverlayWidget == nil {
		var so ScrollOverlay
		s.scollOverlayWidget = guigui.NewWidget(&so)
	}

	offsetX, offsetY := s.scollOverlayWidget.Behavior().(*ScrollOverlay).Offset()
	for _, childWidget := range s.childWidgets {
		p := childWidget.position
		p = p.Add(image.Pt(int(offsetX), int(offsetY)))
		appender.AppendChildWidget(childWidget.widget, p)
	}

	appender.AppendChildWidget(s.scollOverlayWidget, widget.Position())

	if s.border == nil {
		b := scrollablePanelBorder{
			scrollOverlay: s.scollOverlayWidget.Behavior().(*ScrollOverlay),
		}
		s.border = guigui.NewWidget(&b)
	}
	appender.AppendChildWidget(s.border, widget.Position())
}

func (s *ScrollablePanel) Update(context *guigui.Context, widget *guigui.Widget) error {
	p := widget.Position()
	var w, h int
	for _, childWidget := range s.childWidgets {
		b := childWidget.bounds(context)
		w = max(w, b.Max.X-p.X+s.paddingX)
		h = max(h, b.Max.Y-p.Y+s.paddingY)
	}
	so := s.scollOverlayWidget.Behavior().(*ScrollOverlay)
	so.SetContentSize(w, h)
	return nil
}

func defaultScrollablePanelSize(context *guigui.Context) (int, int) {
	return 6 * UnitSize(context), 6 * UnitSize(context)
}

func (s *ScrollablePanel) Size(context *guigui.Context, widget *guigui.Widget) (int, int) {
	dw, dh := defaultScrollablePanelSize(context)
	return s.widthMinusDefault + dw, s.heightMinusDefault + dh
}

func (s *ScrollablePanel) SetSize(context *guigui.Context, width, height int) {
	dw, dh := defaultScrollablePanelSize(context)
	s.widthMinusDefault = width - dw
	s.heightMinusDefault = height - dh
}

type scrollablePanelBorder struct {
	guigui.DefaultWidgetBehavior

	scrollOverlay *ScrollOverlay
}

func (s *scrollablePanelBorder) Draw(context *guigui.Context, widget *guigui.Widget, dst *ebiten.Image) {
	// Render borders.
	strokeWidth := float32(1 * context.Scale())
	bounds := s.bounds(context, widget)
	x0 := float32(bounds.Min.X)
	x1 := float32(bounds.Max.X)
	y0 := float32(bounds.Min.Y)
	y1 := float32(bounds.Max.Y)
	offsetX, offsetY := s.scrollOverlay.Offset()
	if offsetX < 0 {
		vector.StrokeLine(dst, x0+strokeWidth/2, y0, x0+strokeWidth/2, y1, strokeWidth, Color(context.ColorMode(), ColorTypeBase, 0.85), false)
	}
	if offsetY < 0 {
		vector.StrokeLine(dst, x0, y0+strokeWidth/2, x1, y0+strokeWidth/2, strokeWidth, Color(context.ColorMode(), ColorTypeBase, 0.85), false)
	}
}

func (s *scrollablePanelBorder) Size(context *guigui.Context, widget *guigui.Widget) (int, int) {
	return widget.Parent().Size(context)
}

func (s *scrollablePanelBorder) bounds(context *guigui.Context, widget *guigui.Widget) image.Rectangle {
	p := widget.Position()
	w, h := s.Size(context, widget)
	return image.Rectangle{
		Min: p,
		Max: p.Add(image.Pt(w, h)),
	}
}
