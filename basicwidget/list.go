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

	"github.com/xackery/guigui"
)

type ListStyle int

const (
	ListStyleNormal ListStyle = iota
	ListStyleSidebar
	ListStyleMenu
)

type ListItem struct {
	Content    guigui.Widget
	Selectable bool
	Wide       bool
	Draggable  bool
	Tag        any
}

func DefaultActiveListItemTextColor(context *guigui.Context) color.Color {
	return Color2(context.ColorMode(), ColorTypeBase, 1, 1)
}

func DefaultDisabledListItemTextColor(context *guigui.Context) color.Color {
	return Color(context.ColorMode(), ColorTypeBase, 0.5)
}

type List struct {
	guigui.DefaultWidget

	listFrame       listFrame
	scrollOverlay   ScrollOverlay
	dragDropOverlay DragDropOverlay

	items                  []ListItem
	selectedItemIndexPlus1 int
	hoveredItemIndexPlus1  int
	showItemBorders        bool
	style                  ListStyle
	lastSelectingItemTime  time.Time

	indexToJumpPlus1        int
	dropSrcIndexPlus1       int
	dropDstIndexPlus1       int
	pressStartX             int
	pressStartY             int
	startPressingIndexPlus1 int
	startPressingLeft       bool

	widthSet            bool
	heightSet           bool
	width               int
	height              int
	cachedDefaultWidth  int
	cachedDefaultHeight int

	onItemSelected func(index int)
}

func listItemPadding(context *guigui.Context) int {
	return UnitSize(context) / 4
}

func (l *List) SetOnItemSelected(f func(index int)) {
	l.onItemSelected = f
}

func (l *List) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	if l.style != ListStyleSidebar && l.style != ListStyleMenu {
		guigui.SetPosition(&l.listFrame, guigui.Position(l))
		appender.AppendChildWidget(&l.listFrame)
	}

	_, offsetY := l.scrollOverlay.Offset()
	p := guigui.Position(l)
	p.X += RoundedCornerRadius(context) + listItemPadding(context)
	p.Y += RoundedCornerRadius(context) + int(offsetY)
	for _, item := range l.items {
		/*r := l.list.itemRect(args, l.index)
		if l.list.items[l.index].Wide {
			r.Min.X -= l.list.settings.SmallUnitSize(args.Scale)
			r.Max.X += l.list.settings.SmallUnitSize(args.Scale)
		}
		return r*/
		guigui.SetPosition(item.Content, p)
		appender.AppendChildWidget(item.Content)
		_, h := item.Content.Size(context)
		p.Y += h
	}

	p = guigui.Position(l)
	guigui.SetPosition(&l.scrollOverlay, p)
	appender.AppendChildWidget(&l.scrollOverlay)
	guigui.SetPosition(&l.dragDropOverlay, p)
	appender.AppendChildWidget(&l.dragDropOverlay)
}

func (l *List) SelectedItem() (ListItem, bool) {
	idx := l.SelectedItemIndex()
	if idx < 0 || idx >= len(l.items) {
		return ListItem{}, false
	}
	return l.items[idx], true
}

func (l *List) ItemAt(index int) (ListItem, bool) {
	if index < 0 || index >= len(l.items) {
		return ListItem{}, false
	}
	return l.items[index], true
}

func (l *List) SelectedItemIndex() int {
	return l.selectedItemIndexPlus1 - 1
}

func (l *List) HoveredItemIndex() int {
	return l.hoveredItemIndexPlus1 - 1
}

func (l *List) SetItems(items []ListItem) {
	l.items = make([]ListItem, len(items))
	copy(l.items, items)
	l.cachedDefaultWidth = 0
	l.cachedDefaultHeight = 0
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
	if l.SelectedItemIndex() != index {
		l.selectedItemIndexPlus1 = index + 1
		guigui.RequestRedraw(l)
	}
	if l.onItemSelected != nil {
		l.onItemSelected(index)
	}
}

func (l *List) JumpToItemIndex(index int) {
	if index < 0 || index >= len(l.items) {
		return
	}
	l.indexToJumpPlus1 = index + 1
}

func (l *List) setHoveredItemIndex(index int) {
	if index < 0 || index >= len(l.items) {
		index = -1
	}
	if l.HoveredItemIndex() == index {
		return
	}
	l.hoveredItemIndexPlus1 = index + 1
	if l.isHoveringVisible() {
		guigui.RequestRedraw(l)
	}
}

func (l *List) ShowItemBorders(show bool) {
	if l.showItemBorders == show {
		return
	}
	l.showItemBorders = true
	guigui.RequestRedraw(l)
}

func (l *List) isHoveringVisible() bool {
	return l.style == ListStyleMenu
}

func (l *List) Style() ListStyle {
	return l.style
}

func (l *List) SetStyle(style ListStyle) {
	if l.style == style {
		return
	}
	l.style = style
	guigui.RequestRedraw(l)
}

func (l *List) calcDropDstIndex(context *guigui.Context) int {
	_, y := ebiten.CursorPosition()
	for i := range l.items {
		if r := l.itemRect(context, i); y < (r.Min.Y+r.Max.Y)/2 {
			return i
		}
	}
	return len(l.items)
}

func (l *List) HandleInput(context *guigui.Context) guigui.HandleInputResult {
	// Process dragging.
	if l.dragDropOverlay.IsDragging() {
		_, y := ebiten.CursorPosition()
		p := guigui.Position(l)
		_, h := l.Size(context)
		var dy float64
		if upperY := p.Y + UnitSize(context); y < upperY {
			dy = float64(upperY-y) / 4
		}
		if lowerY := p.Y + h - UnitSize(context); y >= lowerY {
			dy = float64(lowerY-y) / 4
		}
		l.scrollOverlay.SetOffsetByDelta(0, dy)
		i := l.calcDropDstIndex(context)
		if l.dropDstIndexPlus1-1 != i {
			l.dropDstIndexPlus1 = i + 1
			guigui.RequestRedraw(l)
		}
		return guigui.HandleInputByWidget(l)
	}

	// Process dropping.
	var dropped bool
	if l.dropSrcIndexPlus1 > 0 && l.dropDstIndexPlus1 > 0 {
		dropped = true
		/*if l.callback != nil && l.callback.OnItemDropped != nil {
			l.callback.OnItemDropped(l.dropSrcIndex, l.dropDstIndex)
		}*/
	}

	l.dropSrcIndexPlus1 = 0
	if l.dropDstIndexPlus1 != 0 {
		l.dropDstIndexPlus1 = 0
		guigui.RequestRedraw(l)
	}

	if dropped {
		return guigui.HandleInputByWidget(l)
	}

	if x, y := ebiten.CursorPosition(); image.Pt(x, y).In(guigui.VisibleBounds(l)) {
		_, offsetY := l.scrollOverlay.Offset()
		y -= RoundedCornerRadius(context)
		y -= guigui.Position(l).Y
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
					return guigui.HandleInputByWidget(l)
				}

				wasFocused := guigui.IsFocused(l)
				guigui.Focus(l)
				if l.SelectedItemIndex() != index || !wasFocused {
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
				l.startPressingIndexPlus1 = index + 1
				l.startPressingLeft = left

			case ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft):
				if l.items[index].Draggable && l.SelectedItemIndex() == index && l.startPressingIndexPlus1-1 == index && (l.pressStartX != x || l.pressStartY != y) {
					l.dragDropOverlay.Start(index)
				}

			case inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft):
				if l.SelectedItemIndex() == index && l.startPressingLeft && time.Since(l.lastSelectingItemTime) > 400*time.Millisecond {
					/*if l.callback != nil && l.callback.OnItemEditStarted != nil {
						l.callback.OnItemEditStarted(index)
					}*/
				}
				l.pressStartX = 0
				l.pressStartY = 0
				l.startPressingIndexPlus1 = 0
				l.startPressingLeft = false
			}

			return guigui.HandleInputByWidget(l)
		}
		l.dropSrcIndexPlus1 = 0
		l.pressStartX = 0
		l.pressStartY = 0
	} else {
		l.setHoveredItemIndex(-1)
	}

	return guigui.HandleInputResult{}
}

func (l *List) Update(context *guigui.Context) error {
	w, _ := l.Size(context)
	l.scrollOverlay.SetContentSize(w, l.defaultHeight(context))

	idx := l.indexToJumpPlus1 - 1
	if idx >= 0 {
		y := l.itemYFromIndex(context, idx) - RoundedCornerRadius(context)
		l.scrollOverlay.SetOffset(0, float64(-y))
		l.indexToJumpPlus1 = 0
	}

	return nil
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

func (l *List) itemRect(context *guigui.Context, index int) image.Rectangle {
	_, offsetY := l.scrollOverlay.Offset()
	p := guigui.Position(l)
	w, h := l.Size(context)
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

func (l *List) selectedItemColor(context *guigui.Context) color.Color {
	if l.SelectedItemIndex() < 0 || l.SelectedItemIndex() >= len(l.items) {
		return nil
	}
	if l.style == ListStyleMenu {
		return nil
	}
	if guigui.IsFocused(l) || l.style == ListStyleSidebar {
		return Color(context.ColorMode(), ColorTypeAccent, 0.5)
	}
	return Color(context.ColorMode(), ColorTypeBase, 0.8)
}

func (l *List) Draw(context *guigui.Context, dst *ebiten.Image) {
	if l.style != ListStyleSidebar {
		clr := Color(context.ColorMode(), ColorTypeBase, 1)
		if l.style == ListStyleMenu {
			clr = Color(context.ColorMode(), ColorTypeBase, 0.95)
		}
		p := guigui.Position(l)
		w, h := l.Size(context)
		bounds := image.Rectangle{
			Min: p,
			Max: p.Add(image.Pt(w, h)),
		}
		DrawRoundedRect(context, dst, bounds, clr, RoundedCornerRadius(context))
	}

	// Draw item borders.
	if l.showItemBorders && len(l.items) > 0 {
		_, offsetY := l.scrollOverlay.Offset()
		p := guigui.Position(l)
		w, _ := l.Size(context)
		y := float32(p.Y) + float32(RoundedCornerRadius(context)) + float32(offsetY)
		for i, item := range l.items {
			_, ih := item.Content.Size(context)
			y += float32(ih)
			if i == l.SelectedItemIndex() || i+1 == l.SelectedItemIndex() {
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

	if clr := l.selectedItemColor(context); clr != nil && l.SelectedItemIndex() >= 0 && l.SelectedItemIndex() < len(l.items) {
		r := l.itemRect(context, l.SelectedItemIndex())
		r.Min.X -= RoundedCornerRadius(context)
		r.Max.X += RoundedCornerRadius(context)
		if r.Overlaps(guigui.VisibleBounds(l)) {
			DrawRoundedRect(context, dst, r, clr, RoundedCornerRadius(context))
		}
	}

	if l.isHoveringVisible() && l.HoveredItemIndex() >= 0 && l.HoveredItemIndex() < len(l.items) && l.items[l.HoveredItemIndex()].Selectable {
		r := l.itemRect(context, l.HoveredItemIndex())
		r.Min.X -= RoundedCornerRadius(context)
		r.Max.X += RoundedCornerRadius(context)
		if r.Overlaps(guigui.VisibleBounds(l)) {
			clr := Color(context.ColorMode(), ColorTypeBase, 0.9)
			if l.style == ListStyleMenu {
				clr = Color(context.ColorMode(), ColorTypeAccent, 0.5)
			}
			DrawRoundedRect(context, dst, r, clr, RoundedCornerRadius(context))
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
	if l.dropDstIndexPlus1 > 0 {
		p := guigui.Position(l)
		w, _ := l.Size(context)
		x0 := float32(p.X) + float32(RoundedCornerRadius(context))
		x1 := float32(p.X+w) - float32(RoundedCornerRadius(context))
		y := float32(p.Y)
		y += float32(l.itemYFromIndex(context, l.dropDstIndexPlus1-1))
		_, offsetY := l.scrollOverlay.Offset()
		y += float32(offsetY)
		vector.StrokeLine(dst, x0, y, x1, y, 2*float32(context.Scale()), Color(context.ColorMode(), ColorTypeBase, 0.1), false)
	}
}

/*func (l *List) onDrop(data any) {
	l.dropSrcIndex = data.(int)
}*/

func (l *List) defaultWidth(context *guigui.Context) int {
	if l.cachedDefaultWidth > 0 {
		return l.cachedDefaultWidth
	}
	var w int
	for _, item := range l.items {
		iw, _ := item.Content.Size(context)
		w = max(w, iw)
	}
	w += 2*RoundedCornerRadius(context) + 2*listItemPadding(context)
	l.cachedDefaultWidth = w
	return w
}

func (l *List) defaultHeight(context *guigui.Context) int {
	if l.cachedDefaultHeight > 0 {
		return l.cachedDefaultHeight
	}

	var h int
	h += RoundedCornerRadius(context)
	for _, w := range l.items {
		_, wh := w.Content.Size(context)
		h += wh
	}
	h += RoundedCornerRadius(context)
	l.cachedDefaultHeight = h
	return h
}

func (l *List) Size(context *guigui.Context) (int, int) {
	var w, h int
	if l.widthSet {
		w = l.width
	} else {
		w = l.defaultWidth(context)
	}
	if l.heightSet {
		h = l.height
	} else {
		h = l.defaultHeight(context)
	}
	return w, h
}

func (l *List) SetSize(width, height int) {
	l.width = width
	l.widthSet = true
	l.height = height
	l.heightSet = true
}

func (l *List) SetWidth(width int) {
	l.width = width
	l.widthSet = true
}

func (l *List) SetHeight(height int) {
	l.height = height
	l.heightSet = true
}

func (l *List) ResetWidth() {
	l.widthSet = false
	l.width = 0
}

func (l *List) ResetHeight() {
	l.heightSet = false
	l.height = 0
}

type listFrame struct {
	guigui.DefaultWidget
}

func (l *listFrame) Draw(context *guigui.Context, dst *ebiten.Image) {
	border := RoundedRectBorderTypeInset
	if guigui.Parent(l).(*List).style != ListStyleNormal {
		border = RoundedRectBorderTypeOutset
	}
	p := guigui.Position(l)
	w, h := l.Size(context)
	bounds := image.Rectangle{
		Min: p,
		Max: p.Add(image.Pt(w, h)),
	}
	clr := Color2(context.ColorMode(), ColorTypeBase, 0.7, 0)
	borderWidth := float32(1 * context.Scale())
	DrawRoundedRectBorder(context, dst, bounds, clr, RoundedCornerRadius(context), borderWidth, border)
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
