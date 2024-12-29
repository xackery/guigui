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

	borderInvisible bool
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
	appender.AppendChildWidget(b.mouseEventHandlerWidget, b.buttonBounds(context, widget))
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

/*func (b *Button) SetContent(content view.Widget) {
	if b.content != nil {
		if view.NodeFromWidget(b.content) == view.NodeFromWidget(content) {
			return
		}
		b.RemoveChild(b.content)
	}
	b.content = content
	b.AddChild(b.content, view.LayoutFunc(func(args view.WidgetArgs) image.Rectangle {
		bounds := args.Bounds
		bounds.Min.X += b.settings.SmallUnitSize(args.Scale)
		bounds.Max.X -= b.settings.SmallUnitSize(args.Scale)
		bounds.Min.Y += b.settings.SmallUnitSize(args.Scale)
		bounds.Max.Y -= b.settings.SmallUnitSize(args.Scale)
		return bounds
	}))
}*/

/*func (b *Button) setBorderVisibility(visibility bool) {
	if b.borderVisibility == visibility {
		return
	}

	b.borderVisibility = visibility
	b.Invalidate()
}*/

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

	bounds := b.buttonBounds(context, widget)
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

func (b *Button) buttonBounds(context *guigui.Context, widget *guigui.Widget) image.Rectangle {
	bounds := widget.Bounds()
	bounds.Max.Y = bounds.Min.Y + buttonHeight(context)
	return bounds
}

func buttonHeight(context *guigui.Context) int {
	return UnitSize(context)
}

func (b *Button) ContentSize(context *guigui.Context, widget *guigui.Widget) (int, int) {
	return widget.Bounds().Dx(), buttonHeight(context)
}

type TextButton struct {
	guigui.DefaultWidgetBehavior

	button       Button
	buttonWidget *guigui.Widget
	text         Text
	textWidget   *guigui.Widget

	textColor color.Color

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
	bounds := t.button.buttonBounds(context, widget)

	if t.buttonWidget == nil {
		t.buttonWidget = guigui.NewWidget(&t.button)
	}
	appender.AppendChildWidget(t.buttonWidget, bounds)

	if t.textWidget == nil {
		t.text.SetHorizontalAlign(HorizontalAlignCenter)
		t.text.SetVerticalAlign(VerticalAlignMiddle)
		t.textWidget = guigui.NewWidget(&t.text)
	}
	if t.button.isActive() {
		// As the text is centered, shift it down by double sizes of the stroke width.
		bounds.Min.Y += int(2 * context.Scale())
	} else if !t.buttonWidget.IsEnabled() {
		bounds.Min.Y += int(1 * context.Scale())
	}
	appender.AppendChildWidget(t.textWidget, bounds)
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

/*func (t *TextButton) MinimumWidth(scale float64) int {
	return t.label.Width(scale) + 4*t.button.settings.SmallUnitSize(scale)
}

func (t *TextButton) MinimumHeight(scale float64) int {
	return t.label.Height(scale)
}*/
