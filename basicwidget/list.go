// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package basicwidget

import (
	"image"
	"image/color"
	"slices"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/hajimehoshi/guigui"
)

type ListStyle int

const (
	ListStyleNormal ListStyle = iota
	ListStyleSidebar
	ListStyleMenu
	ListStyleDropdownMenu
)

type ListItem struct {
	Content    *guigui.Widget
	Selectable bool
	Wide       bool
	Draggable  bool
	Tag        any
}

func DefaultActiveListItemTextColor(context *guigui.Context) color.Color {
	return Color2(context.ColorMode(), ColorTypeBase, 1, 1)
}

type List struct {
	guigui.DefaultWidgetBehavior

	listFrameWidget       *guigui.Widget
	scrollOverlayWidget   *guigui.Widget
	dragDropOverlayWidget *guigui.Widget

	items                 []ListItem
	selectedItemIndex     int
	hoveredItemIndex      int
	showItemBorders       bool
	style                 ListStyle
	lastSelectingItemTime time.Time

	indexToJump        int
	dropSrcIndex       int
	dropDstIndex       int
	pressStartX        int
	pressStartY        int
	startPressingIndex int
	startPressingLeft  bool

	widthMinusDefault  int
	heightMinusDefault int

	needsRedraw bool
}

/*l := &List{
	selectedItemIndex:  -1,
	hoveredItemIndex:   -1,
	indexToJump:        -1,
	dropSrcIndex:       -1,
	dropDstIndex:       -1,
	startPressingIndex: -1,
}*/

func listItemPadding(context *guigui.Context) int {
	return UnitSize(context) / 4
}

func (l *List) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	if l.style != ListStyleSidebar {
		if l.listFrameWidget == nil {
			l.listFrameWidget = guigui.NewWidget(&listFrame{})
		}
		appender.AppendChildWidget(l.listFrameWidget, widget.Position())
	}

	p := widget.Position()
	p.X += RoundedCornerRadius(context) + listItemPadding(context)
	p.Y += RoundedCornerRadius(context)
	for _, item := range l.items {
		/*r := l.list.itemRect(args, l.index)
		if l.list.items[l.index].Wide {
			r.Min.X -= l.list.settings.SmallUnitSize(args.Scale)
			r.Max.X += l.list.settings.SmallUnitSize(args.Scale)
		}
		return r*/
		appender.AppendChildWidget(item.Content, p)
		_, h := item.Content.Size(context)
		p.Y += h
	}

	if l.scrollOverlayWidget == nil {
		l.scrollOverlayWidget = guigui.NewWidget(&ScrollOverlay{})
	}
	appender.AppendChildWidget(l.scrollOverlayWidget, widget.Position())

	if l.dragDropOverlayWidget == nil {
		l.dragDropOverlayWidget = guigui.NewWidget(&DragDropOverlay{})
	}
	appender.AppendChildWidget(l.dragDropOverlayWidget, widget.Position())
}

func (l *List) SelectedItem() (ListItem, bool) {
	if l.selectedItemIndex < 0 || l.selectedItemIndex >= len(l.items) {
		return ListItem{}, false
	}
	return l.items[l.selectedItemIndex], true
}

func (l *List) ItemAt(index int) (ListItem, bool) {
	if index < 0 || index >= len(l.items) {
		return ListItem{}, false
	}
	return l.items[index], true
}

func (l *List) SelectedItemIndex() int {
	return l.selectedItemIndex
}

func (l *List) HoveredItemIndex() int {
	return l.hoveredItemIndex
}

func (l *List) SetItems(items []ListItem) {
	l.items = make([]ListItem, len(items))
	copy(l.items, items)
}

func (l *List) SetItem(item ListItem, index int) {
	l.items[index] = item
}

func (l *List) AddItem(item ListItem, index int) {
	l.items = slices.Insert(l.items, index, item)
	// TODO: Send an event.
}

func (l *List) RemoveItem(index int) {
	l.items = slices.Delete(l.items, index, index+1)
	// TODO: Send an event.
}

func (l *List) MoveItem(from int, to int) {
	moveItemInSlice(l.items, from, 1, to)
	// TODO: Send an event.
}

func (l *List) SetSelectedItemIndex(index int) {
	if index < 0 || index >= len(l.items) {
		index = -1
	}
	changed := l.selectedItemIndex != index
	l.selectedItemIndex = index
	if changed {
		l.needsRedraw = true
	}
	/*if index >= 0 && l.callback != nil && l.callback.OnItemSelected != nil {
		l.callback.OnItemSelected(index)
	}*/
}

func (l *List) JumpToItemIndex(index int) {
	if index < 0 || index >= len(l.items) {
		return
	}
	l.indexToJump = index
}

func (l *List) setHoveredItemIndex(index int) {
	if l.hoveredItemIndex == index {
		return
	}

	if index < 0 || index >= len(l.items) {
		index = -1
	}
	l.hoveredItemIndex = index
	l.needsRedraw = true
}

func (l *List) ShowItemBorders(show bool) {
	if l.showItemBorders == show {
		return
	}
	l.showItemBorders = true
	l.needsRedraw = true
}

func (l *List) IsHoveringVisible() bool {
	return l.style == ListStyleMenu || l.style == ListStyleDropdownMenu
}

func (l *List) SetStyle(style ListStyle) {
	if l.style == style {
		return
	}
	l.style = style
	l.needsRedraw = true
}

func (l *List) calcDropDstIndex(context *guigui.Context, widget *guigui.Widget) int {
	_, y := ebiten.CursorPosition()
	for i := range l.items {
		if r := l.itemRect(context, widget, i); y < (r.Min.Y+r.Max.Y)/2 {
			return i
		}
	}
	return len(l.items)
}

func (l *List) HandleInput(context *guigui.Context, widget *guigui.Widget) guigui.HandleInputResult {
	// Process dragging.
	if l.dragDropOverlayWidget.Behavior().(*DragDropOverlay).IsDragging() {
		_, y := ebiten.CursorPosition()
		p := widget.Position()
		_, h := widget.Size(context)
		var dy float64
		if upperY := p.Y + UnitSize(context); y < upperY {
			dy = float64(upperY-y) / 4
		}
		if lowerY := p.Y + h - UnitSize(context); y >= lowerY {
			dy = float64(lowerY-y) / 4
		}
		l.scrollOverlayWidget.Behavior().(*ScrollOverlay).SetOffsetByDelta(0, dy)
		i := l.calcDropDstIndex(context, widget)
		if l.dropDstIndex != i {
			l.dropDstIndex = i
			widget.RequestRedraw()
		}
		return guigui.HandleInputByWidget(widget)
	}

	// Process dropping.
	var dropped bool
	if l.dropSrcIndex >= 0 && l.dropDstIndex >= 0 {
		dropped = true
		/*if l.callback != nil && l.callback.OnItemDropped != nil {
			l.callback.OnItemDropped(l.dropSrcIndex, l.dropDstIndex)
		}*/
	}

	l.dropSrcIndex = -1
	if l.dropDstIndex != -1 {
		l.dropDstIndex = -1
		widget.RequestRedraw()
	}

	if dropped {
		return guigui.HandleInputByWidget(widget)
	}

	if x, y := ebiten.CursorPosition(); image.Pt(x, y).In(widget.VisibleBounds()) {
		_, offsetY := l.scrollOverlayWidget.Behavior().(*ScrollOverlay).Offset()
		y -= RoundedCornerRadius(context)
		y -= widget.Position().Y
		y -= int(offsetY)
		index := -1
		var cy int
		for i, item := range l.items {
			_, h := item.Content.Size(context)
			if cy <= y && y < cy+h {
				index = i
				break
			}
			cy += h
		}
		l.setHoveredItemIndex(index)
		if index >= 0 && index < len(l.items) {
			left := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
			right := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)

			switch {
			case left || right:
				if !l.items[index].Selectable {
					return guigui.HandleInputByWidget(widget)
				}

				wasFocused := widget.IsFocused()
				widget.Focus()
				if l.selectedItemIndex != index || !wasFocused {
					l.SetSelectedItemIndex(index)
					l.lastSelectingItemTime = time.Now()
				}
				l.pressStartX = x
				l.pressStartY = y
				if right {
					/*if l.callback != nil && l.callback.OnContextMenu != nil {
						x, y := ebiten.CursorPosition()
						l.callback.OnContextMenu(index, x, y)
					}*/
				}
				l.startPressingIndex = index
				l.startPressingLeft = left

			case ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft):
				if l.items[index].Draggable && l.selectedItemIndex == index && l.startPressingIndex == index && (l.pressStartX != x || l.pressStartY != y) {
					l.dragDropOverlayWidget.Behavior().(*DragDropOverlay).Start(index)
				}

			case inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft):
				if l.selectedItemIndex == index && l.startPressingLeft && time.Since(l.lastSelectingItemTime) > 400*time.Millisecond {
					/*if l.callback != nil && l.callback.OnItemEditStarted != nil {
						l.callback.OnItemEditStarted(index)
					}*/
				}
				l.pressStartX = 0
				l.pressStartY = 0
				l.startPressingIndex = -1
				l.startPressingLeft = false
			}

			return guigui.HandleInputByWidget(widget)
		}
		l.dropSrcIndex = -1
		l.pressStartX = 0
		l.pressStartY = 0
	} else {
		l.setHoveredItemIndex(-1)
	}

	return guigui.HandleInputResult{}
}

func (l *List) Update(context *guigui.Context, widget *guigui.Widget) error {
	if l.needsRedraw {
		widget.RequestRedraw()
		l.needsRedraw = false
	}

	w, _ := widget.Size(context)
	l.scrollOverlayWidget.Behavior().(*ScrollOverlay).SetContentSize(w, l.ContentHeight(context))

	if l.indexToJump >= 0 {
		y := l.itemYFromIndex(context, l.indexToJump) - RoundedCornerRadius(context)
		l.scrollOverlayWidget.Behavior().(*ScrollOverlay).SetOffset(0, float64(-y))
		l.indexToJump = -1
	}

	return nil
}

func (l *List) ItemWidth(context *guigui.Context, widget *guigui.Widget) int {
	w, _ := widget.Size(context)
	w -= 2 * RoundedCornerRadius(context)
	w -= 2 * listItemPadding(context)
	return w
}

func (l *List) ContentHeight(context *guigui.Context) int {
	var h int
	h += RoundedCornerRadius(context)
	for _, w := range l.items {
		_, wh := w.Content.Size(context)
		h += wh
	}
	h += RoundedCornerRadius(context)
	return h
}

func (l *List) itemYFromIndex(context *guigui.Context, index int) int {
	y := RoundedCornerRadius(context)
	for i, item := range l.items {
		if i == index {
			break
		}
		_, wh := item.Content.Size(context)
		y += wh
	}
	return y
}

func (l *List) itemRect(context *guigui.Context, widget *guigui.Widget, index int) image.Rectangle {
	_, offsetY := l.scrollOverlayWidget.Behavior().(*ScrollOverlay).Offset()
	p := widget.Position()
	w, h := widget.Size(context)
	b := image.Rectangle{
		Min: p,
		Max: p.Add(image.Pt(w, h)),
	}
	padding := listItemPadding(context)
	b.Min.X += RoundedCornerRadius(context) + padding
	b.Max.X -= RoundedCornerRadius(context) + padding
	b.Min.Y += l.itemYFromIndex(context, index)
	b.Min.Y += int(offsetY)
	_, ih := l.items[index].Content.Size(context)
	b.Max.Y = b.Min.Y + ih
	return b
}

func (l *List) selectedItemColor(context *guigui.Context, widget *guigui.Widget) color.Color {
	if l.selectedItemIndex < 0 || l.selectedItemIndex >= len(l.items) {
		return nil
	}
	if widget.IsFocused() || l.style == ListStyleSidebar {
		return Color(context.ColorMode(), ColorTypeAccent, 0.5)
	}
	return Color(context.ColorMode(), ColorTypeBase, 0.8)
}

func (l *List) Draw(context *guigui.Context, widget *guigui.Widget, dst *ebiten.Image) {
	if l.style != ListStyleSidebar {
		clr := Color(context.ColorMode(), ColorTypeBase, 1)
		if l.style == ListStyleMenu {
			clr = Color(context.ColorMode(), ColorTypeBase, 0.8)
		}
		p := widget.Position()
		w, h := widget.Size(context)
		bounds := image.Rectangle{
			Min: p,
			Max: p.Add(image.Pt(w, h)),
		}
		DrawRoundedRect(context, dst, bounds, clr, RoundedCornerRadius(context))
	}

	// Draw item borders.
	if l.showItemBorders && len(l.items) > 0 {
		_, offsetY := l.scrollOverlayWidget.Behavior().(*ScrollOverlay).Offset()
		p := widget.Position()
		w, _ := widget.Size(context)
		y := float32(p.Y) + float32(RoundedCornerRadius(context)) + float32(offsetY)
		for i, item := range l.items {
			_, ih := item.Content.Size(context)
			y += float32(ih)
			if i == l.selectedItemIndex || i+1 == l.selectedItemIndex {
				continue
			}
			if i == len(l.items)-1 {
				continue
			}
			x0 := p.X + RoundedCornerRadius(context)
			x1 := p.X + w - RoundedCornerRadius(context)
			width := 1 * float32(context.Scale())
			clr := Color(context.ColorMode(), ColorTypeBase, 0.5)
			vector.StrokeLine(dst, float32(x0), y, float32(x1), y, width, clr, false)
		}
	}

	if clr := l.selectedItemColor(context, widget); clr != nil && l.selectedItemIndex >= 0 && l.selectedItemIndex < len(l.items) {
		r := l.itemRect(context, widget, l.selectedItemIndex)
		r.Min.X -= RoundedCornerRadius(context)
		r.Max.X += RoundedCornerRadius(context)
		if r.Overlaps(widget.VisibleBounds()) {
			DrawRoundedRect(context, dst, r, clr, RoundedCornerRadius(context)/2)
		}
	}

	if l.IsHoveringVisible() && l.hoveredItemIndex >= 0 && l.hoveredItemIndex < len(l.items) && l.items[l.hoveredItemIndex].Selectable {
		r := l.itemRect(context, widget, l.hoveredItemIndex)
		r.Min.X -= RoundedCornerRadius(context)
		r.Max.X += RoundedCornerRadius(context)
		if r.Overlaps(widget.VisibleBounds()) {
			DrawRoundedRect(context, dst, r, Color(context.ColorMode(), ColorTypeBase, 0.9), RoundedCornerRadius(context)/2)
		}
	}

	// Draw a drag indicator.
	/*if l.hoveredItemIndex >= 0 && l.hoveredItemIndex < len(l.items) && l.items[l.hoveredItemIndex].Draggable && !l.dragDropOverlayWidget.Behavior().(*DragDropOverlay).IsDragging() {
		img := resource.Image("dragindicator", l.settings.Theme().UIForegroundColor)
		op := &ebiten.DrawImageOptions{}
		s := float64(2*RoundedCornerRadius(context)) / float64(img.Bounds().Dy())
		op.GeoM.Scale(s, s)
		r := l.itemRect(context, widget, l.hoveredItemIndex)
		op.GeoM.Translate(float64(r.Min.X-2*RoundedCornerRadius(context)), float64(r.Min.Y)+(float64(r.Dy())-float64(img.Bounds().Dy())*s)/2)
		dst.DrawImage(img, op)
	}*/

	// Draw a dragging guideline.
	if l.dropDstIndex >= 0 {
		p := widget.Position()
		w, _ := widget.Size(context)
		x0 := float32(p.X) + float32(RoundedCornerRadius(context))
		x1 := float32(p.X+w) - float32(RoundedCornerRadius(context))
		y := float32(p.Y)
		y += float32(l.itemYFromIndex(context, l.dropDstIndex))
		_, offsetY := l.scrollOverlayWidget.Behavior().(*ScrollOverlay).Offset()
		y += float32(offsetY)
		vector.StrokeLine(dst, x0, y, x1, y, 2*float32(context.Scale()), Color(context.ColorMode(), ColorTypeBase, 0.1), false)
	}
}

func (l *List) onDrop(data any) {
	l.dropSrcIndex = data.(int)
}

func defaultListSize(context *guigui.Context) (int, int) {
	return 6 * UnitSize(context), 6 * UnitSize(context)
}

func (l *List) Size(context *guigui.Context, widget *guigui.Widget) (int, int) {
	dw, dh := defaultListSize(context)
	return l.widthMinusDefault + dw, l.heightMinusDefault + dh
}

func (l *List) SetSize(context *guigui.Context, width, height int) {
	dw, dh := defaultListSize(context)
	l.widthMinusDefault = width - dw
	l.heightMinusDefault = height - dh
}

type listFrame struct {
	guigui.DefaultWidgetBehavior
}

func (l *listFrame) Draw(context *guigui.Context, widget *guigui.Widget, dst *ebiten.Image) {
	border := RoundedRectBorderTypeInset
	if widget.Parent().Behavior().(*List).style != ListStyleNormal {
		border = RoundedRectBorderTypeOutset
	}
	p := widget.Position()
	w, h := widget.Size(context)
	bounds := image.Rectangle{
		Min: p,
		Max: p.Add(image.Pt(w, h)),
	}
	clr := Color(context.ColorMode(), ColorTypeBase, 0.85)
	borderWidth := float32(1 * context.Scale())
	DrawRoundedRectBorder(context, dst, bounds, clr, RoundedCornerRadius(context), borderWidth, border)
}

func (l *listFrame) Size(context *guigui.Context, widget *guigui.Widget) (int, int) {
	return widget.Parent().Size(context)
}

func moveItemInSlice[T any](slice []T, from int, count int, to int) {
	if count == 0 {
		return
	}
	if from <= to && to <= from+count {
		return
	}
	if from < to {
		to -= count
	}

	s := make([]T, count)
	copy(s, slice[from:from+count])
	slice = slices.Delete(slice, from, from+count)
	// Assume that the slice has enough capacity, then the underlying array should not change.
	_ = slices.Insert(slice, to, s...)
}
