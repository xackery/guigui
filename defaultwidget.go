// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package guigui

import "github.com/hajimehoshi/ebiten/v2"

type DefaultWidget struct {
	widgetState_ widgetState
}

func (*DefaultWidget) Layout(context *Context, appender *ChildWidgetAppender) {
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

func (d *DefaultWidget) Size(context *Context) (int, int) {
	if d.widgetState_.parent == nil {
		return context.app.bounds().Dx(), context.app.bounds().Dy()
	}
	return d.widgetState_.parent.Size(context)
}

func (d *DefaultWidget) widgetState() *widgetState {
	return &d.widgetState_
}
