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
	Draw(context *Context, widget *Widget, dst *ebiten.Image)
	ContentSize(context *Context, widget *Widget) (int, int)

	private()
}

type EventPropagator interface {
	PropagateEvent(context *Context, widget *Widget, event Event) (Event, bool)
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

func (DefaultWidgetBehavior) Draw(context *Context, widget *Widget, dst *ebiten.Image) {
}

func (DefaultWidgetBehavior) ContentSize(context *Context, widget *Widget) (int, int) {
	return widget.Bounds().Dx(), widget.Bounds().Dy()
}

func (DefaultWidgetBehavior) private() {
}
