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

	mouseEventHandler guigui.MouseEventHandler

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

func (t *ToggleButton) AppendChildWidgets(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	appender.AppendChildWidget(&t.mouseEventHandler, context.WidgetFromBehavior(t).Position())
}

func (t *ToggleButton) Update(context *guigui.Context) error {
	for e := range context.WidgetFromBehavior(&t.mouseEventHandler).DequeueEvents() {
		switch e := e.(type) {
		case guigui.MouseEvent:
			if e.Type == guigui.MouseEventTypeUp {
				t.SetValue(!t.value)
			}
		}
	}

	if t.needsRedraw {
		context.WidgetFromBehavior(t).RequestRedraw()
		t.needsRedraw = false
	}
	if t.count > 0 {
		t.count--
		context.WidgetFromBehavior(t).RequestRedraw()
	}
	return nil
}

func (t *ToggleButton) CursorShape(context *guigui.Context) (ebiten.CursorShapeType, bool) {
	if context.WidgetFromBehavior(t).IsEnabled() && t.mouseEventHandler.IsHovering() {
		return ebiten.CursorShapePointer, true
	}
	return 0, true
}

func (t *ToggleButton) Draw(context *guigui.Context, dst *ebiten.Image) {
	rate := 1 - float64(t.count)/float64(toggleButtonMaxCount())

	bounds := t.bounds(context)

	cm := context.ColorMode()
	backgroundColor := Color(context.ColorMode(), ColorTypeBase, 0.8)
	thumbColor := Color2(cm, ColorTypeBase, 1, 0.6)
	borderColor := Color2(cm, ColorTypeBase, 0.7, 0)
	if t.isActive(context) {
		thumbColor = Color2(cm, ColorTypeBase, 0.95, 0.55)
		borderColor = Color2(cm, ColorTypeBase, 0.7, 0)
	} else if t.mouseEventHandler.IsHovering() && context.WidgetFromBehavior(&t.mouseEventHandler).IsEnabled() {
		thumbColor = Color2(cm, ColorTypeBase, 0.975, 0.575)
		borderColor = Color2(cm, ColorTypeBase, 0.7, 0)
	} else if !context.WidgetFromBehavior(t).IsEnabled() {
		thumbColor = Color2(cm, ColorTypeBase, 0.95, 0.55)
		borderColor = Color2(cm, ColorTypeBase, 0.8, 0.1)
	}

	// Background
	bgColorOff := backgroundColor
	bgColorOn := Color(context.ColorMode(), ColorTypeAccent, 0.5)
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

func (t *ToggleButton) isActive(context *guigui.Context) bool {
	return context.WidgetFromBehavior(t).IsEnabled() && t.mouseEventHandler.IsHovering() && t.mouseEventHandler.IsPressing()
}

func (t *ToggleButton) bounds(context *guigui.Context) image.Rectangle {
	cw, ch := t.Size(context)
	p := context.WidgetFromBehavior(t).Position()
	return image.Rectangle{
		Min: p,
		Max: p.Add(image.Pt(cw, ch)),
	}
}

func (t *ToggleButton) Size(context *guigui.Context) (int, int) {
	return int(LineHeight(context) * 1.75), int(LineHeight(context))
}
