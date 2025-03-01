// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import (
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"maps"
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

var theApp *app

type app struct {
	root      Widget
	context   Context
	visitedZs map[int]struct{}
	zs        []int

	invalidatedRegions image.Rectangle
	invalidatedWidgets []Widget

	invalidatedRegionsForDebug []invalidatedRegionsForDebugItem

	screenWidth  float64
	screenHeight float64

	lastScreenWidth  float64
	lastScreenHeight float64
	lastScale        float64

	focusedWidget Widget

	offscreen   *ebiten.Image
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

func Run(root Widget, options *RunOptions) error {
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
		root: root,
	}
	theApp = a
	a.root.widgetState().root = true
	a.context.app = a
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
	rootState := a.root.widgetState()
	rootState.position = image.Point{}

	a.context.setDeviceScale(ebiten.Monitor().DeviceScaleFactor())

	// AppendChildWidgets
	a.layout()

	// HandleInput
	// TODO: Handle this in Ebitengine's HandleInput in the future (hajimehoshi/ebiten#1704)
	a.handleInputWidget()

	if !a.cursorShape() {
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
	}

	// Update
	if err := a.updateWidget(a.root); err != nil {
		return err
	}

	clearEventQueues(a.root)

	// Invalidate the engire screen if the screen size is changed.
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
	} else {
		// Invalidate regions if a widget's children state is changed.
		// A widget's bounds might be changed in Update, so do this after updating.
		a.requestRedrawIfTreeChanged(a.root)
	}
	a.resetPrevWidgets(a.root)

	// Resolve invalidatedWidgets.
	if len(a.invalidatedWidgets) > 0 {
		for _, widget := range a.invalidatedWidgets {
			vb := VisibleBounds(widget)
			if vb.Empty() {
				continue
			}
			if theDebugMode.showRenderingRegions {
				slog.Info("Request redrawing", "requester", fmt.Sprintf("%T", widget), "region", vb)
			}
			a.invalidatedRegions = a.invalidatedRegions.Union(vb)
		}
		a.invalidatedWidgets = slices.Delete(a.invalidatedWidgets, 0, len(a.invalidatedWidgets))
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

		if !a.invalidatedRegions.Empty() {
			idx := slices.IndexFunc(a.invalidatedRegionsForDebug, func(i invalidatedRegionsForDebugItem) bool {
				return i.region.Eq(a.invalidatedRegions)
			})
			if idx < 0 {
				a.invalidatedRegionsForDebug = append(a.invalidatedRegionsForDebug, invalidatedRegionsForDebugItem{
					region: a.invalidatedRegions,
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
	origScreen := screen
	if theDebugMode.showRenderingRegions {
		if a.offscreen != nil {
			if a.offscreen.Bounds().Dx() != screen.Bounds().Dx() || a.offscreen.Bounds().Dy() != screen.Bounds().Dy() {
				a.offscreen.Deallocate()
				a.offscreen = nil
			}
		}
		if a.offscreen == nil {
			a.offscreen = ebiten.NewImage(screen.Bounds().Dx(), screen.Bounds().Dy())
		}
		screen = a.offscreen
	}
	a.drawWidget(screen)
	a.drawDebugIfNeeded(origScreen)
	a.invalidatedRegions = image.Rectangle{}
	a.invalidatedWidgets = slices.Delete(a.invalidatedWidgets, 0, len(a.invalidatedWidgets))
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
	a.invalidatedRegions = a.invalidatedRegions.Union(region)
}

func (a *app) requestRedrawWidget(widget Widget) {
	a.invalidatedWidgets = append(a.invalidatedWidgets, widget)
	for _, child := range widget.widgetState().children {
		theApp.requestRedrawIfPopup(child)
	}
}

func (a *app) requestRedrawIfPopup(widget Widget) {
	if widget.IsPopup() {
		a.requestRedrawWidget(widget)
		return
	}
	for _, child := range widget.widgetState().children {
		a.requestRedrawIfPopup(child)
	}
}

func (a *app) layout() {
	a.doLayout(a.root)

	// Calculate z values.
	clear(a.visitedZs)
	var z int
	traverseWidget(a.root, func(widget Widget, push bool) {
		if widget.IsPopup() {
			if push {
				z++
			} else {
				z--
			}
		}
		if a.visitedZs == nil {
			a.visitedZs = map[int]struct{}{}
		}
		a.visitedZs[z] = struct{}{}
	})

	a.zs = slices.Delete(a.zs, 0, len(a.zs))
	a.zs = slices.AppendSeq(a.zs, maps.Keys(a.visitedZs))
	slices.Sort(a.zs)
}

func (a *app) doLayout(widget Widget) {
	widgetState := widget.widgetState()
	widgetState.children = slices.Delete(widgetState.children, 0, len(widgetState.children))
	widget.Layout(&a.context, &ChildWidgetAppender{
		app:    a,
		widget: widget,
	})
	for _, child := range widgetState.children {
		a.doLayout(child)
	}
}

func (a *app) handleInputWidget() HandleInputResult {
	var startZ int
	if a.root.IsPopup() {
		startZ = 1
	}
	for i := len(a.zs) - 1; i >= 0; i-- {
		z := a.zs[i]
		if r := a.doHandleInputWidget(a.root, z, startZ); r.ShouldRaise() {
			return r
		}
	}
	return HandleInputResult{}
}

func (a *app) doHandleInputWidget(widget Widget, zToHandle int, currentZ int) HandleInputResult {
	if zToHandle < currentZ {
		return HandleInputResult{}
	}

	widgetState := widget.widgetState()
	if widgetState.hidden {
		return HandleInputResult{}
	}

	// Iterate the children in the reverse order of rendering.
	for i := len(widgetState.children) - 1; i >= 0; i-- {
		child := widgetState.children[i]
		l := currentZ
		if child.IsPopup() {
			l++
		}
		if r := a.doHandleInputWidget(child, zToHandle, l); r.ShouldRaise() {
			return r
		}
	}

	if zToHandle != currentZ {
		return HandleInputResult{}
	}
	return widget.HandleInput(&a.context)
}

func (a *app) cursorShape() bool {
	var startZ int
	if a.root.IsPopup() {
		startZ = 1
	}
	for i := len(a.zs) - 1; i >= 0; i-- {
		z := a.zs[i]
		if a.doCursorShape(a.root, z, startZ) {
			return true
		}
	}
	return false
}

func (a *app) doCursorShape(widget Widget, zToHandle int, currentZ int) bool {
	if zToHandle < currentZ {
		return false
	}

	widgetState := widget.widgetState()
	if widgetState.hidden {
		return false
	}

	// Iterate the children in the reverse order of rendering.
	for i := len(widgetState.children) - 1; i >= 0; i-- {
		child := widgetState.children[i]
		l := currentZ
		if child.IsPopup() {
			l++
		}
		if a.doCursorShape(child, zToHandle, l) {
			return true
		}
	}

	if zToHandle != currentZ {
		return false
	}

	if !image.Pt(ebiten.CursorPosition()).In(VisibleBounds(widget)) {
		return false
	}

	shape, ok := widget.CursorShape(&a.context)
	if !ok {
		return false
	}
	ebiten.SetCursorShape(shape)
	return true
}

func (a *app) updateWidget(widget Widget) error {
	widgetState := widget.widgetState()
	if err := widget.Update(&a.context); err != nil {
		return err
	}

	for _, child := range widgetState.children {
		if err := a.updateWidget(child); err != nil {
			return err
		}
	}

	return nil
}

func clearEventQueues(widget Widget) {
	widgetState := widget.widgetState()
	for _, child := range widgetState.children {
		clearEventQueues(child)
	}
}

func (a *app) requestRedrawIfTreeChanged(widget Widget) {
	widgetState := widget.widgetState()
	// If the children and/or children's bounds are changed, request redraw.
	if !widgetState.prev.equals(widgetState.children) {
		// Popups are outside of widget, so redraw the regions explicitly.
		widgetState.prev.redrawPopupRegions()
		a.requestRedraw(VisibleBounds(widget))
		for _, child := range widgetState.children {
			if child.IsPopup() {
				a.requestRedraw(VisibleBounds(child))
			}
		}
	}
	for _, child := range widgetState.children {
		a.requestRedrawIfTreeChanged(child)
	}
}

func (a *app) resetPrevWidgets(widget Widget) {
	widgetState := widget.widgetState()
	// Reset the states.
	widgetState.prev.reset()
	for _, child := range widgetState.children {
		widgetState.prev.append(child, Bounds(child))
	}
	for _, child := range widgetState.children {
		a.resetPrevWidgets(child)
	}
}

func (a *app) drawWidget(screen *ebiten.Image) {
	if a.invalidatedRegions.Empty() {
		return
	}
	dst := screen.SubImage(a.invalidatedRegions).(*ebiten.Image)
	var startZ int
	if a.root.IsPopup() {
		startZ = 1
	}
	for _, z := range a.zs {
		a.doDrawWidget(dst, a.root, z, startZ)
	}
}

func (a *app) doDrawWidget(dst *ebiten.Image, widget Widget, zToRender int, currentZ int) {
	if zToRender < currentZ {
		return
	}

	vb := VisibleBounds(widget)
	if vb.Empty() {
		return
	}

	widgetState := widget.widgetState()
	if widgetState.hidden {
		return
	}
	if widgetState.opacity() == 0 {
		return
	}

	var origDst *ebiten.Image
	renderCurrent := zToRender == currentZ
	if renderCurrent {
		if widgetState.opacity() < 1 {
			origDst = dst
			dst = widgetState.ensureOffscreen(dst.Bounds())
			dst.Clear()
		}
		widget.Draw(&a.context, dst.SubImage(vb).(*ebiten.Image))
	}

	for _, child := range widgetState.children {
		l := currentZ
		if child.IsPopup() {
			l++
		}
		a.doDrawWidget(dst, child, zToRender, l)
	}

	if renderCurrent {
		if widgetState.opacity() < 1 {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(dst.Bounds().Min.X), float64(dst.Bounds().Min.Y))
			op.ColorScale.ScaleAlpha(float32(widgetState.opacity()))
			origDst.DrawImage(dst, op)
		}
	}
}

func (a *app) drawDebugIfNeeded(screen *ebiten.Image) {
	if !theDebugMode.showRenderingRegions {
		return
	}

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
	op := &ebiten.DrawImageOptions{}
	op.Blend = ebiten.BlendCopy
	screen.DrawImage(a.offscreen, op)
	screen.DrawImage(a.debugScreen, nil)
}
