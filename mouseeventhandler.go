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

func (m *MouseEventHandler) HandleInput(context *Context, widget *Widget) HandleInputResult {
	x, y := ebiten.CursorPosition()
	m.setHovering(image.Pt(x, y).In(widget.VisibleBounds()) && widget.IsVisible(), context)

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if !image.Pt(ebiten.CursorPosition()).In(widget.VisibleBounds()) {
			return HandleInputResult{}
		}
		if widget.IsEnabled() {
			m.setPressing(true, widget)
		}
		widget.Focus()
		return HandleInputByWidget(widget)
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) && m.pressing {
		m.setPressing(false, widget)
		if !image.Pt(ebiten.CursorPosition()).In(widget.VisibleBounds()) {
			return HandleInputResult{}
		}
		if widget.IsEnabled() {
			return HandleInputByWidget(widget)
		}
	}

	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		m.setPressing(false, widget)
	}

	return HandleInputResult{}
}

func (m *MouseEventHandler) Update(context *Context) error {
	if m.needsRedraw {
		context.WidgetFromBehavior(m).RequestRedraw()
		m.needsRedraw = false
	}
	if !context.WidgetFromBehavior(m).IsVisible() {
		m.setHovering(false, context)
	}
	return nil
}

func (m *MouseEventHandler) Size(context *Context) (int, int) {
	return context.WidgetFromBehavior(m).Parent().Size(context)
}

func (m *MouseEventHandler) setPressing(pressing bool, widget *Widget) {
	if m.pressing == pressing {
		return
	}

	if widget.IsEnabled() {
		if pressing {
			x, y := ebiten.CursorPosition()
			widget.EnqueueEvent(MouseEvent{
				Type:            MouseEventTypeDown,
				MouseButton:     ebiten.MouseButtonLeft,
				CursorPositionX: x,
				CursorPositionY: y,
			})
		} else {
			x, y := ebiten.CursorPosition()
			widget.EnqueueEvent(MouseEvent{
				Type:            MouseEventTypeUp,
				MouseButton:     ebiten.MouseButtonLeft,
				CursorPositionX: x,
				CursorPositionY: y,
			})
		}
	}

	m.pressing = pressing
	widget.RequestRedraw()
}

func (m *MouseEventHandler) setHovering(hovering bool, context *Context) {
	if m.hovering == hovering {
		return
	}

	if context.WidgetFromBehavior(m).IsEnabled() {
		x, y := ebiten.CursorPosition()
		if hovering {
			context.WidgetFromBehavior(m).EnqueueEvent(MouseEvent{
				Type:            MouseEventTypeEnter,
				CursorPositionX: x,
				CursorPositionY: y,
			})
		} else {
			context.WidgetFromBehavior(m).EnqueueEvent(MouseEvent{
				Type:            MouseEventTypeLeave,
				CursorPositionX: x,
				CursorPositionY: y,
			})
		}
	}

	m.hovering = hovering
	context.WidgetFromBehavior(m).RequestRedraw()
}

func (m *MouseEventHandler) IsPressing() bool {
	return m.pressing
}

func (m *MouseEventHandler) IsHovering() bool {
	return m.hovering
}
