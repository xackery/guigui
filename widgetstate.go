// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import (
	"fmt"
	"image"
	"iter"
	"log/slog"
	"math"

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
		if b != bounds(widget) {
			return false
		}
	}
	return true
}

func (w *widgetsAndBounds) redrawPopupRegions() {
	for widget, bounds := range w.bounds {
		if widget.IsPopup() {
			requestRedrawWithRegion(widget, bounds)
		}
	}
}

type widgetState struct {
	app_ *app

	position      image.Point
	visibleBounds image.Rectangle

	parent   Widget
	children []Widget
	prev     widgetsAndBounds

	hidden       bool
	disabled     bool
	transparency float64

	mightNeedRedraw bool
	origState       stateForRedraw
	redrawBounds    image.Rectangle

	eventQueue EventQueue

	offscreen *ebiten.Image
}

func Position(widget Widget) image.Point {
	return widget.widgetState(widget).position
}

func SetPosition(widget Widget, position image.Point) {
	widget.widgetState(widget).position = position
	// Rerendering happens at a.addInvalidatedRegions if necessary.
}

func bounds(widget Widget) image.Rectangle {
	widgetState := widget.widgetState(widget)
	width, height := widget.Size(widgetState.app().context)
	return image.Rectangle{
		Min: widgetState.position,
		Max: widgetState.position.Add(image.Point{width, height}),
	}
}

func VisibleBounds(widget Widget) image.Rectangle {
	return widget.widgetState(widget).visibleBounds
}

func EnqueueEvent(widget Widget, event Event) {
	widget.widgetState(widget).enqueueEvent(event)
}

func (w *widgetState) enqueueEvent(event Event) {
	w.eventQueue.Enqueue(event)
}

func DequeueEvents(widget Widget) iter.Seq[Event] {
	return widget.widgetState(widget).dequeueEvents()
}

func (w *widgetState) dequeueEvents() iter.Seq[Event] {
	return func(yield func(event Event) bool) {
		for {
			e, ok := w.eventQueue.Dequeue()
			if !ok {
				break
			}
			if !yield(e) {
				break
			}
		}
	}
}

type stateForRedraw struct {
	hidden       bool
	disabled     bool
	transparency float64
	focused      bool

	// nan is NaN if you want to make this state never equal to any other states.
	nan float64
}

func (w *widgetState) app() *app {
	p := w
	for ; p.parent != nil; p = p.parent.widgetState(p.parent) {
	}
	return p.app_
}

func widgetStateForRedraw(widget Widget) stateForRedraw {
	widgetState := widget.widgetState(widget)
	return stateForRedraw{
		hidden:       widgetState.hidden,
		disabled:     widgetState.disabled,
		transparency: widgetState.transparency,
		focused:      IsFocused(widget),
	}
}

func Show(widget Widget) {
	widgetState := widget.widgetState(widget)
	if !widgetState.hidden {
		return
	}
	oldState := widgetStateForRedraw(widget)
	widgetState.hidden = false
	requestRedrawIfNeeded(widget, oldState, widgetState.visibleBounds)
}

func Hide(widget Widget) {
	widgetState := widget.widgetState(widget)
	if widgetState.hidden {
		return
	}
	oldState := widgetStateForRedraw(widget)
	widgetState.hidden = true
	Blur(widget)
	requestRedrawIfNeeded(widget, oldState, widgetState.visibleBounds)
}

func IsVisible(widget Widget) bool {
	return widget.widgetState(widget).isVisible()
}

func (w *widgetState) isVisible() bool {
	if w.parent != nil {
		return !w.hidden && IsVisible(w.parent)
	}
	return !w.hidden
}

func Enable(widget Widget) {
	widgetState := widget.widgetState(widget)
	if !widgetState.disabled {
		return
	}
	oldState := widgetStateForRedraw(widget)
	widgetState.disabled = false
	requestRedrawIfNeeded(widget, oldState, widgetState.visibleBounds)
}

func Disable(widget Widget) {
	widgetState := widget.widgetState(widget)
	if widgetState.disabled {
		return
	}
	oldState := widgetStateForRedraw(widget)
	widgetState.disabled = true
	Blur(widget)
	requestRedrawIfNeeded(widget, oldState, widgetState.visibleBounds)
}

func IsEnabled(widget Widget) bool {
	return widget.widgetState(widget).isEnabled()
}

func (w *widgetState) isEnabled() bool {
	if w.parent != nil {
		return !w.disabled && IsEnabled(w.parent)
	}
	return !w.disabled
}

func Focus(widget Widget) {
	widgetState := widget.widgetState(widget)
	if !widgetState.isVisible() {
		return
	}
	if !widgetState.isEnabled() {
		return
	}

	a := widgetState.app()
	if a == nil {
		return
	}
	if a.focusedWidget == widget {
		return
	}

	var oldWidget Widget
	if a.focusedWidget != nil {
		oldWidget = a.focusedWidget
	}

	newWidgetOldState := widgetStateForRedraw(widget)
	var oldWidgetOldState stateForRedraw
	if oldWidget != nil {
		oldWidgetOldState = widgetStateForRedraw(oldWidget)
	}

	a.focusedWidget = widget
	requestRedrawIfNeeded(a.focusedWidget, newWidgetOldState, a.focusedWidget.widgetState(a.focusedWidget).visibleBounds)
	if oldWidget != nil {
		requestRedrawIfNeeded(oldWidget, oldWidgetOldState, oldWidget.widgetState(oldWidget).visibleBounds)
	}
}

func Blur(widget Widget) {
	widgetState := widget.widgetState(widget)
	a := widgetState.app()
	if a == nil {
		return
	}
	if a.focusedWidget != widget {
		return
	}
	oldState := widgetStateForRedraw(widget)
	a.focusedWidget = nil
	requestRedrawIfNeeded(widget, oldState, widgetState.visibleBounds)
}

func IsFocused(widget Widget) bool {
	widgetState := widget.widgetState(widget)
	a := widgetState.app()
	return a != nil && a.focusedWidget == widget && widgetState.isVisible()
}

func HasFocusedChildWidget(widget Widget) bool {
	widgetState := widget.widgetState(widget)
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
	return widget.widgetState(widget).opacity()
}

func (w *widgetState) opacity() float64 {
	return 1 - w.transparency
}

func SetOpacity(widget Widget, opacity float64) {
	widgetState := widget.widgetState(widget)
	if 1-widgetState.transparency == opacity {
		return
	}
	oldState := widgetStateForRedraw(widget)
	if opacity < 0 {
		opacity = 0
	}
	if opacity > 1 {
		opacity = 1
	}
	widgetState.transparency = 1 - opacity
	requestRedrawIfNeeded(widget, oldState, widgetState.visibleBounds)
}

func RequestRedraw(widget Widget) {
	widgetState := widget.widgetState(widget)
	requestRedrawWithRegion(widget, widgetState.visibleBounds)
	for _, child := range widgetState.children {
		requestRedrawIfPopup(child)
	}
}

func requestRedrawIfPopup(widget Widget) {
	widgetState := widget.widgetState(widget)
	if widget.IsPopup() {
		requestRedrawWithRegion(widget, widgetState.visibleBounds)
	}
	for _, child := range widgetState.children {
		requestRedrawIfPopup(child)
	}
}

func requestRedrawWithRegion(widget Widget, region image.Rectangle) {
	requestRedrawIfNeeded(widget, stateForRedraw{
		nan: math.NaN(),
	}, region)
}

func requestRedrawIfNeeded(widget Widget, oldState stateForRedraw, region image.Rectangle) {
	if region.Empty() {
		return
	}

	newState := widgetStateForRedraw(widget)
	if oldState == newState {
		return
	}

	if theDebugMode.showRenderingRegions {
		slog.Info("Request redrawing", "requester", fmt.Sprintf("%T", widget), "region", region)
	}

	widgetState := widget.widgetState(widget)
	widgetState.redrawBounds = widgetState.redrawBounds.Union(region)

	if !widgetState.mightNeedRedraw {
		widgetState.mightNeedRedraw = true
		widgetState.origState = oldState
		return
	}

	if widgetState.origState == newState {
		widgetState.mightNeedRedraw = false
		widgetState.origState = stateForRedraw{}
		return
	}
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
