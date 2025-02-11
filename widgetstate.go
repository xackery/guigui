// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type widgetsAndBounds struct {
	bounds map[Widget]image.Rectangle
}

func (w *widgetsAndBounds) reset() {
	clear(w.bounds)
}

func (w *widgetsAndBounds) append(widget Widget, bounds image.Rectangle) {
	if w.bounds == nil {
		w.bounds = map[Widget]image.Rectangle{}
	}
	w.bounds[widget] = bounds
}

func (w *widgetsAndBounds) equals(currentWidgets []Widget) bool {
	if len(w.bounds) != len(currentWidgets) {
		return false
	}
	for _, widget := range currentWidgets {
		b, ok := w.bounds[widget]
		if !ok {
			return false
		}
		if b != Bounds(widget) {
			return false
		}
	}
	return true
}

func (w *widgetsAndBounds) redrawPopupRegions() {
	for widget := range w.bounds {
		if widget.IsPopup() {
			RequestRedraw(widget)
		}
	}
}

type widgetState struct {
	root bool

	position image.Point

	parent   Widget
	children []Widget
	prev     widgetsAndBounds

	hidden       bool
	disabled     bool
	transparency float64

	offscreen *ebiten.Image
}

func Position(widget Widget) image.Point {
	return widget.widgetState().position
}

func SetPosition(widget Widget, position image.Point) {
	widget.widgetState().position = position
	// Rerendering happens at a.addInvalidatedRegions if necessary.
}

func Bounds(widget Widget) image.Rectangle {
	widgetState := widget.widgetState()
	width, height := widget.Size(&theApp.context)
	return image.Rectangle{
		Min: widgetState.position,
		Max: widgetState.position.Add(image.Point{width, height}),
	}
}

func VisibleBounds(widget Widget) image.Rectangle {
	if widget.IsPopup() {
		return Bounds(widget)
	}
	parent := widget.widgetState().parent
	if parent == nil {
		return theApp.bounds()
	}
	return VisibleBounds(parent).Intersect(Bounds(widget))
}

func (w *widgetState) isInTree() bool {
	p := w
	for ; p.parent != nil; p = p.parent.widgetState() {
	}
	return p.root
}

func Show(widget Widget) {
	widgetState := widget.widgetState()
	if !widgetState.hidden {
		return
	}
	widgetState.hidden = false
	RequestRedraw(widget)
}

func Hide(widget Widget) {
	widgetState := widget.widgetState()
	if widgetState.hidden {
		return
	}
	widgetState.hidden = true
	Blur(widget)
	RequestRedraw(widget)
}

func IsVisible(widget Widget) bool {
	return widget.widgetState().isVisible()
}

func (w *widgetState) isVisible() bool {
	if w.parent != nil {
		return !w.hidden && IsVisible(w.parent)
	}
	return !w.hidden
}

func Enable(widget Widget) {
	widgetState := widget.widgetState()
	if !widgetState.disabled {
		return
	}
	widgetState.disabled = false
	RequestRedraw(widget)
}

func Disable(widget Widget) {
	widgetState := widget.widgetState()
	if widgetState.disabled {
		return
	}
	widgetState.disabled = true
	Blur(widget)
	RequestRedraw(widget)
}

func IsEnabled(widget Widget) bool {
	return widget.widgetState().isEnabled()
}

func (w *widgetState) isEnabled() bool {
	if w.parent != nil {
		return !w.disabled && IsEnabled(w.parent)
	}
	return !w.disabled
}

func Focus(widget Widget) {
	widgetState := widget.widgetState()
	if !widgetState.isVisible() {
		return
	}
	if !widgetState.isEnabled() {
		return
	}

	if !widgetState.isInTree() {
		return
	}
	if theApp.focusedWidget == widget {
		return
	}

	var oldWidget Widget
	if theApp.focusedWidget != nil {
		oldWidget = theApp.focusedWidget
	}

	theApp.focusedWidget = widget
	RequestRedraw(theApp.focusedWidget)
	if oldWidget != nil {
		RequestRedraw(oldWidget)
	}
}

func Blur(widget Widget) {
	widgetState := widget.widgetState()
	if !widgetState.isInTree() {
		return
	}
	if theApp.focusedWidget != widget {
		return
	}
	theApp.focusedWidget = nil
	RequestRedraw(widget)
}

func IsFocused(widget Widget) bool {
	widgetState := widget.widgetState()
	return widgetState.isInTree() && theApp.focusedWidget == widget && widgetState.isVisible()
}

func HasFocusedChildWidget(widget Widget) bool {
	widgetState := widget.widgetState()
	if IsFocused(widget) {
		return true
	}
	for _, child := range widgetState.children {
		if HasFocusedChildWidget(child) {
			return true
		}
	}
	return false
}

func Opacity(widget Widget) float64 {
	return widget.widgetState().opacity()
}

func (w *widgetState) opacity() float64 {
	return 1 - w.transparency
}

func SetOpacity(widget Widget, opacity float64) {
	if opacity < 0 {
		opacity = 0
	}
	if opacity > 1 {
		opacity = 1
	}
	widgetState := widget.widgetState()
	if widgetState.transparency == 1-opacity {
		return
	}
	widgetState.transparency = 1 - opacity
	RequestRedraw(widget)
}

func RequestRedraw(widget Widget) {
	theApp.requestRedrawWidget(widget)
}

func (w *widgetState) ensureOffscreen(bounds image.Rectangle) *ebiten.Image {
	if w.offscreen != nil {
		if !w.offscreen.Bounds().In(bounds) {
			w.offscreen.Deallocate()
			w.offscreen = nil
		}
	}
	if w.offscreen == nil {
		w.offscreen = ebiten.NewImage(bounds.Max.X, bounds.Max.Y)
	}
	return w.offscreen.SubImage(bounds).(*ebiten.Image)
}

func traverseWidget(widget Widget, f func(widget Widget, push bool)) {
	f(widget, true)
	for _, child := range widget.widgetState().children {
		traverseWidget(child, f)
	}
	f(widget, false)
}
