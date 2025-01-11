// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Widget interface {
	AppendChildWidgets(context *Context, appender *ChildWidgetAppender)
	HandleInput(context *Context) HandleInputResult
	Update(context *Context) error
	CursorShape(context *Context) (ebiten.CursorShapeType, bool)
	Draw(context *Context, dst *ebiten.Image)
	IsPopup() bool
	Size(context *Context) (int, int)

	widgetState() *widgetState
}

type EventPropagator interface {
	PropagateEvent(context *Context, event Event) (Event, bool)
}

type HandleInputResult struct {
	widget  Widget
	aborted bool
}

func HandleInputByWidget(widget Widget) HandleInputResult {
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

func Parent(widget Widget) Widget {
	return widget.widgetState().parent
}

type DefaultWidget struct {
	widgetState_ widgetState
}

func (*DefaultWidget) AppendChildWidgets(context *Context, appender *ChildWidgetAppender) {
}

func (*DefaultWidget) HandleInput(context *Context) HandleInputResult {
	return HandleInputResult{}
}

func (*DefaultWidget) Update(context *Context) error {
	return nil
}

func (*DefaultWidget) CursorShape(context *Context) (ebiten.CursorShapeType, bool) {
	return 0, false
}

func (*DefaultWidget) Draw(context *Context, dst *ebiten.Image) {
}

func (*DefaultWidget) IsPopup() bool {
	return false
}

func (*DefaultWidget) Size(context *Context) (int, int) {
	return int(16 * context.Scale()), int(16 * context.Scale())
}

func (d *DefaultWidget) widgetState() *widgetState {
	return &d.widgetState_
}

type RootWidget struct {
	DefaultWidget
}

func (*RootWidget) Size(context *Context) (int, int) {
	bounds := context.app.bounds()
	return bounds.Dx(), bounds.Dy()
}
