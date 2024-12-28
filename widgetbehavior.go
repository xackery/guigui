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

func (DefaultWidgetBehavior) private() {
}
