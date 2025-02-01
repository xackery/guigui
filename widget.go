// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Widget interface {
	Layout(context *Context, appender *ChildWidgetAppender)
	HandleInput(context *Context) HandleInputResult
	Update(context *Context) error
	CursorShape(context *Context) (ebiten.CursorShapeType, bool)
	Draw(context *Context, dst *ebiten.Image)
	IsPopup() bool
	Size(context *Context) (int, int)

	widgetState() *widgetState
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

type RootWidget struct {
	DefaultWidget
}

func (*RootWidget) Size(context *Context) (int, int) {
	bounds := context.app.bounds()
	return bounds.Dx(), bounds.Dy()
}
