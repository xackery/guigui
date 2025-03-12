// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/xackery/guigui"
)

type ToggleButton struct {
	guigui.DefaultWidget

	mouseOverlay guigui.MouseOverlay

	value        bool
	onceRendered bool

	count int

	onValueChanged func(value bool)
}

func (t *ToggleButton) SetOnValueChanged(f func(value bool)) {
	t.onValueChanged = f
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
	guigui.RequestRedraw(t)

	if t.onValueChanged != nil {
		t.onValueChanged(value)
	}
}

func toggleButtonMaxCount() int {
	return ebiten.TPS() / 12
}

func (t *ToggleButton) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	t.mouseOverlay.SetOnUp(func(mouseButton ebiten.MouseButton, cursorPosition image.Point) {
		if mouseButton != ebiten.MouseButtonLeft {
			return
		}
		t.SetValue(!t.value)
	})

	guigui.SetPosition(&t.mouseOverlay, guigui.Position(t))
	appender.AppendChildWidget(&t.mouseOverlay)
}

func (t *ToggleButton) Update(context *guigui.Context) error {
	if t.count > 0 {
		t.count--
		guigui.RequestRedraw(t)
	}
	return nil
}

func (t *ToggleButton) CursorShape(context *guigui.Context) (ebiten.CursorShapeType, bool) {
	if guigui.IsEnabled(t) && t.mouseOverlay.IsHovering() {
		return ebiten.CursorShapePointer, true
	}
	return 0, true
}

func (t *ToggleButton) Draw(context *guigui.Context, dst *ebiten.Image) {
	rate := 1 - float64(t.count)/float64(toggleButtonMaxCount())

	bounds := guigui.Bounds(t)

	cm := context.ColorMode()
	backgroundColor := Color(context.ColorMode(), ColorTypeBase, 0.8)
	thumbColor := Color2(cm, ColorTypeBase, 1, 0.6)
	borderColor := Color2(cm, ColorTypeBase, 0.7, 0)
	if t.isActive() {
		thumbColor = Color2(cm, ColorTypeBase, 0.95, 0.55)
		borderColor = Color2(cm, ColorTypeBase, 0.7, 0)
	} else if t.mouseOverlay.IsHovering() && guigui.IsEnabled(&t.mouseOverlay) {
		thumbColor = Color2(cm, ColorTypeBase, 0.975, 0.575)
		borderColor = Color2(cm, ColorTypeBase, 0.7, 0)
	} else if !guigui.IsEnabled(t) {
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

func (t *ToggleButton) isActive() bool {
	return guigui.IsEnabled(t) && t.mouseOverlay.IsHovering() && t.mouseOverlay.IsPressing()
}

func (t *ToggleButton) Size(context *guigui.Context) (int, int) {
	return int(LineHeight(context) * 1.75), int(LineHeight(context))
}
