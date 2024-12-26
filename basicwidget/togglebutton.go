// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/guigui"
)

type ToggleButtonEvent struct {
	Value bool
}

type ToggleButton struct {
	guigui.DefaultWidgetBehavior

	mouseEventHandlerWidget *guigui.Widget

	value        bool
	onceRendered bool

	count int

	needsRedraw bool
}

func (t *ToggleButton) Value() bool {
	return t.value
}

func (t *ToggleButton) SetValue(value bool) {
	if t.value == value {
		return
	}

	t.value = value
	if t.onceRendered {
		t.count = toggleButtonMaxCount() - t.count
	}
	t.needsRedraw = true
}

func toggleButtonMaxCount() int {
	return ebiten.TPS() / 12
}

func (t *ToggleButton) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	if t.mouseEventHandlerWidget == nil {
		t.mouseEventHandlerWidget = guigui.NewWidget(&guigui.MouseEventHandler{})
	}
	appender.AppendChildWidgetWithBounds(t.mouseEventHandlerWidget, t.bounds(context, widget))
}

func (t *ToggleButton) Update(context *guigui.Context, widget *guigui.Widget) error {
	for e := range t.mouseEventHandlerWidget.DequeueEvents() {
		switch e := e.(type) {
		case guigui.MouseEvent:
			if e.Type == guigui.MouseEventTypeUp {
				t.SetValue(!t.value)
			}
		}
	}

	if t.needsRedraw {
		widget.RequestRedraw()
		t.needsRedraw = false
	}
	if t.count > 0 {
		t.count--
		widget.RequestRedraw()
	}
	return nil
}

func (t *ToggleButton) CursorShape(context *guigui.Context, widget *guigui.Widget) (ebiten.CursorShapeType, bool) {
	if widget.IsEnabled() && t.mouseEventHandler().IsHovering() {
		return ebiten.CursorShapePointer, true
	}
	return 0, true
}

func (t *ToggleButton) Draw(context *guigui.Context, widget *guigui.Widget, dst *ebiten.Image) {
	rate := 1 - float64(t.count)/float64(toggleButtonMaxCount())

	bounds := t.bounds(context, widget)

	cm := context.ColorMode()
	backgroundColor := Color(context.ColorMode(), ColorTypeBase, 0.8)
	thumbColor := Color2(cm, ColorTypeBase, 1, 0.6)
	borderColor := Color2(cm, ColorTypeBase, 0.7, 0.3)
	if t.isActive() {
		thumbColor = Color2(cm, ColorTypeBase, 0.95, 0.55)
		borderColor = Color2(cm, ColorTypeBase, 0.7, 0.3)
	} else if t.mouseEventHandler().IsHovering() && t.mouseEventHandlerWidget.IsEnabled() {
		thumbColor = Color2(cm, ColorTypeBase, 0.975, 0.575)
		borderColor = Color2(cm, ColorTypeBase, 0.7, 0.3)
	} else if !widget.IsEnabled() {
		thumbColor = Color2(cm, ColorTypeBase, 0.95, 0.55)
		borderColor = Color2(cm, ColorTypeBase, 0.8, 0.4)
	}

	// Background
	bgColorOff := backgroundColor
	bgColorOn := Color(context.ColorMode(), ColorTypeAccent, 0.6)
	var bgColor color.Color
	if t.value {
		bgColor = mixColor(bgColorOff, bgColorOn, rate)
	} else {
		bgColor = mixColor(bgColorOn, bgColorOff, rate)
	}
	r := bounds.Dy() / 2
	DrawRoundedRect(context, dst, bounds, bgColor, r)

	// Border (upper)
	b := bounds
	b.Max.Y = b.Min.Y + b.Dy()/2
	DrawRoundedRectBorder(context, dst.SubImage(b).(*ebiten.Image), bounds, borderColor, r, float32(1*context.Scale()), RoundedRectBorderTypeInset)

	// Thumb
	cxOff := float64(bounds.Min.X) + float64(r)
	cxOn := float64(bounds.Max.X) - float64(r)
	var cx int
	if t.value {
		cx = int((1-rate)*cxOff + rate*cxOn)
	} else {
		cx = int((1-rate)*cxOn + rate*cxOff)
	}
	cy := bounds.Min.Y + r
	DrawRoundedRect(context, dst, image.Rect(cx-r, cy-r, cx+r, cy+r), thumbColor, r)
	DrawRoundedRectBorder(context, dst, image.Rect(cx-r, cy-r, cx+r, cy+r), borderColor, r, float32(1*context.Scale()), RoundedRectBorderTypeOutset)

	// Border (lower)
	b = bounds
	b.Min.Y = b.Max.Y - b.Dy()/2
	DrawRoundedRectBorder(context, dst.SubImage(b).(*ebiten.Image), bounds, borderColor, r, float32(1*context.Scale()), RoundedRectBorderTypeInset)

	t.onceRendered = true
}

func (t *ToggleButton) mouseEventHandler() *guigui.MouseEventHandler {
	if t.mouseEventHandlerWidget == nil {
		return nil
	}
	return t.mouseEventHandlerWidget.Behavior().(*guigui.MouseEventHandler)
}

func (t *ToggleButton) isActive() bool {
	if t.mouseEventHandlerWidget == nil {
		return false
	}
	return t.mouseEventHandlerWidget.IsEnabled() && t.mouseEventHandler().IsHovering() && t.mouseEventHandler().IsPressing()
}

func (t *ToggleButton) bounds(context *guigui.Context, widget *guigui.Widget) image.Rectangle {
	cw, ch := t.Size(context, widget)
	b := widget.Bounds()
	b.Max.X = b.Min.X + cw
	b.Max.Y = b.Min.Y + ch
	return b
}

func (t *ToggleButton) Size(context *guigui.Context, widget *guigui.Widget) (int, int) {
	return int(LineHeight(context) * 1.75), int(LineHeight(context))
}
