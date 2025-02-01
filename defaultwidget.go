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

func (*DefaultWidget) Size(context *Context) (int, int) {
	return int(16 * context.Scale()), int(16 * context.Scale())
}

func (d *DefaultWidget) widgetState() *widgetState {
	return &d.widgetState_
}
