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

	text        Text
	textWidget  *guigui.Widget
	focusWidget *guigui.Widget

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

func (t *TextField) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	if t.textWidget == nil {
		t.text.SetEditable(true)
		t.textWidget = guigui.NewWidget(&t.text)
	}
	bounds := t.bounds(context, widget)
	bounds.Min.X += UnitSize(context) / 2
	bounds.Max.X -= UnitSize(context) / 2
	// TODO: Consider multiline.
	if !t.text.IsMultiline() {
		t.text.SetVerticalAlign(VerticalAlignMiddle)
	}
	appender.AppendChildWidgetWithBounds(t.textWidget, bounds)

	if widget.HasFocusedChildWidget() {
		if t.focusWidget == nil {
			t.focusWidget = guigui.NewPopupWidget(&textFieldFocus{})
		}
		w := textFieldFocusBorderWidth(context)
		p := widget.Position().Add(image.Pt(-w, -w))
		appender.AppendChildWidget(t.focusWidget, p)
	}
}

func (t *TextField) HandleInput(context *guigui.Context, widget *guigui.Widget) guigui.HandleInputResult {
	x, y := ebiten.CursorPosition()
	t.hovering = image.Pt(x, y).In(widget.VisibleBounds())
	if t.hovering {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			t.textWidget.Focus()
			t.text.selectAll()
			return guigui.HandleInputByWidget(widget)
		}
	}
	return guigui.HandleInputResult{}
}

func (t *TextField) PropagateEvent(context *guigui.Context, widget *guigui.Widget, event guigui.Event) (guigui.Event, bool) {
	return event, true
}

func (t *TextField) Update(context *guigui.Context, widget *guigui.Widget) error {
	if t.prevFocused != widget.HasFocusedChildWidget() {
		t.prevFocused = widget.HasFocusedChildWidget()
		widget.RequestRedraw()
	}
	if widget.IsFocused() {
		t.textWidget.Focus()
		widget.RequestRedraw()
	}
	return nil
}

func (t *TextField) Draw(context *guigui.Context, widget *guigui.Widget, dst *ebiten.Image) {
	bounds := t.bounds(context, widget)
	DrawRoundedRect(context, dst, bounds, Color(context.ColorMode(), ColorTypeBase, 0.85), RoundedCornerRadius(context))
	DrawRoundedRectBorder(context, dst, bounds, Color(context.ColorMode(), ColorTypeBase, 0.8), RoundedCornerRadius(context), float32(1*context.Scale()), RoundedRectBorderTypeInset)
}

func defaultTextFieldSize(context *guigui.Context) (int, int) {
	// TODO: Increase the height for multiple lines.
	return 6 * UnitSize(context), UnitSize(context)
}

func (t *TextField) bounds(context *guigui.Context, widget *guigui.Widget) image.Rectangle {
	dw, dh := defaultTextFieldSize(context)
	p := widget.Position()
	return image.Rectangle{
		Min: p,
		Max: p.Add(image.Pt(t.widthMinusDefault+dw, t.heightMinusDefault+dh)),
	}
}

func (t *TextField) SetSize(context *guigui.Context, width, height int) {
	dw, dh := defaultTextFieldSize(context)
	t.widthMinusDefault = width - dw
	t.heightMinusDefault = height - dh
}

func (t *TextField) Size(context *guigui.Context, widget *guigui.Widget) (int, int) {
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

func (t *textFieldFocus) Draw(context *guigui.Context, widget *guigui.Widget, dst *ebiten.Image) {
	textFieldWidget := widget.Parent()
	bounds := textFieldWidget.Behavior().(*TextField).bounds(context, textFieldWidget)
	w := textFieldFocusBorderWidth(context)
	bounds = bounds.Inset(-w)
	DrawRoundedRectBorder(context, dst, bounds, Color(context.ColorMode(), ColorTypeAccent, 0.8), int(4*context.Scale())+RoundedCornerRadius(context), float32(4*context.Scale()), RoundedRectBorderTypeRegular)
}

func (t *textFieldFocus) Size(context *guigui.Context, widget *guigui.Widget) (int, int) {
	textFieldWidget := widget.Parent()
	w, h := textFieldWidget.Behavior().(*TextField).Size(context, textFieldWidget)
	w += 2 * textFieldFocusBorderWidth(context)
	h += 2 * textFieldFocusBorderWidth(context)
	return w, h
}
