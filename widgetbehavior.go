// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type WidgetBehavior interface {
	AppendChildWidgets(context *Context, appender *ChildWidgetAppender)
	HandleInput(context *Context) HandleInputResult
	Update(context *Context) error
	CursorShape(context *Context) (ebiten.CursorShapeType, bool)
	Draw(context *Context, dst *ebiten.Image)
	IsPopup() bool
	Size(context *Context) (int, int)

	internalWidget(behavior WidgetBehavior) *Widget
}

type EventPropagator interface {
	PropagateEvent(context *Context, event Event) (Event, bool)
}

type HandleInputResult struct {
	widget  *Widget
	aborted bool
}

func HandleInputByWidget(widgetBehavior WidgetBehavior) HandleInputResult {
	return HandleInputResult{
		widget: widgetBehavior.internalWidget(widgetBehavior),
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

func (*DefaultWidgetBehavior) AppendChildWidgets(context *Context, appender *ChildWidgetAppender) {
}

func (*DefaultWidgetBehavior) HandleInput(context *Context) HandleInputResult {
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

func (d *DefaultWidgetBehavior) internalWidget(behavior WidgetBehavior) *Widget {
	// The argument might not match with d.
	if d.widget.behavior != nil && d.widget.behavior != behavior {
		panic("guigui: internalWidget must be called with the same behavior")
	}
	d.widget.behavior = behavior
	return &d.widget
}

type RootWidgetBehavior struct {
	DefaultWidgetBehavior
}

func (*RootWidgetBehavior) Size(context *Context) (int, int) {
	bounds := context.app.bounds()
	return bounds.Dx(), bounds.Dy()
}
