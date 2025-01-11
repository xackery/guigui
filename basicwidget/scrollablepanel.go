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

type widgetWithPosition struct {
	widget   guigui.Widget
	position image.Point
}

func (w *widgetWithPosition) bounds(context *guigui.Context) image.Rectangle {
	return image.Rectangle{
		Min: w.position,
		Max: w.position.Add(image.Pt(w.widget.Size(context))),
	}
}

type ScrollablePanel struct {
	guigui.DefaultWidget

	setContentFunc func(context *guigui.Context, childAppender *ScrollablePanelChildWidgetAppender)
	childWidgets   []widgetWithPosition
	scollOverlay   ScrollOverlay
	border         scrollablePanelBorder

	paddingX           int
	paddingY           int
	widthMinusDefault  int
	heightMinusDefault int
}

type ScrollablePanelChildWidgetAppender struct {
	context         *guigui.Context
	scrollablePanel *ScrollablePanel
}

func (s *ScrollablePanelChildWidgetAppender) AppendChildWidget(widget guigui.Widget, position image.Point) {
	s.scrollablePanel.childWidgets = append(s.scrollablePanel.childWidgets, widgetWithPosition{
		widget:   widget,
		position: position,
	})
}

func (s *ScrollablePanel) SetContent(f func(context *guigui.Context, childAppender *ScrollablePanelChildWidgetAppender)) {
	s.setContentFunc = f
}

func (s *ScrollablePanel) SetPadding(paddingX, paddingY int) {
	s.paddingX = paddingX
	s.paddingY = paddingY
}

func (s *ScrollablePanel) AppendChildWidgets(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	s.childWidgets = slices.Delete(s.childWidgets, 0, len(s.childWidgets))
	if s.setContentFunc != nil {
		s.setContentFunc(context, &ScrollablePanelChildWidgetAppender{
			context:         context,
			scrollablePanel: s,
		})
	}

	offsetX, offsetY := s.scollOverlay.Offset()
	for _, childWidget := range s.childWidgets {
		p := childWidget.position
		p = p.Add(image.Pt(int(offsetX), int(offsetY)))
		appender.AppendChildWidget(childWidget.widget, p)
	}

	appender.AppendChildWidget(&s.scollOverlay, guigui.Position(s))

	s.border.scrollOverlay = &s.scollOverlay
	appender.AppendChildWidget(&s.border, guigui.Position(s))
}

func (s *ScrollablePanel) Update(context *guigui.Context) error {
	p := guigui.Position(s)
	var w, h int
	for _, childWidget := range s.childWidgets {
		b := childWidget.bounds(context)
		w = max(w, b.Max.X-p.X+s.paddingX)
		h = max(h, b.Max.Y-p.Y+s.paddingY)
	}
	s.scollOverlay.SetContentSize(w, h)
	return nil
}

func defaultScrollablePanelSize(context *guigui.Context) (int, int) {
	return 6 * UnitSize(context), 6 * UnitSize(context)
}

func (s *ScrollablePanel) Size(context *guigui.Context) (int, int) {
	dw, dh := defaultScrollablePanelSize(context)
	return s.widthMinusDefault + dw, s.heightMinusDefault + dh
}

func (s *ScrollablePanel) SetSize(context *guigui.Context, width, height int) {
	dw, dh := defaultScrollablePanelSize(context)
	s.widthMinusDefault = width - dw
	s.heightMinusDefault = height - dh
}

type scrollablePanelBorder struct {
	guigui.DefaultWidget

	scrollOverlay *ScrollOverlay
}

func (s *scrollablePanelBorder) Draw(context *guigui.Context, dst *ebiten.Image) {
	// Render borders.
	strokeWidth := float32(1 * context.Scale())
	bounds := guigui.Bounds(s)
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

func (s *scrollablePanelBorder) Size(context *guigui.Context) (int, int) {
	return guigui.Parent(s).Size(context)
}
