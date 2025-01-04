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

	mouseEventHandlerWidget *guigui.Widget

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

func (b *Button) mouseEventHandler() *guigui.MouseEventHandler {
	return b.mouseEventHandlerWidget.Behavior().(*guigui.MouseEventHandler)
}

func (b *Button) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	if b.mouseEventHandlerWidget == nil {
		b.mouseEventHandlerWidget = guigui.NewWidget(&guigui.MouseEventHandler{})
	}
	appender.AppendChildWidget(b.mouseEventHandlerWidget, widget.Position())
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

func (b *Button) CursorShape(context *guigui.Context, widget *guigui.Widget) (ebiten.CursorShapeType, bool) {
	if widget.IsEnabled() && b.mouseEventHandler().IsHovering() {
		return ebiten.CursorShapePointer, true
	}
	return 0, true
}

func (b *Button) Draw(context *guigui.Context, widget *guigui.Widget, dst *ebiten.Image) {
	// TODO: In the dark theme, the color should be different.
	// At least, shadow should be darker.
	// See macOS's buttons.
	cm := context.ColorMode()
	backgroundColor := Color2(cm, ColorTypeBase, 1, 0.3)
	borderColor := Color2(cm, ColorTypeBase, 0.7, 0)
	if b.isActive() {
		backgroundColor = Color2(cm, ColorTypeBase, 0.95, 0.25)
		borderColor = Color2(cm, ColorTypeBase, 0.7, 0)
	} else if b.mouseEventHandler().IsHovering() && b.mouseEventHandlerWidget.IsEnabled() {
		backgroundColor = Color2(cm, ColorTypeBase, 0.975, 0.275)
		borderColor = Color2(cm, ColorTypeBase, 0.7, 0)
	} else if !widget.IsEnabled() {
		backgroundColor = Color2(cm, ColorTypeBase, 0.95, 0.25)
		borderColor = Color2(cm, ColorTypeBase, 0.8, 0.1)
	}

	bounds := b.bounds(context, widget)
	r := min(RoundedCornerRadius(context), bounds.Dx()/4, bounds.Dy()/4)
	border := !b.borderInvisible
	if b.mouseEventHandler().IsHovering() && b.mouseEventHandlerWidget.IsEnabled() {
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
		} else if !widget.IsEnabled() {
			borderType = RoundedRectBorderTypeRegular
		}
		DrawRoundedRectBorder(context, dst, bounds, borderColor, r, float32(1*context.Scale()), borderType)
	}
}

func (b *Button) isActive() bool {
	if b.mouseEventHandlerWidget == nil {
		return false
	}
	return b.mouseEventHandlerWidget.IsEnabled() && b.mouseEventHandler().IsHovering() && b.mouseEventHandler().IsPressing()
}

func (b *Button) bounds(context *guigui.Context, widget *guigui.Widget) image.Rectangle {
	dw, dh := defaultButtonSize(context)
	p := widget.Position()
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

func (b *Button) Size(context *guigui.Context, widget *guigui.Widget) (int, int) {
	dw, dh := defaultButtonSize(context)
	return b.widthMinusDefault + dw, b.heightMinusDefault + dh
}

type TextButton struct {
	guigui.DefaultWidgetBehavior

	button       Button
	buttonWidget *guigui.Widget
	text         Text
	textWidget   *guigui.Widget

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

func (t *TextButton) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	w, h := widget.Size(context)

	if t.buttonWidget == nil {
		t.buttonWidget = guigui.NewWidget(&t.button)
	}
	t.buttonWidget.Behavior().(*Button).SetSize(context, w, h)
	appender.AppendChildWidget(t.buttonWidget, widget.Position())

	if t.textWidget == nil {
		t.text.SetHorizontalAlign(HorizontalAlignCenter)
		t.text.SetVerticalAlign(VerticalAlignMiddle)
		t.textWidget = guigui.NewWidget(&t.text)
	}
	p := widget.Position()
	if t.button.isActive() {
		// As the text is centered, shift it down by double sizes of the stroke width.
		p.Y += int(2 * context.Scale())
	} else if !t.buttonWidget.IsEnabled() {
		p.Y += int(1 * context.Scale())
	}
	t.textWidget.Behavior().(*Text).SetSize(w, h)
	appender.AppendChildWidget(t.textWidget, p)
}

func (t *TextButton) PropagateEvent(context *guigui.Context, widget *guigui.Widget, event guigui.Event) (guigui.Event, bool) {
	return event, true
}

func (t *TextButton) Update(context *guigui.Context, widget *guigui.Widget) error {
	if t.needsRedraw {
		widget.RequestRedraw()
		t.needsRedraw = false
	}

	if !t.buttonWidget.IsEnabled() {
		t.text.SetColor(Color(context.ColorMode(), ColorTypeBase, 0.5))
	} else {
		t.text.SetColor(t.textColor)
	}
	return nil
}

func (t *TextButton) Size(context *guigui.Context, widget *guigui.Widget) (int, int) {
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
