// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type WidgetBehavior interface {
	AppendChildWidgets(context *Context, widget *Widget, appender *ChildWidgetAppender)
	HandleInput(context *Context, widget *Widget) HandleInputResult
	Update(context *Context) error
	CursorShape(context *Context) (ebiten.CursorShapeType, bool)
	Draw(context *Context, dst *ebiten.Image)
	IsPopup() bool
	Size(context *Context) (int, int)

	internalWidget() *Widget
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

// TODO: Add more Widget functions.
type DefaultWidgetBehavior struct {
	widget Widget
}

func (*DefaultWidgetBehavior) AppendChildWidgets(context *Context, widget *Widget, appender *ChildWidgetAppender) {
}

func (*DefaultWidgetBehavior) HandleInput(context *Context, widget *Widget) HandleInputResult {
	return HandleInputResult{}
}

func (*DefaultWidgetBehavior) Update(context *Context) error {
	return nil
}

func (*DefaultWidgetBehavior) CursorShape(context *Context) (ebiten.CursorShapeType, bool) {
	return 0, false
}

func (*DefaultWidgetBehavior) Draw(context *Context, dst *ebiten.Image) {
}

func (*DefaultWidgetBehavior) IsPopup() bool {
	return false
}

func (*DefaultWidgetBehavior) Size(context *Context) (int, int) {
	return int(16 * context.Scale()), int(16 * context.Scale())
}

func (d *DefaultWidgetBehavior) internalWidget() *Widget {
	// d.widget.behavior cannot be set here.
	return &d.widget
}

type RootWidgetBehavior struct {
	DefaultWidgetBehavior
}

func (RootWidgetBehavior) Size(context *Context) (int, int) {
	bounds := context.app.bounds()
	return bounds.Dx(), bounds.Dy()
}
