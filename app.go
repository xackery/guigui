// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import (
	"image"
	"image/color"
	"math"
	"os"
	"slices"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/hajimehoshi/oklab"
)

type debugMode struct {
	showRenderingRegions bool
}

var theDebugMode debugMode

func init() {
	for _, token := range strings.Split(os.Getenv("GUIGUI_DEBUG"), ",") {
		switch token {
		case "showrenderingregions":
			theDebugMode.showRenderingRegions = true
		}
	}
}

type invalidatedRegionsForDebugItem struct {
	region image.Rectangle
	time   int
}

func invalidatedRegionForDebugMaxTime() int {
	return ebiten.TPS() / 5
}

type app struct {
	root    *Widget
	context *Context

	invalidated                image.Rectangle
	invalidatedRegionsForDebug []invalidatedRegionsForDebugItem

	prevWidgets    map[*Widget]image.Rectangle
	currentWidgets map[*Widget]struct{}

	screenWidth  float64
	screenHeight float64

	lastScreenWidth  float64
	lastScreenHeight float64
	lastScale        float64

	focusedWidget *Widget

	debugScreen *ebiten.Image
}

type RunOptions struct {
	Title           string
	WindowMinWidth  int
	WindowMinHeight int
	WindowMaxWidth  int
	WindowMaxHeight int
	AppScale        float64
}

func Run(root WidgetBehavior, options *RunOptions) error {
	if options == nil {
		options = &RunOptions{}
	}

	ebiten.SetWindowTitle(options.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetScreenClearedEveryFrame(false)
	minW := -1
	minH := -1
	maxW := -1
	maxH := -1
	if options.WindowMinWidth > 0 {
		minW = options.WindowMinWidth
	}
	if options.WindowMinHeight > 0 {
		minH = options.WindowMinHeight
	}
	if options.WindowMaxWidth > 0 {
		maxW = options.WindowMaxWidth
	}
	if options.WindowMaxHeight > 0 {
		maxH = options.WindowMaxHeight
	}
	ebiten.SetWindowSizeLimits(minW, minH, maxW, maxH)

	a := &app{
		root: widgetFromBehavior(root),
	}
	a.root.app_ = a
	a.context = &Context{
		app: a,
	}
	if options.AppScale > 0 {
		a.context.appScaleMinus1 = options.AppScale - 1
	}
	eop := &ebiten.RunGameOptions{
		ColorSpace: ebiten.ColorSpaceSRGB,
	}
	return ebiten.RunGameWithOptions(a, eop)
}

func (a app) bounds() image.Rectangle {
	return image.Rect(0, 0, int(math.Ceil(a.screenWidth)), int(math.Ceil(a.screenHeight)))
}

func (a *app) Update() error {
	a.root.position = image.Point{}
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

	if theDebugMode.showRenderingRegions {
		// Update the regions in the reversed order to remove items.
		for idx := len(a.invalidatedRegionsForDebug) - 1; idx >= 0; idx-- {
			if a.invalidatedRegionsForDebug[idx].time > 0 {
				a.invalidatedRegionsForDebug[idx].time--
			} else {
				a.invalidatedRegionsForDebug = slices.Delete(a.invalidatedRegionsForDebug, idx, idx+1)
			}
		}

		if !a.invalidated.Empty() {
			idx := slices.IndexFunc(a.invalidatedRegionsForDebug, func(i invalidatedRegionsForDebugItem) bool {
				return i.region.Eq(a.invalidated)
			})
			if idx < 0 {
				a.invalidatedRegionsForDebug = append(a.invalidatedRegionsForDebug, invalidatedRegionsForDebugItem{
					region: a.invalidated,
					time:   invalidatedRegionForDebugMaxTime(),
				})
			} else {
				a.invalidatedRegionsForDebug[idx].time = invalidatedRegionForDebugMaxTime()
			}
		}
	}

	return nil
}

func (a *app) Draw(screen *ebiten.Image) {
	a.drawWidget(screen)
	a.invalidated = image.Rectangle{}
}

func (a *app) Layout(outsideWidth, outsideHeight int) (int, int) {
	panic("guigui: game.Layout should never be called")
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

type WidgetType int

const (
	WidgetTypeRegular WidgetType = iota
	WidgetTypePopup
)

func (a *app) appendChildWidgets() {
	for w := range a.currentWidgets {
		if a.prevWidgets == nil {
			a.prevWidgets = map[*Widget]image.Rectangle{}
		}
		a.prevWidgets[w] = w.bounds()
		// Do not reset parent here, as bounds might require parent info.
	}
	for w := range a.currentWidgets {
		w.parent = nil
	}
	clear(a.currentWidgets)

	a.doAppendChildWidgets(a.root)

	// If the previous children are not in the current children, redraw the region.
	for w, bounds := range a.prevWidgets {
		if _, ok := a.currentWidgets[w]; !ok {
			// Do not use w.bounds() as w.parent might be nil.
			a.requestRedraw(bounds)
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
	shape, ok := widget.behavior.CursorShape(a.context)
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

func (a *app) drawWidget(screen *ebiten.Image) {
	if !theDebugMode.showRenderingRegions {
		if !a.invalidated.Empty() {
			dst := screen.SubImage(a.invalidated).(*ebiten.Image)
			a.doDrawWidget(dst, a.root)
		}
	} else {
		a.doDrawWidget(screen, a.root)
	}

	if theDebugMode.showRenderingRegions {
		if a.debugScreen != nil {
			if a.debugScreen.Bounds().Dx() != screen.Bounds().Dx() || a.debugScreen.Bounds().Dy() != screen.Bounds().Dy() {
				a.debugScreen.Deallocate()
				a.debugScreen = nil
			}
		}
		if a.debugScreen == nil {
			a.debugScreen = ebiten.NewImage(screen.Bounds().Dx(), screen.Bounds().Dy())
		}

		a.debugScreen.Clear()
		for _, item := range a.invalidatedRegionsForDebug {
			clr := oklab.OklchModel.Convert(color.RGBA{R: 0xff, G: 0x4b, B: 0x00, A: 0xff}).(oklab.Oklch)
			clr.Alpha = float64(item.time) / float64(invalidatedRegionForDebugMaxTime())
			if clr.Alpha > 0 {
				w := float32(4 * a.context.Scale())
				vector.StrokeRect(a.debugScreen, float32(item.region.Min.X)+w/2, float32(item.region.Min.Y)+w/2, float32(item.region.Dx())-w, float32(item.region.Dy())-w, w, clr, false)
			}
		}
		screen.DrawImage(a.debugScreen, nil)
	}
}

func (a *app) doDrawWidget(dst *ebiten.Image, widget *Widget) {
	if widget.visibleBounds.Empty() {
		return
	}

	if widget.hidden {
		return
	}
	if widget.Opacity() == 0 {
		return
	}

	var origDst *ebiten.Image
	if widget.Opacity() < 1 {
		origDst = dst
		dst = widget.ensureOffscreen(dst.Bounds())
		dst.Clear()
	}
	widget.behavior.Draw(a.context, dst.SubImage(widget.visibleBounds).(*ebiten.Image))

	for _, child := range widget.children {
		a.doDrawWidget(dst, child)
	}

	if widget.Opacity() < 1 {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(dst.Bounds().Min.X), float64(dst.Bounds().Min.Y))
		op.ColorScale.ScaleAlpha(float32(widget.Opacity()))
		origDst.DrawImage(dst, op)
	}
}
