// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/guigui"
)

type TextField struct {
	guigui.DefaultWidgetBehavior

	text  Text
	focus textFieldFocus

	widthMinusDefault  int
	heightMinusDefault int

	hovering bool
	readonly bool

	prevFocused bool
}

func (t *TextField) Text() string {
	return t.text.Text()
}

func (t *TextField) SetText(text string) {
	t.text.SetText(text)
}

func (t *TextField) SetMultiline(multiline bool) {
	t.text.SetMultiline(multiline)
}

func (t *TextField) SetHorizontalAlign(halign HorizontalAlign) {
	t.text.SetHorizontalAlign(halign)
}

func (t *TextField) SetVerticalAlign(valign VerticalAlign) {
	t.text.SetVerticalAlign(valign)
}

func (t *TextField) SetEditable(editable bool) {
	t.text.SetEditable(editable)
	t.readonly = !editable
}

func (t *TextField) SelectAll() {
	t.text.selectAll()
}

func (t *TextField) AppendChildWidgets(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	t.text.SetEditable(true)
	b := t.bounds(context)
	b.Min.X += UnitSize(context) / 2
	b.Max.X -= UnitSize(context) / 2
	t.text.SetSize(b.Dx(), b.Dy())
	// TODO: Consider multiline.
	if !t.text.IsMultiline() {
		t.text.SetVerticalAlign(VerticalAlignMiddle)
	}
	appender.AppendChildWidget(&t.text, b.Min)

	if context.WidgetFromBehavior(t).HasFocusedChildWidget() {
		w := textFieldFocusBorderWidth(context)
		p := context.WidgetFromBehavior(t).Position().Add(image.Pt(-w, -w))
		appender.AppendChildWidget(&t.focus, p)
	}
}

func (t *TextField) HandleInput(context *guigui.Context) guigui.HandleInputResult {
	x, y := ebiten.CursorPosition()
	t.hovering = image.Pt(x, y).In(context.WidgetFromBehavior(t).VisibleBounds())
	if t.hovering {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			context.WidgetFromBehavior(&t.text).Focus()
			t.text.selectAll()
			return guigui.HandleInputByWidget(t)
		}
	}
	return guigui.HandleInputResult{}
}

func (t *TextField) PropagateEvent(context *guigui.Context, widget *guigui.Widget, event guigui.Event) (guigui.Event, bool) {
	return event, true
}

func (t *TextField) Update(context *guigui.Context) error {
	if t.prevFocused != context.WidgetFromBehavior(t).HasFocusedChildWidget() {
		t.prevFocused = context.WidgetFromBehavior(t).HasFocusedChildWidget()
		context.WidgetFromBehavior(t).RequestRedraw()
	}
	if context.WidgetFromBehavior(t).IsFocused() {
		context.WidgetFromBehavior(&t.text).Focus()
		context.WidgetFromBehavior(t).RequestRedraw()
	}
	return nil
}

func (t *TextField) Draw(context *guigui.Context, dst *ebiten.Image) {
	bounds := t.bounds(context)
	DrawRoundedRect(context, dst, bounds, Color(context.ColorMode(), ColorTypeBase, 0.85), RoundedCornerRadius(context))
	DrawRoundedRectBorder(context, dst, bounds, Color2(context.ColorMode(), ColorTypeBase, 0.7, 0), RoundedCornerRadius(context), float32(1*context.Scale()), RoundedRectBorderTypeInset)
}

func defaultTextFieldSize(context *guigui.Context) (int, int) {
	// TODO: Increase the height for multiple lines.
	return 6 * UnitSize(context), UnitSize(context)
}

func (t *TextField) bounds(context *guigui.Context) image.Rectangle {
	w, h := t.Size(context)
	p := context.WidgetFromBehavior(t).Position()
	return image.Rectangle{
		Min: p,
		Max: p.Add(image.Pt(w, h)),
	}
}

func (t *TextField) SetSize(context *guigui.Context, width, height int) {
	dw, dh := defaultTextFieldSize(context)
	t.widthMinusDefault = width - dw
	t.heightMinusDefault = height - dh
}

func (t *TextField) Size(context *guigui.Context) (int, int) {
	dw, dh := defaultTextFieldSize(context)
	if t.text.multiline {
		return t.widthMinusDefault + dw, t.heightMinusDefault + dh
	}
	return t.widthMinusDefault + dw, dh
}

func textFieldFocusBorderWidth(context *guigui.Context) int {
	return int(3 * context.Scale())
}

type textFieldFocus struct {
	guigui.DefaultWidgetBehavior
}

func (t *textFieldFocus) Draw(context *guigui.Context, dst *ebiten.Image) {
	textField := context.WidgetFromBehavior(t).Parent().Behavior().(*TextField)
	bounds := textField.bounds(context)
	w := textFieldFocusBorderWidth(context)
	bounds = bounds.Inset(-w)
	DrawRoundedRectBorder(context, dst, bounds, Color(context.ColorMode(), ColorTypeAccent, 0.8), int(4*context.Scale())+RoundedCornerRadius(context), float32(4*context.Scale()), RoundedRectBorderTypeRegular)
}

func (t *textFieldFocus) IsPopup() bool {
	return true
}

func (t *textFieldFocus) Size(context *guigui.Context) (int, int) {
	w, h := context.WidgetFromBehavior(t).Parent().Behavior().Size(context)
	w += 2 * textFieldFocusBorderWidth(context)
	h += 2 * textFieldFocusBorderWidth(context)
	return w, h
}
