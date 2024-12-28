// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type WidgetBehavior interface {
	AppendChildWidgets(context *Context, widget *Widget, appender *ChildWidgetAppender)
	HandleInput(context *Context, widget *Widget) HandleInputResult
	Update(context *Context, widget *Widget) error
	CursorShape(context *Context, widget *Widget) (ebiten.CursorShapeType, bool)
	ContentSize(context *Context, widget *Widget) (int, int)

	private()
}

type EventPropagator interface {
	PropagateEvent(context *Context, widget *Widget, event Event) (Event, bool)
}

type Drawer interface {
	Draw(context *Context, widget *Widget, dst *ebiten.Image)
}

type HandleInputResult struct {
	widget  *Widget
	aborted bool
}

func HandleInputByWidget(widget *Widget) HandleInputResult {
	return HandleInputResult{
		widget: widget,
	}
}

func AbortHandlingInput() HandleInputResult {
	return HandleInputResult{
		aborted: true,
	}
}

func (r *HandleInputResult) ShouldRaise() bool {
	return r.widget != nil || r.aborted
}

type DefaultWidgetBehavior struct {
}

func (DefaultWidgetBehavior) AppendChildWidgets(context *Context, widget *Widget, appender *ChildWidgetAppender) {
}

func (DefaultWidgetBehavior) HandleInput(context *Context, widget *Widget) HandleInputResult {
	return HandleInputResult{}
}

func (DefaultWidgetBehavior) Update(context *Context, widget *Widget) error {
	return nil
}

func (DefaultWidgetBehavior) CursorShape(context *Context, widget *Widget) (ebiten.CursorShapeType, bool) {
	return 0, false
}

func (DefaultWidgetBehavior) ContentSize(context *Context, widget *Widget) (int, int) {
	return widget.Bounds().Dx(), widget.Bounds().Dy()
}

func (DefaultWidgetBehavior) private() {
}

func NewWidgetWithPadding(behavior WidgetBehavior, paddingLeft, paddingRight, paddingTop, paddingBottom int) *Widget {
	return NewWidget(&widgetBehaviorWithPadding{
		content:       NewWidget(behavior),
		paddingLeft:   paddingLeft,
		paddingRight:  paddingRight,
		paddingTop:    paddingTop,
		paddingBottom: paddingBottom,
	})
}

type widgetBehaviorWithPadding struct {
	DefaultWidgetBehavior

	content       *Widget
	paddingLeft   int
	paddingRight  int
	paddingTop    int
	paddingBottom int
}

func (w *widgetBehaviorWithPadding) AppendChildWidgets(context *Context, widget *Widget, appender *ChildWidgetAppender) {
	if w.content != nil {
		b := widget.Bounds()
		b.Min.X += w.paddingLeft
		b.Max.X -= w.paddingRight
		b.Min.Y += w.paddingTop
		b.Max.Y -= w.paddingBottom
		appender.AppendChildWidget(w.content, b)
	}
}

func (w *widgetBehaviorWithPadding) ContentSize(context *Context, widget *Widget) (int, int) {
	if w.content == nil {
		return widget.Bounds().Dx(), widget.Bounds().Dy()
	}
	width, height := w.content.ContentSize(context)
	return width + w.paddingLeft + w.paddingRight, height + w.paddingTop + w.paddingBottom
}
