// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type MouseEventHandler struct {
	DefaultWidgetBehavior

	hovering bool
	pressing bool

	needsRedraw bool
}

type MouseEvent struct {
	Type            MouseEventType
	MouseButton     ebiten.MouseButton
	CursorPositionX int
	CursorPositionY int
}

type MouseEventType int

const (
	MouseEventTypeDown MouseEventType = iota
	MouseEventTypeUp
	MouseEventTypeEnter
	MouseEventTypeLeave
)

func (m *MouseEventHandler) HandleInput(context *Context) HandleInputResult {
	x, y := ebiten.CursorPosition()
	m.setHovering(image.Pt(x, y).In(context.Widget(m).VisibleBounds()) && context.Widget(m).IsVisible(), context)

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if !image.Pt(ebiten.CursorPosition()).In(context.Widget(m).VisibleBounds()) {
			return HandleInputResult{}
		}
		if context.Widget(m).IsEnabled() {
			m.setPressing(true, context)
		}
		context.Widget(m).Focus()
		return HandleInputByWidget(m)
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) && m.pressing {
		m.setPressing(false, context)
		if !image.Pt(ebiten.CursorPosition()).In(context.Widget(m).VisibleBounds()) {
			return HandleInputResult{}
		}
		if context.Widget(m).IsEnabled() {
			return HandleInputByWidget(m)
		}
	}

	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		m.setPressing(false, context)
	}

	return HandleInputResult{}
}

func (m *MouseEventHandler) Update(context *Context) error {
	if m.needsRedraw {
		context.Widget(m).RequestRedraw()
		m.needsRedraw = false
	}
	if !context.Widget(m).IsVisible() {
		m.setHovering(false, context)
	}
	return nil
}

func (m *MouseEventHandler) Size(context *Context) (int, int) {
	return Parent(m).Size(context)
}

func (m *MouseEventHandler) setPressing(pressing bool, context *Context) {
	if m.pressing == pressing {
		return
	}

	if context.Widget(m).IsEnabled() {
		if pressing {
			x, y := ebiten.CursorPosition()
			context.Widget(m).EnqueueEvent(MouseEvent{
				Type:            MouseEventTypeDown,
				MouseButton:     ebiten.MouseButtonLeft,
				CursorPositionX: x,
				CursorPositionY: y,
			})
		} else {
			x, y := ebiten.CursorPosition()
			context.Widget(m).EnqueueEvent(MouseEvent{
				Type:            MouseEventTypeUp,
				MouseButton:     ebiten.MouseButtonLeft,
				CursorPositionX: x,
				CursorPositionY: y,
			})
		}
	}

	m.pressing = pressing
	context.Widget(m).RequestRedraw()
}

func (m *MouseEventHandler) setHovering(hovering bool, context *Context) {
	if m.hovering == hovering {
		return
	}

	if context.Widget(m).IsEnabled() {
		x, y := ebiten.CursorPosition()
		if hovering {
			context.Widget(m).EnqueueEvent(MouseEvent{
				Type:            MouseEventTypeEnter,
				CursorPositionX: x,
				CursorPositionY: y,
			})
		} else {
			context.Widget(m).EnqueueEvent(MouseEvent{
				Type:            MouseEventTypeLeave,
				CursorPositionX: x,
				CursorPositionY: y,
			})
		}
	}

	m.hovering = hovering
	context.Widget(m).RequestRedraw()
}

func (m *MouseEventHandler) IsPressing() bool {
	return m.pressing
}

func (m *MouseEventHandler) IsHovering() bool {
	return m.hovering
}
