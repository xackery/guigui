// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type MouseEventHandler struct {
	DefaultWidget

	hovering bool
	pressing bool
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
	m.setHovering(image.Pt(x, y).In(VisibleBounds(m)) && IsVisible(m))

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if !image.Pt(ebiten.CursorPosition()).In(VisibleBounds(m)) {
			return HandleInputResult{}
		}
		if IsEnabled(m) {
			m.setPressing(true)
		}
		Focus(m)
		return HandleInputByWidget(m)
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) && m.pressing {
		m.setPressing(false)
		if !image.Pt(ebiten.CursorPosition()).In(VisibleBounds(m)) {
			return HandleInputResult{}
		}
		if IsEnabled(m) {
			return HandleInputByWidget(m)
		}
	}

	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		m.setPressing(false)
	}

	return HandleInputResult{}
}

func (m *MouseEventHandler) Update(context *Context) error {
	if !IsVisible(m) {
		m.setHovering(false)
	}
	return nil
}

func (m *MouseEventHandler) Size(context *Context) (int, int) {
	return Parent(m).Size(context)
}

func (m *MouseEventHandler) setPressing(pressing bool) {
	if m.pressing == pressing {
		return
	}

	if IsEnabled(m) {
		if pressing {
			x, y := ebiten.CursorPosition()
			EnqueueEvent(m, MouseEvent{
				Type:            MouseEventTypeDown,
				MouseButton:     ebiten.MouseButtonLeft,
				CursorPositionX: x,
				CursorPositionY: y,
			})
		} else {
			x, y := ebiten.CursorPosition()
			EnqueueEvent(m, MouseEvent{
				Type:            MouseEventTypeUp,
				MouseButton:     ebiten.MouseButtonLeft,
				CursorPositionX: x,
				CursorPositionY: y,
			})
		}
	}

	m.pressing = pressing
	RequestRedraw(m)
}

func (m *MouseEventHandler) setHovering(hovering bool) {
	if m.hovering == hovering {
		return
	}

	if IsEnabled(m) {
		x, y := ebiten.CursorPosition()
		if hovering {
			EnqueueEvent(m, MouseEvent{
				Type:            MouseEventTypeEnter,
				CursorPositionX: x,
				CursorPositionY: y,
			})
		} else {
			EnqueueEvent(m, MouseEvent{
				Type:            MouseEventTypeLeave,
				CursorPositionX: x,
				CursorPositionY: y,
			})
		}
	}

	m.hovering = hovering
	RequestRedraw(m)
}

func (m *MouseEventHandler) IsPressing() bool {
	return m.pressing
}

func (m *MouseEventHandler) IsHovering() bool {
	return m.hovering
}
