// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/guigui"
)

type Button struct {
	guigui.DefaultWidget

	mouseOverlay guigui.MouseOverlay

	widthMinusDefault  int
	heightMinusDefault int
	borderInvisible    bool

	onDown func()
	onUp   func()
}

func (b *Button) SetOnDown(f func()) {
	b.onDown = f
}

func (b *Button) SetOnUp(f func()) {
	b.onUp = f
}

func (b *Button) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	b.mouseOverlay.SetOnDown(func(mouseButton ebiten.MouseButton, cursorPosition image.Point) {
		if mouseButton != ebiten.MouseButtonLeft {
			return
		}
		if b.onDown != nil {
			b.onDown()
		}
	})
	b.mouseOverlay.SetOnUp(func(mouseButton ebiten.MouseButton, cursorPosition image.Point) {
		if mouseButton != ebiten.MouseButtonLeft {
			return
		}
		if b.onUp != nil {
			b.onUp()
		}
	})

	guigui.SetPosition(&b.mouseOverlay, guigui.Position(b))
	appender.AppendChildWidget(&b.mouseOverlay)
}

func (b *Button) CursorShape(context *guigui.Context) (ebiten.CursorShapeType, bool) {
	if guigui.IsEnabled(b) && b.mouseOverlay.IsHovering() {
		return ebiten.CursorShapePointer, true
	}
	return 0, true
}

func (b *Button) Draw(context *guigui.Context, dst *ebiten.Image) {
	// TODO: In the dark theme, the color should be different.
	// At least, shadow should be darker.
	// See macOS's buttons.
	cm := context.ColorMode()
	backgroundColor := Color2(cm, ColorTypeBase, 1, 0.3)
	borderColor := Color2(cm, ColorTypeBase, 0.7, 0)
	if b.isActive() {
		backgroundColor = Color2(cm, ColorTypeBase, 0.95, 0.25)
		borderColor = Color2(cm, ColorTypeBase, 0.7, 0)
	} else if b.mouseOverlay.IsHovering() && guigui.IsEnabled(&b.mouseOverlay) {
		backgroundColor = Color2(cm, ColorTypeBase, 0.975, 0.275)
		borderColor = Color2(cm, ColorTypeBase, 0.7, 0)
	} else if !guigui.IsEnabled(b) {
		backgroundColor = Color2(cm, ColorTypeBase, 0.95, 0.25)
		borderColor = Color2(cm, ColorTypeBase, 0.8, 0.1)
	}

	bounds := guigui.Bounds(b)
	r := min(RoundedCornerRadius(context), bounds.Dx()/4, bounds.Dy()/4)
	border := !b.borderInvisible
	if b.mouseOverlay.IsHovering() && guigui.IsEnabled(&b.mouseOverlay) {
		border = true
	}
	if border || b.isActive() {
		bounds := bounds.Inset(int(1 * context.Scale()))
		DrawRoundedRect(context, dst, bounds, backgroundColor, r)
	}

	if border {
		borderType := RoundedRectBorderTypeOutset
		if b.isActive() {
			borderType = RoundedRectBorderTypeInset
		} else if !guigui.IsEnabled(b) {
			borderType = RoundedRectBorderTypeRegular
		}
		DrawRoundedRectBorder(context, dst, bounds, borderColor, r, float32(1*context.Scale()), borderType)
	}
}

func (b *Button) isActive() bool {
	return guigui.IsEnabled(b) && b.mouseOverlay.IsHovering() && b.mouseOverlay.IsPressing()
}

func defaultButtonSize(context *guigui.Context) (int, int) {
	return 6 * UnitSize(context), UnitSize(context)
}

func (b *Button) SetSize(context *guigui.Context, width, height int) {
	dw, dh := defaultButtonSize(context)
	b.widthMinusDefault = width - dw
	b.heightMinusDefault = height - dh
}

func (b *Button) Size(context *guigui.Context) (int, int) {
	dw, dh := defaultButtonSize(context)
	return b.widthMinusDefault + dw, b.heightMinusDefault + dh
}
