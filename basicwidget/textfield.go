// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/xackery/guigui"
)

type TextField struct {
	guigui.DefaultWidget

	text  Text
	focus textFieldFocus

	widthMinusDefault  int
	heightMinusDefault int

	hovering bool
	readonly bool

	prevFocused bool
}

func (t *TextField) SetOnEnterPressed(f func(text string)) {
	t.text.SetOnEnterPressed(f)
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

func (t *TextField) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	t.text.SetEditable(true)
	b := guigui.Bounds(t)
	b.Min.X += UnitSize(context) / 2
	b.Max.X -= UnitSize(context) / 2
	t.text.SetSize(b.Dx(), b.Dy())
	// TODO: Consider multiline.
	if !t.text.IsMultiline() {
		t.text.SetVerticalAlign(VerticalAlignMiddle)
	}
	guigui.SetPosition(&t.text, b.Min)
	appender.AppendChildWidget(&t.text)

	if guigui.HasFocusedChildWidget(t) {
		w := textFieldFocusBorderWidth(context)
		p := guigui.Position(t).Add(image.Pt(-w, -w))
		guigui.SetPosition(&t.focus, p)
		appender.AppendChildWidget(&t.focus)
	}
}

func (t *TextField) HandleInput(context *guigui.Context) guigui.HandleInputResult {
	x, y := ebiten.CursorPosition()
	t.hovering = image.Pt(x, y).In(guigui.VisibleBounds(t))
	if t.hovering {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			guigui.Focus(&t.text)
			t.text.selectAll()
			return guigui.HandleInputByWidget(t)
		}
	}
	return guigui.HandleInputResult{}
}

func (t *TextField) Update(context *guigui.Context) error {
	if t.prevFocused != guigui.HasFocusedChildWidget(t) {
		t.prevFocused = guigui.HasFocusedChildWidget(t)
		guigui.RequestRedraw(t)
	}
	if guigui.IsFocused(t) {
		guigui.Focus(&t.text)
		guigui.RequestRedraw(t)
	}
	return nil
}

func (t *TextField) Draw(context *guigui.Context, dst *ebiten.Image) {
	bounds := guigui.Bounds(t)
	DrawRoundedRect(context, dst, bounds, Color(context.ColorMode(), ColorTypeBase, 0.85), RoundedCornerRadius(context))
	DrawRoundedRectBorder(context, dst, bounds, Color2(context.ColorMode(), ColorTypeBase, 0.7, 0), RoundedCornerRadius(context), float32(1*context.Scale()), RoundedRectBorderTypeInset)
}

func defaultTextFieldSize(context *guigui.Context) (int, int) {
	// TODO: Increase the height for multiple lines.
	return 6 * UnitSize(context), UnitSize(context)
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
	guigui.DefaultWidget
}

func (t *textFieldFocus) Draw(context *guigui.Context, dst *ebiten.Image) {
	textField := guigui.Parent(t).(*TextField)
	bounds := guigui.Bounds(textField)
	w := textFieldFocusBorderWidth(context)
	bounds = bounds.Inset(-w)
	DrawRoundedRectBorder(context, dst, bounds, Color(context.ColorMode(), ColorTypeAccent, 0.8), int(4*context.Scale())+RoundedCornerRadius(context), float32(4*context.Scale()), RoundedRectBorderTypeRegular)
}

func (t *textFieldFocus) IsPopup() bool {
	return true
}

func (t *textFieldFocus) Size(context *guigui.Context) (int, int) {
	w, h := guigui.Parent(t).Size(context)
	w += 2 * textFieldFocusBorderWidth(context)
	h += 2 * textFieldFocusBorderWidth(context)
	return w, h
}
