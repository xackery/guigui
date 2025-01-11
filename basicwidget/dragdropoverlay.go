// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/hajimehoshi/guigui"
)

type DragDropOverlay struct {
	guigui.DefaultWidget

	object any
}

type DragDropOverlayEventType int

const (
	DragDropOverlayEventTypeDropped DragDropOverlayEventType = iota
)

type DragDropOverlayEvent struct {
	Type   DragDropOverlayEventType
	Object any
}

func (d *DragDropOverlay) IsDragging() bool {
	return d.object != nil
}

func (d *DragDropOverlay) Start(object any) {
	d.object = object
}

func (d *DragDropOverlay) HandleInput(context *guigui.Context) guigui.HandleInputResult {
	if d.object != nil {
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			if image.Pt(ebiten.CursorPosition()).In(guigui.VisibleBounds(d)) {
				guigui.EnqueueEvent(d, DragDropOverlayEvent{
					Type:   DragDropOverlayEventTypeDropped,
					Object: d.object,
				})
			}
			d.object = nil
			return guigui.HandleInputResult{}
		}
		if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			d.object = nil
		}
		return guigui.HandleInputResult{}
	}

	return guigui.HandleInputResult{}
}

func (d *DragDropOverlay) Size(context *guigui.Context) (int, int) {
	return guigui.Parent(d).Size(context)
}
