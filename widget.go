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
	bounds map[*Widget]image.Rectangle
}

func (w *widgetsAndBounds) reset() {
	clear(w.bounds)
}

func (w *widgetsAndBounds) append(widget *Widget, bounds image.Rectangle) {
	if w.bounds == nil {
		w.bounds = map[*Widget]image.Rectangle{}
	}
	w.bounds[widget] = bounds
}

func (w *widgetsAndBounds) equals(currentWidgets []*Widget) bool {
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
		if widget.behavior.IsPopup() {
			widget.requestRedrawWithRegion(bounds)
		}
	}
}

type Widget struct {
	app_ *app

	behavior      WidgetBehavior
	position      image.Point
	visibleBounds image.Rectangle

	parent   *Widget
	children []*Widget
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

func Position(widgetBehavior WidgetBehavior) image.Point {
	return widgetBehavior.internalWidget(widgetBehavior).position
}

func SetPosition(widgetBehavior WidgetBehavior, position image.Point) {
	widgetBehavior.internalWidget(widgetBehavior).position = position
	// Rerendering happens at a.addInvalidatedRegions if necessary.
}

func (w *Widget) bounds() image.Rectangle {
	width, height := w.behavior.Size(w.app().context)
	return image.Rectangle{
		Min: w.position,
		Max: w.position.Add(image.Point{width, height}),
	}
}

func VisibleBounds(widgetBehavior WidgetBehavior) image.Rectangle {
	return widgetBehavior.internalWidget(widgetBehavior).visibleBounds
}

func EnqueueEvent(widgetBehavior WidgetBehavior, event Event) {
	widgetBehavior.internalWidget(widgetBehavior).enqueueEvent(event)
}

func (w *Widget) enqueueEvent(event Event) {
	w.eventQueue.Enqueue(event)
}

func DequeueEvents(widgetBehavior WidgetBehavior) iter.Seq[Event] {
	return widgetBehavior.internalWidget(widgetBehavior).dequeueEvents()
}

func (w *Widget) dequeueEvents() iter.Seq[Event] {
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

func (w *Widget) app() *app {
	p := w
	for ; p.parent != nil; p = p.parent {
	}
	return p.app_
}

func (w *Widget) currentState() state {
	return state{
		hidden:       w.hidden,
		disabled:     w.disabled,
		transparency: w.transparency,
		focused:      w.isFocused(),
	}
}

func Show(widgetBehavior WidgetBehavior) {
	widgetBehavior.internalWidget(widgetBehavior).show()
}

func (w *Widget) show() {
	if !w.hidden {
		return
	}
	oldState := w.currentState()
	w.hidden = false
	w.requestRedrawIfNeeded(oldState, w.visibleBounds)
}

func Hide(widgetBehavior WidgetBehavior) {
	widgetBehavior.internalWidget(widgetBehavior).hide()
}

func (w *Widget) hide() {
	if w.hidden {
		return
	}
	oldState := w.currentState()
	w.hidden = true
	w.blur()
	w.requestRedrawIfNeeded(oldState, w.visibleBounds)
}

func IsVisible(widgetBehavior WidgetBehavior) bool {
	return widgetBehavior.internalWidget(widgetBehavior).isVisible()
}

func (w *Widget) isVisible() bool {
	if w.parent != nil {
		return !w.hidden && w.parent.isVisible()
	}
	return !w.hidden
}

func Enable(widgetBehavior WidgetBehavior) {
	widgetBehavior.internalWidget(widgetBehavior).enable()
}

func (w *Widget) enable() {
	if !w.disabled {
		return
	}
	oldState := w.currentState()
	w.disabled = false
	w.requestRedrawIfNeeded(oldState, w.visibleBounds)
}

func Disable(widgetBehavior WidgetBehavior) {
	widgetBehavior.internalWidget(widgetBehavior).disable()
}

func (w *Widget) disable() {
	if w.disabled {
		return
	}
	oldState := w.currentState()
	w.disabled = true
	w.blur()
	w.requestRedrawIfNeeded(oldState, w.visibleBounds)
}

func IsEnabled(widgetBehavior WidgetBehavior) bool {
	return widgetBehavior.internalWidget(widgetBehavior).isEnabled()
}

func (w *Widget) isEnabled() bool {
	if w.parent != nil {
		return !w.disabled && w.parent.isEnabled()
	}
	return !w.disabled
}

func Focus(widgetBehavior WidgetBehavior) {
	widgetBehavior.internalWidget(widgetBehavior).focus()
}

func (w *Widget) focus() {
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

	var oldWidget *Widget
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

func Blur(widgetBehavior WidgetBehavior) {
	widgetBehavior.internalWidget(widgetBehavior).blur()
}

func (w *Widget) blur() {
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

func IsFocused(widgetBehavior WidgetBehavior) bool {
	return widgetBehavior.internalWidget(widgetBehavior).isFocused()
}

func (w *Widget) isFocused() bool {
	a := w.app()
	return a != nil && a.focusedWidget == w && w.isVisible()
}

func HasFocusedChildWidget(widgetBehavior WidgetBehavior) bool {
	return widgetBehavior.internalWidget(widgetBehavior).hasFocusedChildWidget()
}

func (w *Widget) hasFocusedChildWidget() bool {
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

func Opacity(widgetBehavior WidgetBehavior) float64 {
	return widgetBehavior.internalWidget(widgetBehavior).opacity()
}

func (w *Widget) opacity() float64 {
	return 1 - w.transparency
}

func SetOpacity(widgetBehavior WidgetBehavior, opacity float64) {
	widgetBehavior.internalWidget(widgetBehavior).setOpacity(opacity)
}

func (w *Widget) setOpacity(opacity float64) {
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

func RequestRedraw(widgetBehavior WidgetBehavior) {
	widgetBehavior.internalWidget(widgetBehavior).requestRedraw()
}

func (w *Widget) requestRedraw() {
	w.requestRedrawWithRegion(w.visibleBounds)
	for _, child := range w.children {
		child.requestRedrawIfPopup()
	}
}

func (w *Widget) requestRedrawIfPopup() {
	if w.behavior.IsPopup() {
		w.requestRedrawWithRegion(w.visibleBounds)
	}
	for _, child := range w.children {
		child.requestRedrawIfPopup()
	}
}

func (w *Widget) requestRedrawWithRegion(region image.Rectangle) {
	w.requestRedrawIfNeeded(state{
		nan: math.NaN(),
	}, region)
}

func (w *Widget) requestRedrawIfNeeded(oldState state, region image.Rectangle) {
	if region.Empty() {
		return
	}

	newState := w.currentState()
	if oldState == newState {
		return
	}

	if theDebugMode.showRenderingRegions {
		slog.Info("Request redrawing", "requester", fmt.Sprintf("%T", w.behavior), "region", region)
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

func (w *Widget) ensureOffscreen(bounds image.Rectangle) *ebiten.Image {
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
