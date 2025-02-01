// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type MouseOverlay struct {
	DefaultWidget

	hovering      bool
	pressingLeft  bool
	pressingRight bool
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

func (m *MouseOverlay) HandleInput(context *Context) HandleInputResult {
	x, y := ebiten.CursorPosition()
	m.setHovering(image.Pt(x, y).In(VisibleBounds(m)) && IsVisible(m))

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		if !image.Pt(ebiten.CursorPosition()).In(VisibleBounds(m)) {
			return HandleInputResult{}
		}
		if IsEnabled(m) {
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				m.setPressing(true, ebiten.MouseButtonLeft)
			}
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
				m.setPressing(true, ebiten.MouseButtonRight)
			}
		}
		Focus(m)
		return HandleInputByWidget(m)
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) && m.pressingLeft ||
		inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight) && m.pressingRight {
		if m.pressingLeft {
			m.setPressing(false, ebiten.MouseButtonLeft)
		}
		if m.pressingRight {
			m.setPressing(false, ebiten.MouseButtonRight)
		}
		if !image.Pt(ebiten.CursorPosition()).In(VisibleBounds(m)) {
			return HandleInputResult{}
		}
		if IsEnabled(m) {
			return HandleInputByWidget(m)
		}
	}

	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		m.setPressing(false, ebiten.MouseButtonLeft)
	}
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		m.setPressing(false, ebiten.MouseButtonRight)
	}

	return HandleInputResult{}
}

func (m *MouseOverlay) Update(context *Context) error {
	if !IsVisible(m) {
		m.setHovering(false)
	}
	return nil
}

func (m *MouseOverlay) Size(context *Context) (int, int) {
	return Parent(m).Size(context)
}

func (m *MouseOverlay) setPressing(pressing bool, mouseButton ebiten.MouseButton) {
	switch mouseButton {
	case ebiten.MouseButtonLeft:
		if m.pressingLeft == pressing {
			return
		}
	case ebiten.MouseButtonRight:
		if m.pressingRight == pressing {
			return
		}
	}

	if IsEnabled(m) {
		if pressing {
			x, y := ebiten.CursorPosition()
			EnqueueEvent(m, MouseEvent{
				Type:            MouseEventTypeDown,
				MouseButton:     mouseButton,
				CursorPositionX: x,
				CursorPositionY: y,
			})
		} else {
			x, y := ebiten.CursorPosition()
			EnqueueEvent(m, MouseEvent{
				Type:            MouseEventTypeUp,
				MouseButton:     mouseButton,
				CursorPositionX: x,
				CursorPositionY: y,
			})
		}
	}

	switch mouseButton {
	case ebiten.MouseButtonLeft:
		m.pressingLeft = pressing
	case ebiten.MouseButtonRight:
		m.pressingRight = pressing
	}
	RequestRedraw(m)
}

func (m *MouseOverlay) setHovering(hovering bool) {
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

func (m *MouseOverlay) IsPressing() bool {
	return m.pressingLeft || m.pressingRight
}

func (m *MouseOverlay) IsHovering() bool {
	return m.hovering
}
