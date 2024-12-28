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
	bounds := widget.Bounds()
	bounds.Min.X += UnitSize(context) / 2
	bounds.Max.X -= UnitSize(context) / 2
	// TODO: Consider multiline.
	if !t.text.IsMultiline() {
		t.text.SetVerticalAlign(VerticalAlignMiddle)
	}
	appender.AppendChildWidget(t.textWidget, bounds)

	if widget.HasFocusedChildWidget() {
		if t.focusWidget == nil {
			t.focusWidget = guigui.NewPopupWidget(&textFieldFocus{})
		}
		bounds := widget.Bounds().Inset(-int(3 * context.Scale()))
		appender.AppendChildWidget(t.focusWidget, bounds)
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
	DrawRoundedRect(context, dst, widget.Bounds(), Color(context.ColorMode(), ColorTypeBase, 0.85), RoundedCornerRadius(context))
	DrawRoundedRectBorder(context, dst, widget.Bounds(), Color(context.ColorMode(), ColorTypeBase, 0.8), RoundedCornerRadius(context), float32(1*context.Scale()), RoundedRectBorderTypeInset)
}

type textFieldFocus struct {
	guigui.DefaultWidgetBehavior
}

func (t *textFieldFocus) Draw(context *guigui.Context, widget *guigui.Widget, dst *ebiten.Image) {
	DrawRoundedRectBorder(context, dst, widget.Bounds(), Color(context.ColorMode(), ColorTypeAccent, 0.8), int(4*context.Scale())+RoundedCornerRadius(context), float32(4*context.Scale()), RoundedRectBorderTypeRegular)
}
