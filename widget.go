// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import (
	"image"
	"iter"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Widget struct {
	app_ *app

	behavior      WidgetBehavior
	popup         bool
	bounds        image.Rectangle
	visibleBounds image.Rectangle

	parent   *Widget
	children []*Widget

	hidden       bool
	disabled     bool
	transparency float64

	mightNeedRedraw bool
	origState       state
	redrawBounds    image.Rectangle

	eventQueue EventQueue

	offscreen *ebiten.Image
}

func NewWidget(behavior WidgetBehavior) *Widget {
	return &Widget{
		behavior: behavior,
	}
}

func NewPopupWidget(behavior WidgetBehavior) *Widget {
	return &Widget{
		behavior: behavior,
		popup:    true,
	}
}

func (w *Widget) Behavior() WidgetBehavior {
	return w.behavior
}

func (w *Widget) Parent() *Widget {
	return w.parent
}

func (w *Widget) Bounds() image.Rectangle {
	return w.bounds
}

func (w *Widget) VisibleBounds() image.Rectangle {
	return w.visibleBounds
}

func (w *Widget) EnqueueEvent(event Event) {
	w.eventQueue.Enqueue(event)
}

func (w *Widget) DequeueEvents() iter.Seq[Event] {
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
		focused:      w.IsFocused(),
	}
}

func (w *Widget) Show() {
	if !w.hidden {
		return
	}
	oldState := w.currentState()
	w.hidden = false
	w.requestRedrawIfNeeded(oldState, w.visibleBounds)
}

func (w *Widget) Hide() {
	if w.hidden {
		return
	}
	oldState := w.currentState()
	w.hidden = true
	w.Blur()
	w.requestRedrawIfNeeded(oldState, w.visibleBounds)
}

func (w *Widget) IsVisible() bool {
	if w.parent != nil {
		return !w.hidden && w.parent.IsVisible()
	}
	return !w.hidden
}

func (w *Widget) Enable() {
	if !w.disabled {
		return
	}
	oldState := w.currentState()
	w.disabled = false
	w.requestRedrawIfNeeded(oldState, w.visibleBounds)
}

func (w *Widget) Disable() {
	if w.disabled {
		return
	}
	oldState := w.currentState()
	w.disabled = true
	w.Blur()
	w.requestRedrawIfNeeded(oldState, w.visibleBounds)
}

func (w *Widget) IsEnabled() bool {
	if w.parent != nil {
		return !w.disabled && w.parent.IsEnabled()
	}
	return !w.disabled
}

func (w *Widget) Focus() {
	if !w.IsVisible() {
		return
	}
	if !w.IsEnabled() {
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

func (w *Widget) Blur() {
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

func (w *Widget) IsFocused() bool {
	a := w.app()
	return a != nil && a.focusedWidget == w && w.IsVisible()
}

func (w *Widget) HasFocusedChildWidget() bool {
	if w.IsFocused() {
		return true
	}
	for _, child := range w.children {
		if child.HasFocusedChildWidget() {
			return true
		}
	}
	return false
}

func (w *Widget) Opacity() float64 {
	return 1 - w.transparency
}

func (w *Widget) SetOpacity(opacity float64) {
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

func (w *Widget) RequestRedraw() {
	w.requestRedraw(w.visibleBounds)
	for _, child := range w.children {
		child.requestRedrawIfPopup()
	}
}

func (w *Widget) requestRedrawIfPopup() {
	if w.popup {
		w.requestRedraw(w.visibleBounds)
	}
	for _, child := range w.children {
		child.requestRedrawIfPopup()
	}
}

func (w *Widget) requestRedraw(region image.Rectangle) {
	w.requestRedrawIfNeeded(state{
		nan: math.NaN(),
	}, region)
}

func (w *Widget) requestRedrawIfNeeded(oldState state, region image.Rectangle) {
	newState := w.currentState()
	if oldState == newState {
		return
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

func (w *Widget) ContentSize(context *Context) (int, int) {
	return w.behavior.ContentSize(context, w)
}
