// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/guigui"
)

type Button struct {
	guigui.DefaultWidgetBehavior

	mouseEventHandler guigui.MouseEventHandler

	widthMinusDefault  int
	heightMinusDefault int
	borderInvisible    bool
}

type ButtonEventType int

const (
	ButtonEventTypeDown ButtonEventType = iota
	ButtonEventTypeUp
)

type ButtonEvent struct {
	Type ButtonEventType
}

func (b *Button) AppendChildWidgets(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	appender.AppendChildWidget(&b.mouseEventHandler, context.WidgetFromBehavior(b).Position())
}

func (b *Button) PropagateEvent(context *guigui.Context, widget *guigui.Widget, event guigui.Event) (guigui.Event, bool) {
	args, ok := event.(guigui.MouseEvent)
	if !ok {
		return nil, false
	}
	if !image.Pt(args.CursorPositionX, args.CursorPositionY).In(widget.VisibleBounds()) {
		return nil, false
	}
	var typ ButtonEventType
	switch args.Type {
	case guigui.MouseEventTypeDown:
		typ = ButtonEventTypeDown
	case guigui.MouseEventTypeUp:
		typ = ButtonEventTypeUp
	default:
		return nil, false
	}

	return ButtonEvent{
		Type: typ,
	}, true
}

func (b *Button) CursorShape(context *guigui.Context) (ebiten.CursorShapeType, bool) {
	if context.WidgetFromBehavior(b).IsEnabled() && b.mouseEventHandler.IsHovering() {
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
	if b.isActive(context) {
		backgroundColor = Color2(cm, ColorTypeBase, 0.95, 0.25)
		borderColor = Color2(cm, ColorTypeBase, 0.7, 0)
	} else if b.mouseEventHandler.IsHovering() && context.WidgetFromBehavior(&b.mouseEventHandler).IsEnabled() {
		backgroundColor = Color2(cm, ColorTypeBase, 0.975, 0.275)
		borderColor = Color2(cm, ColorTypeBase, 0.7, 0)
	} else if !context.WidgetFromBehavior(b).IsEnabled() {
		backgroundColor = Color2(cm, ColorTypeBase, 0.95, 0.25)
		borderColor = Color2(cm, ColorTypeBase, 0.8, 0.1)
	}

	bounds := b.bounds(context)
	r := min(RoundedCornerRadius(context), bounds.Dx()/4, bounds.Dy()/4)
	border := !b.borderInvisible
	if b.mouseEventHandler.IsHovering() && context.WidgetFromBehavior(&b.mouseEventHandler).IsEnabled() {
		border = true
	}
	if border || b.isActive(context) {
		bounds := bounds.Inset(int(1 * context.Scale()))
		DrawRoundedRect(context, dst, bounds, backgroundColor, r)
	}

	if border {
		borderType := RoundedRectBorderTypeOutset
		if b.isActive(context) {
			borderType = RoundedRectBorderTypeInset
		} else if !context.WidgetFromBehavior(b).IsEnabled() {
			borderType = RoundedRectBorderTypeRegular
		}
		DrawRoundedRectBorder(context, dst, bounds, borderColor, r, float32(1*context.Scale()), borderType)
	}
}

func (b *Button) isActive(context *guigui.Context) bool {
	return context.WidgetFromBehavior(&b.mouseEventHandler).IsEnabled() && b.mouseEventHandler.IsHovering() && b.mouseEventHandler.IsPressing()
}

func (b *Button) bounds(context *guigui.Context) image.Rectangle {
	dw, dh := defaultButtonSize(context)
	p := context.WidgetFromBehavior(b).Position()
	return image.Rectangle{
		Min: p,
		Max: p.Add(image.Pt(b.widthMinusDefault+dw, b.heightMinusDefault+dh)),
	}
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

type TextButton struct {
	guigui.DefaultWidgetBehavior

	button Button
	text   Text

	textColor color.Color

	width    int
	widthSet bool

	needsRedraw bool
}

func (t *TextButton) SetText(text string) {
	t.text.SetText(text)
}

func (t *TextButton) SetTextColor(clr color.Color) {
	if equalColor(t.textColor, clr) {
		return
	}
	t.textColor = clr
	t.needsRedraw = true
}

func (t *TextButton) AppendChildWidgets(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	w, h := t.Size(context)

	t.button.SetSize(context, w, h)
	appender.AppendChildWidget(&t.button, context.WidgetFromBehavior(t).Position())

	t.text.SetHorizontalAlign(HorizontalAlignCenter)
	t.text.SetVerticalAlign(VerticalAlignMiddle)
	p := context.WidgetFromBehavior(t).Position()
	if t.button.isActive(context) {
		// As the text is centered, shift it down by double sizes of the stroke width.
		p.Y += int(2 * context.Scale())
	} else if !context.WidgetFromBehavior(&t.button).IsEnabled() {
		p.Y += int(1 * context.Scale())
	}
	t.text.SetSize(w, h)
	appender.AppendChildWidget(&t.text, p)
}

func (t *TextButton) PropagateEvent(context *guigui.Context, widget *guigui.Widget, event guigui.Event) (guigui.Event, bool) {
	return event, true
}

func (t *TextButton) Update(context *guigui.Context) error {
	if t.needsRedraw {
		context.WidgetFromBehavior(t).RequestRedraw()
		t.needsRedraw = false
	}

	if !context.WidgetFromBehavior(&t.button).IsEnabled() {
		t.text.SetColor(Color(context.ColorMode(), ColorTypeBase, 0.5))
	} else {
		t.text.SetColor(t.textColor)
	}
	return nil
}

func (t *TextButton) Size(context *guigui.Context) (int, int) {
	_, dh := defaultButtonSize(context)
	if t.widthSet {
		return t.width, dh
	}
	tw, _ := t.text.TextSize(context)
	return tw + UnitSize(context), dh
}

func (t *TextButton) SetWidth(width int) {
	t.width = width
	t.widthSet = true
}

func (t *TextButton) ResetWidth() {
	t.width = 0
	t.widthSet = false
}
