// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import (
	"image"
	"math"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

type app struct {
	root    *Widget
	context *Context

	invalidated image.Rectangle

	prevWidgets    map[*Widget]struct{}
	currentWidgets map[*Widget]struct{}

	screenWidth  float64
	screenHeight float64

	lastScreenWidth  float64
	lastScreenHeight float64
	lastScale        float64

	focusedWidget *Widget
}

type RunOptions struct {
	Title string
}

func Run(root *Widget, options *RunOptions) error {
	if options == nil {
		options = &RunOptions{}
	}

	ebiten.SetWindowTitle(options.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetScreenClearedEveryFrame(false)
	a := &app{
		root: root,
	}
	a.root.app_ = a
	a.context = &Context{
		app: a,
	}
	return ebiten.RunGame(a)
}

func (a app) bounds() image.Rectangle {
	return image.Rect(0, 0, int(math.Ceil(a.screenWidth)), int(math.Ceil(a.screenHeight)))
}

func (a *app) Update() error {
	a.root.bounds = a.bounds()
	a.root.visibleBounds = a.bounds()

	a.context.setDeviceScale(ebiten.Monitor().DeviceScaleFactor())

	// AppendChildWidgets
	a.appendChildWidgets()

	// HandleInput
	// TODO: Handle this in Ebitengine's HandleInput in the future (hajimehoshi/ebiten#1704)
	a.handleInputWidget(a.root)

	if !a.cursorShape(a.root) {
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
	}

	a.propagateEvents(a.root)

	// Update
	if err := a.updateWidget(a.root); err != nil {
		return err
	}

	clearEventQueues(a.root)

	var invalidated bool
	if a.lastScreenWidth != a.screenWidth {
		invalidated = true
		a.lastScreenWidth = a.screenWidth
	}
	if a.lastScreenHeight != a.screenHeight {
		invalidated = true
		a.lastScreenHeight = a.screenHeight
	}
	if s := ebiten.Monitor().DeviceScaleFactor(); a.lastScale != s {
		invalidated = true
		a.lastScale = s
	}

	if invalidated {
		a.requestRedraw(a.bounds())
	}

	return nil
}

func (a *app) Draw(screen *ebiten.Image) {
	if a.invalidated.Empty() {
		return
	}

	dst := screen.SubImage(a.invalidated).(*ebiten.Image)
	a.drawWidget(dst, a.root)

	a.invalidated = image.Rectangle{}
}

func (a *app) Layout(outsideWidth, outsideHeight int) (int, int) {
	panic("gui: game.Layout should never be called")
}

func (a *app) LayoutF(outsideWidth, outsideHeight float64) (float64, float64) {
	s := ebiten.Monitor().DeviceScaleFactor()
	a.screenWidth = outsideWidth * s
	a.screenHeight = outsideHeight * s
	return a.screenWidth, a.screenHeight
}

func (a *app) requestRedraw(region image.Rectangle) {
	a.invalidated = a.invalidated.Union(region)
}

type ChildWidgetAppender struct {
	app    *app
	widget *Widget
}

type WidgetType int

const (
	WidgetTypeRegular WidgetType = iota
	WidgetTypePopup
)

func (c *ChildWidgetAppender) AppendChildWidget(widget *Widget, bounds image.Rectangle) {
	if _, ok := c.app.currentWidgets[widget]; ok {
		panic("guigui: the widget is already in the widget tree")
	}
	if c.app.currentWidgets == nil {
		c.app.currentWidgets = map[*Widget]struct{}{}
	}
	c.app.currentWidgets[widget] = struct{}{}

	// Redraw if the child is a new one, or the bounds are changed.
	if _, ok := widget.behavior.(Drawer); ok {
		if _, ok := c.app.prevWidgets[widget]; !ok {
			if widget.popup {
				c.app.requestRedraw(bounds)
			} else {
				c.app.requestRedraw(bounds.Intersect(c.widget.visibleBounds))
			}
		} else if !widget.bounds.Eq(bounds) {
			if widget.popup {
				c.app.requestRedraw(bounds)
				c.app.requestRedraw(widget.bounds)
			} else {
				c.app.requestRedraw(bounds.Intersect(c.widget.visibleBounds))
				c.app.requestRedraw(widget.bounds.Intersect(c.widget.visibleBounds))
			}
		}
	}

	widget.parent = c.widget
	widget.bounds = bounds
	if widget.popup {
		widget.visibleBounds = widget.bounds
	} else {
		widget.visibleBounds = c.widget.visibleBounds.Intersect(widget.bounds)
	}

	c.widget.children = append(c.widget.children, widget)
}

func (a *app) appendChildWidgets() {
	for w := range a.currentWidgets {
		if a.prevWidgets == nil {
			a.prevWidgets = map[*Widget]struct{}{}
		}
		w.parent = nil
		a.prevWidgets[w] = struct{}{}
	}
	clear(a.currentWidgets)

	a.doAppendChildWidgets(a.root)

	// If the previous children are not in the current children, redraw the region.
	for w := range a.prevWidgets {
		if _, ok := w.behavior.(Drawer); !ok {
			continue
		}
		if _, ok := a.currentWidgets[w]; !ok {
			a.requestRedraw(w.bounds)
		}
	}

	clear(a.prevWidgets)
}

func (a *app) doAppendChildWidgets(widget *Widget) {
	widget.children = slices.Delete(widget.children, 0, len(widget.children))
	widget.behavior.AppendChildWidgets(a.context, widget, &ChildWidgetAppender{
		app:    a,
		widget: widget,
	})

	for _, child := range widget.children {
		a.doAppendChildWidgets(child)
	}
}

func (a *app) handleInputWidget(widget *Widget) HandleInputResult {
	if widget.hidden {
		return HandleInputResult{}
	}

	// Iterate the children in the reverse order of rendering.
	for i := len(widget.children) - 1; i >= 0; i-- {
		child := widget.children[i]
		if r := a.handleInputWidget(child); r.ShouldRaise() {
			return r
		}
	}

	return widget.behavior.HandleInput(a.context, widget)
}

func (a *app) cursorShape(widget *Widget) bool {
	if widget.hidden {
		return false
	}

	// Iterate the children in the reverse order of rendering.
	for i := len(widget.children) - 1; i >= 0; i-- {
		child := widget.children[i]
		if a.cursorShape(child) {
			return true
		}
	}

	if !image.Pt(ebiten.CursorPosition()).In(widget.visibleBounds) {
		return false
	}
	c, ok := widget.behavior.(CursorShaper)
	if !ok {
		return false
	}
	shape, ok := c.CursorShape(a.context, widget)
	if !ok {
		return false
	}
	ebiten.SetCursorShape(shape)
	return true
}

func (a *app) propagateEvents(widget *Widget) {
	for _, child := range widget.children {
		a.propagateEvents(child)
	}

	w, ok := widget.behavior.(EventPropagator)
	if !ok {
		return
	}

	for _, child := range widget.children {
		for ev := range child.DequeueEvents() {
			ev, ok = w.PropagateEvent(a.context, child, ev)
			if !ok {
				continue
			}
			widget.EnqueueEvent(ev)
		}
	}
}

func (a *app) updateWidget(widget *Widget) error {
	if err := widget.behavior.Update(a.context, widget); err != nil {
		return err
	}

	for _, child := range widget.children {
		if err := a.updateWidget(child); err != nil {
			return err
		}
	}

	if widget.mightNeedRedraw {
		b := widget.visibleBounds.Union(widget.redrawBounds)
		a.requestRedraw(b)
		widget.mightNeedRedraw = false
		widget.redrawBounds = image.Rectangle{}
	}

	return nil
}

func clearEventQueues(widget *Widget) {
	widget.eventQueue.Clear()
	for _, child := range widget.children {
		clearEventQueues(child)
	}
}

func (a *app) drawWidget(dst *ebiten.Image, widget *Widget) {
	if widget.visibleBounds.Empty() {
		return
	}

	if widget.hidden {
		return
	}
	if widget.Opacity() == 0 {
		return
	}

	drawer, canDraw := widget.behavior.(Drawer)

	var origDst *ebiten.Image
	if canDraw {
		if widget.Opacity() < 1 {
			origDst = dst
			dst = widget.ensureOffscreen(dst.Bounds())
			dst.Clear()
		}
		drawer.Draw(a.context, widget, dst.SubImage(widget.visibleBounds).(*ebiten.Image))
	}

	for _, child := range widget.children {
		a.drawWidget(dst, child)
	}

	if canDraw {
		if widget.Opacity() < 1 {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(dst.Bounds().Min.X), float64(dst.Bounds().Min.Y))
			op.ColorScale.ScaleAlpha(float32(widget.Opacity()))
			origDst.DrawImage(dst, op)
		}
	}
}
