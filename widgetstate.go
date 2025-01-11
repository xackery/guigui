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
	bounds map[*widgetState]image.Rectangle
}

func (w *widgetsAndBounds) reset() {
	clear(w.bounds)
}

func (w *widgetsAndBounds) append(widget *widgetState, bounds image.Rectangle) {
	if w.bounds == nil {
		w.bounds = map[*widgetState]image.Rectangle{}
	}
	w.bounds[widget] = bounds
}

func (w *widgetsAndBounds) equals(currentWidgets []*widgetState) bool {
	if len(w.bounds) != len(currentWidgets) {
		return false
	}
	for _, widget := range currentWidgets {
		b, ok := w.bounds[widget]
		if !ok {
			return false
		}
		if b != widget.bounds() {
			return false
		}
	}
	return true
}

func (w *widgetsAndBounds) redrawPopupRegions() {
	for widget, bounds := range w.bounds {
		if widget.widget.IsPopup() {
			widget.requestRedrawWithRegion(bounds)
		}
	}
}

type widgetState struct {
	app_ *app

	widget        Widget
	position      image.Point
	visibleBounds image.Rectangle

	parent   *widgetState
	children []*widgetState
	prev     widgetsAndBounds

	hidden       bool
	disabled     bool
	transparency float64

	mightNeedRedraw bool
	origState       state
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

func (w *widgetState) bounds() image.Rectangle {
	width, height := w.widget.Size(w.app().context)
	return image.Rectangle{
		Min: w.position,
		Max: w.position.Add(image.Point{width, height}),
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

type state struct {
	hidden       bool
	disabled     bool
	transparency float64
	focused      bool

	// nan is NaN if you want to make this state never equal to any other states.
	nan float64
}

func (w *widgetState) app() *app {
	p := w
	for ; p.parent != nil; p = p.parent {
	}
	return p.app_
}

func (w *widgetState) currentState() state {
	return state{
		hidden:       w.hidden,
		disabled:     w.disabled,
		transparency: w.transparency,
		focused:      w.isFocused(),
	}
}

func Show(widget Widget) {
	widget.widgetState(widget).show()
}

func (w *widgetState) show() {
	if !w.hidden {
		return
	}
	oldState := w.currentState()
	w.hidden = false
	w.requestRedrawIfNeeded(oldState, w.visibleBounds)
}

func Hide(widget Widget) {
	widget.widgetState(widget).hide()
}

func (w *widgetState) hide() {
	if w.hidden {
		return
	}
	oldState := w.currentState()
	w.hidden = true
	w.blur()
	w.requestRedrawIfNeeded(oldState, w.visibleBounds)
}

func IsVisible(widget Widget) bool {
	return widget.widgetState(widget).isVisible()
}

func (w *widgetState) isVisible() bool {
	if w.parent != nil {
		return !w.hidden && w.parent.isVisible()
	}
	return !w.hidden
}

func Enable(widget Widget) {
	widget.widgetState(widget).enable()
}

func (w *widgetState) enable() {
	if !w.disabled {
		return
	}
	oldState := w.currentState()
	w.disabled = false
	w.requestRedrawIfNeeded(oldState, w.visibleBounds)
}

func Disable(widget Widget) {
	widget.widgetState(widget).disable()
}

func (w *widgetState) disable() {
	if w.disabled {
		return
	}
	oldState := w.currentState()
	w.disabled = true
	w.blur()
	w.requestRedrawIfNeeded(oldState, w.visibleBounds)
}

func IsEnabled(widget Widget) bool {
	return widget.widgetState(widget).isEnabled()
}

func (w *widgetState) isEnabled() bool {
	if w.parent != nil {
		return !w.disabled && w.parent.isEnabled()
	}
	return !w.disabled
}

func Focus(widget Widget) {
	widget.widgetState(widget).focus()
}

func (w *widgetState) focus() {
	if !w.isVisible() {
		return
	}
	if !w.isEnabled() {
		return
	}

	a := w.app()
	if a == nil {
		return
	}
	if a.focusedWidget == w {
		return
	}

	var oldWidget *widgetState
	if a.focusedWidget != nil {
		oldWidget = a.focusedWidget
	}

	newWidgetOldState := w.currentState()
	var oldWidgetOldState state
	if oldWidget != nil {
		oldWidgetOldState = oldWidget.currentState()
	}

	a.focusedWidget = w
	a.focusedWidget.requestRedrawIfNeeded(newWidgetOldState, a.focusedWidget.visibleBounds)
	if oldWidget != nil {
		oldWidget.requestRedrawIfNeeded(oldWidgetOldState, oldWidget.visibleBounds)
	}
}

func Blur(widget Widget) {
	widget.widgetState(widget).blur()
}

func (w *widgetState) blur() {
	a := w.app()
	if a == nil {
		return
	}
	if a.focusedWidget != w {
		return
	}
	oldState := w.currentState()
	a.focusedWidget = nil
	w.requestRedrawIfNeeded(oldState, w.visibleBounds)
}

func IsFocused(widget Widget) bool {
	return widget.widgetState(widget).isFocused()
}

func (w *widgetState) isFocused() bool {
	a := w.app()
	return a != nil && a.focusedWidget == w && w.isVisible()
}

func HasFocusedChildWidget(widget Widget) bool {
	return widget.widgetState(widget).hasFocusedChildWidget()
}

func (w *widgetState) hasFocusedChildWidget() bool {
	if w.isFocused() {
		return true
	}
	for _, child := range w.children {
		if child.hasFocusedChildWidget() {
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
	widget.widgetState(widget).setOpacity(opacity)
}

func (w *widgetState) setOpacity(opacity float64) {
	if 1-w.transparency == opacity {
		return
	}
	oldState := w.currentState()
	if opacity < 0 {
		opacity = 0
	}
	if opacity > 1 {
		opacity = 1
	}
	w.transparency = 1 - opacity
	w.requestRedrawIfNeeded(oldState, w.visibleBounds)
}

func RequestRedraw(widget Widget) {
	widget.widgetState(widget).requestRedraw()
}

func (w *widgetState) requestRedraw() {
	w.requestRedrawWithRegion(w.visibleBounds)
	for _, child := range w.children {
		child.requestRedrawIfPopup()
	}
}

func (w *widgetState) requestRedrawIfPopup() {
	if w.widget.IsPopup() {
		w.requestRedrawWithRegion(w.visibleBounds)
	}
	for _, child := range w.children {
		child.requestRedrawIfPopup()
	}
}

func (w *widgetState) requestRedrawWithRegion(region image.Rectangle) {
	w.requestRedrawIfNeeded(state{
		nan: math.NaN(),
	}, region)
}

func (w *widgetState) requestRedrawIfNeeded(oldState state, region image.Rectangle) {
	if region.Empty() {
		return
	}

	newState := w.currentState()
	if oldState == newState {
		return
	}

	if theDebugMode.showRenderingRegions {
		slog.Info("Request redrawing", "requester", fmt.Sprintf("%T", w.widget), "region", region)
	}

	w.redrawBounds = w.redrawBounds.Union(region)

	if !w.mightNeedRedraw {
		w.mightNeedRedraw = true
		w.origState = oldState
		return
	}

	if w.origState == newState {
		w.mightNeedRedraw = false
		w.origState = state{}
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
