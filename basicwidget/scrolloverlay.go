// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"image"
	"image/color"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/guigui"
)

const (
	barMaxOpacity  = 10
	barShowingTime = 60
)

type ScrollOverlay struct {
	guigui.DefaultWidgetBehavior

	contentWidth  int
	contentHeight int
	offsetX       float64
	offsetY       float64

	hovering             bool
	lastCursorX          int
	lastCursorY          int
	lastWheelX           float64
	lastWheelY           float64
	lastOffsetX          float64
	lastOffsetY          float64
	draggingX            bool
	draggingY            bool
	draggingStartX       int
	draggingStartY       int
	draggingStartOffsetX float64
	draggingStartOffsetY float64

	barOpacity     int
	barVisibleTime int

	needsAdjustOffset bool
	needsRedraw       bool
}

type ScrollEvent struct {
	OffsetX float64
	OffsetY float64
}

func (s *ScrollOverlay) Reset() {
	s.offsetX = 0
	s.offsetY = 0
}

func (s *ScrollOverlay) SetContentSize(contentWidth, contentHeight int) {
	if s.contentWidth == contentWidth && s.contentHeight == contentHeight {
		return
	}

	s.contentWidth = contentWidth
	s.contentHeight = contentHeight
	s.needsAdjustOffset = true
	s.needsRedraw = true
}

func (s *ScrollOverlay) SetOffsetByDelta(dx, dy float64) {
	s.SetOffset(s.offsetX+dx, s.offsetY+dy)
}

func (s *ScrollOverlay) SetOffset(x, y float64) {
	if s.offsetX == x && s.offsetY == y {
		return
	}
	s.offsetX = x
	s.offsetY = y
	s.needsAdjustOffset = true
	s.needsRedraw = true
}

func (s *ScrollOverlay) setHovering(hovering bool) {
	s.hovering = hovering
}

func (s *ScrollOverlay) setDragging(draggingX, draggingY bool) {
	if s.draggingX == draggingX && s.draggingY == draggingY {
		return
	}

	s.draggingX = draggingX
	s.draggingY = draggingY
}

func adjustedWheel() (float64, float64) {
	x, y := ebiten.Wheel()
	switch runtime.GOOS {
	case "darwin":
		x *= 2
		y *= 2
	}
	return x, y
}

func (s *ScrollOverlay) HandleInput(context *guigui.Context, widget *guigui.Widget) guigui.HandleInputResult {
	s.setHovering(image.Pt(ebiten.CursorPosition()).In(widget.VisibleBounds()) && widget.IsVisible())

	if s.hovering {
		x, y := ebiten.CursorPosition()
		dx, dy := adjustedWheel()
		s.lastCursorX = x
		s.lastCursorY = y
		s.lastWheelX = dx
		s.lastWheelY = dy
	} else {
		s.lastCursorX = -1
		s.lastCursorY = -1
		s.lastWheelX = 0
		s.lastWheelY = 0
	}

	if !s.draggingX && !s.draggingY && s.hovering && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		hb, vb := s.barBounds(widget.Bounds(), context.Scale())
		if image.Pt(x, y).In(hb) {
			s.setDragging(true, s.draggingY)
			s.draggingStartX = x
			s.draggingStartOffsetX = s.offsetX
		} else if image.Pt(x, y).In(vb) {
			s.setDragging(s.draggingX, true)
			s.draggingStartY = y
			s.draggingStartOffsetY = s.offsetY
		}
		if s.draggingX || s.draggingY {
			return guigui.HandleInputByWidget(widget)
		}
	}

	if dx, dy := adjustedWheel(); dx != 0 || dy != 0 {
		s.setDragging(false, false)
	}

	if (s.draggingX || s.draggingY) && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		var dx, dy float64
		if s.draggingX {
			dx = float64(x - s.draggingStartX)
		}
		if s.draggingY {
			dy = float64(y - s.draggingStartY)
		}
		if dx != 0 || dy != 0 {
			prevOffsetX := s.offsetX
			prevOffsetY := s.offsetY

			barWidth, barHeight := s.barSize(widget.Bounds(), context.Scale())
			if s.draggingX && barWidth > 0 && s.contentWidth-widget.Bounds().Dx() > 0 {
				offsetPerPixel := float64(s.contentWidth-widget.Bounds().Dx()) / (float64(widget.Bounds().Dx()) - barWidth)
				s.offsetX = s.draggingStartOffsetX + float64(-dx)*offsetPerPixel
			}
			if s.draggingY && barHeight > 0 && s.contentHeight-widget.Bounds().Dy() > 0 {
				offsetPerPixel := float64(s.contentHeight-widget.Bounds().Dy()) / (float64(widget.Bounds().Dy()) - barHeight)
				s.offsetY = s.draggingStartOffsetY + float64(-dy)*offsetPerPixel
			}
			s.adjustOffset(widget)
			if prevOffsetX != s.offsetX || prevOffsetY != s.offsetY {
				widget.EnqueueEvent(ScrollEvent{
					OffsetX: s.offsetX,
					OffsetY: s.offsetY,
				})
				widget.RequestRedraw()
			}
		}
		return guigui.HandleInputByWidget(widget)
	}

	if (s.draggingX || s.draggingY) && !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		s.setDragging(false, false)
	}

	if dx, dy := adjustedWheel(); dx != 0 || dy != 0 {
		if !s.hovering {
			return guigui.HandleInputResult{}
		}
		s.setDragging(false, false)

		prevOffsetX := s.offsetX
		prevOffsetY := s.offsetY
		s.offsetX += dx * 4 * context.Scale()
		s.offsetY += dy * 4 * context.Scale()
		s.adjustOffset(widget)
		if prevOffsetX != s.offsetX || prevOffsetY != s.offsetY {
			widget.EnqueueEvent(ScrollEvent{
				OffsetX: s.offsetX,
				OffsetY: s.offsetY,
			})
			widget.RequestRedraw()
			return guigui.HandleInputByWidget(widget)
		}
		return guigui.HandleInputResult{}
	}

	return guigui.HandleInputResult{}
}

func (s *ScrollOverlay) CursorShape(context *guigui.Context, widget *guigui.Widget) (ebiten.CursorShapeType, bool) {
	x, y := ebiten.CursorPosition()
	hb, vb := s.barBounds(widget.Bounds(), context.Scale())
	if image.Pt(x, y).In(hb) || image.Pt(x, y).In(vb) {
		return ebiten.CursorShapeDefault, true
	}
	return 0, false
}

func (s *ScrollOverlay) Offset() (float64, float64) {
	return s.offsetX, s.offsetY
}

func (s *ScrollOverlay) adjustOffset(widget *guigui.Widget) {
	origOffsetX := s.offsetX
	origOffsetY := s.offsetY

	bounds := widget.Bounds()

	// Adjust offsets.
	if s.offsetX > 0 {
		s.offsetX = 0
	}
	if s.offsetY > 0 {
		s.offsetY = 0
	}

	w := s.contentWidth - bounds.Dx()
	h := s.contentHeight - bounds.Dy()
	if w < 0 {
		s.offsetX = 0
	} else if s.offsetX < -float64(w) {
		s.offsetX = -float64(w)
	}
	if h < 0 {
		s.offsetY = 0
	} else if s.offsetY < -float64(h) {
		s.offsetY = -float64(h)
	}

	if s.offsetX != origOffsetX || s.offsetY != origOffsetY {
		widget.RequestRedraw()
	}
}

func (s *ScrollOverlay) isBarVisible(context *guigui.Context, widget *guigui.Widget) bool {
	if s.draggingX || s.draggingY {
		return true
	}
	if s.lastWheelX != 0 || s.lastWheelY != 0 {
		return true
	}
	if s.lastOffsetX != s.offsetX || s.lastOffsetY != s.offsetY {
		return true
	}

	if s.contentWidth > widget.Bounds().Dx() && widget.Bounds().Max.Y-UnitSize(context) <= s.lastCursorY {
		return true
	}
	if s.contentHeight > widget.Bounds().Dy() && widget.Bounds().Max.X-UnitSize(context) <= s.lastCursorX {
		return true
	}
	return false
}

func (s *ScrollOverlay) Update(context *guigui.Context, widget *guigui.Widget) error {
	if s.needsAdjustOffset {
		s.adjustOffset(widget)
		s.needsAdjustOffset = false
	}
	if s.needsRedraw {
		widget.RequestRedraw()
		s.needsRedraw = false
	}

	if !widget.IsVisible() {
		s.setHovering(false)
	}

	if s.isBarVisible(context, widget) || (s.barVisibleTime == barShowingTime && s.barOpacity < barMaxOpacity) {
		if s.barOpacity < barMaxOpacity {
			s.barOpacity++
			widget.RequestRedraw()
		}
		s.barVisibleTime = barShowingTime
	} else {
		if s.barVisibleTime > 0 {
			s.barVisibleTime--
		}
		if s.barVisibleTime == 0 && s.barOpacity > 0 {
			s.barOpacity--
			widget.RequestRedraw()
		}
	}

	s.lastOffsetX = s.offsetX
	s.lastOffsetY = s.offsetY

	return nil
}

func (s *ScrollOverlay) Draw(context *guigui.Context, widget *guigui.Widget, dst *ebiten.Image) {
	if s.barOpacity == 0 {
		return
	}

	opacity := float64(s.barOpacity) / barMaxOpacity * 3 / 4
	r, g, b, a := Color(context.ColorMode(), ColorTypeBase, 0.2).RGBA()
	barColor := color.RGBA64{
		R: uint16(float64(r) * opacity),
		G: uint16(float64(g) * opacity),
		B: uint16(float64(b) * opacity),
		A: uint16(float64(a) * opacity),
	}

	hb, vb := s.barBounds(widget.Bounds(), context.Scale())

	// Show a horizontal bar.
	if !hb.Empty() {
		DrawRoundedRect(context, dst, hb, barColor, RoundedCornerRadius(context))
	}

	// Show a vertical bar.
	if !vb.Empty() {
		DrawRoundedRect(context, dst, vb, barColor, RoundedCornerRadius(context))
	}
}

func (s *ScrollOverlay) barWidth(scale float64) float64 {
	const scrollBarStrokeWidthInDIP = 8
	return scrollBarStrokeWidthInDIP * scale
}

func (s *ScrollOverlay) barSize(bounds image.Rectangle, scale float64) (float64, float64) {
	var w, h float64
	if s.contentWidth > bounds.Dx() {
		w = float64(bounds.Dx()) * float64(bounds.Dx()) / float64(s.contentWidth)
		if min := s.barWidth(scale); w < min {
			w = min
		}
	}
	if s.contentHeight > bounds.Dy() {
		h = float64(bounds.Dy()) * float64(bounds.Dy()) / float64(s.contentHeight)
		if min := s.barWidth(scale); h < min {
			h = min
		}
	}
	return w, h
}

func (s *ScrollOverlay) barBounds(bounds image.Rectangle, scale float64) (image.Rectangle, image.Rectangle) {
	offsetX, offsetY := s.Offset()
	barWidth, barHeight := s.barSize(bounds, scale)

	var horizontalBarBounds, verticalBarBounds image.Rectangle
	if s.contentWidth > bounds.Dx() {
		rate := -offsetX / float64(s.contentWidth-bounds.Dx())
		x0 := float64(bounds.Min.X) + rate*(float64(bounds.Dx())-barWidth)
		x1 := x0 + float64(barWidth)
		y0 := float64(bounds.Max.Y) - min(s.barWidth(scale), float64(bounds.Dy())*0.2)
		y1 := float64(bounds.Max.Y)
		horizontalBarBounds = image.Rect(int(x0), int(y0), int(x1), int(y1))
	}
	if s.contentHeight > bounds.Dy() {
		rate := -offsetY / float64(s.contentHeight-bounds.Dy())
		y0 := float64(bounds.Min.Y) + rate*(float64(bounds.Dy())-barHeight)
		y1 := y0 + float64(barHeight)
		x0 := float64(bounds.Max.X) - min(s.barWidth(scale), float64(bounds.Dx())*0.2)
		x1 := float64(bounds.Max.X)
		verticalBarBounds = image.Rect(int(x0), int(y0), int(x1), int(y1))
	}
	return horizontalBarBounds, verticalBarBounds
}
