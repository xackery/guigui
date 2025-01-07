// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package basicwidget

import (
	"image"
	"image/color"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/hajimehoshi/guigui"
)

type TextList struct {
	guigui.DefaultWidgetBehavior

	listWidget    *guigui.Widget
	textListItems []TextListItem
}

/*type TextListCallback struct {
	OnItemSelected    func(index int)
	OnItemEditStarted func(index int, str string) (from int)
	OnItemEditEnded   func(index int, str string)
	OnItemDropped     func(from, to int)
	OnContextMenu     func(index int, x, y int)
}*/

type TextListItem struct {
	Text      string
	DummyText string
	Color     color.Color
	Header    bool
	Disabled  bool
	Border    bool
	Draggable bool
	Tag       any

	listItem ListItem
}

func (t *TextListItem) selectable() bool {
	return !t.Header && !t.Disabled && !t.Border
}

/*func NewTextList(settings *model.Settings, callback *TextListCallback) *TextList {
	t := &TextList{
		settings: settings,
		callback: callback,
	}
	t.list = NewList(settings, &ListCallback{
		OnItemSelected: func(index int) {
			if callback != nil && callback.OnItemSelected != nil {
				callback.OnItemSelected(index)
			}
		},
		OnItemEditStarted: func(index int) {
			if index < 0 || index >= len(t.textListItems) {
				return
			}
			if !t.textListItems[index].selectable() {
				return
			}
			item, ok := t.textListItems[index].listItem.WidgetWithHeight.(*textListTextItem)
			if !ok {
				return
			}
			if callback != nil && callback.OnItemEditStarted != nil {
				item.edit(callback.OnItemEditStarted(index, item.textListItem.Text))
			}
		},
		OnItemDropped: func(from int, to int) {
			if callback != nil && callback.OnItemDropped != nil {
				callback.OnItemDropped(from, to)
			}
		},
		OnContextMenu: func(index int, x, y int) {
			if callback != nil && callback.OnContextMenu != nil {
				callback.OnContextMenu(index, x, y)
			}
		},
	})
	t.AddChild(t.list, &view.IdentityLayouter{})
	return t
}*/

func (t *TextList) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	_ = t.list()
	appender.AppendChildWidget(t.listWidget, widget.Position())
}

func (t *TextList) list() *List {
	if t.listWidget == nil {
		t.listWidget = guigui.NewWidget(&List{})
	}
	return t.listWidget.Behavior().(*List)
}

func (t *TextList) SelectedItemIndex() int {
	return t.list().SelectedItemIndex()
}

func (t *TextList) SelectedItem() (TextListItem, bool) {
	if t.list().SelectedItemIndex() < 0 || t.list().SelectedItemIndex() >= len(t.textListItems) {
		return TextListItem{}, false
	}
	return t.textListItems[t.list().SelectedItemIndex()], true
}

func (t *TextList) ItemAt(index int) (TextListItem, bool) {
	if index < 0 || index >= len(t.textListItems) {
		return TextListItem{}, false
	}
	return t.textListItems[index], true
}

func (t *TextList) SetItemsByStrings(strs []string) {
	items := make([]TextListItem, len(strs))
	for i, str := range strs {
		items[i].Text = str
	}
	t.SetItems(items)
}

func (t *TextList) SetItems(items []TextListItem) {
	t.textListItems = make([]TextListItem, len(items))
	copy(t.textListItems, items)

	listItems := make([]ListItem, len(items))
	for i, item := range items {
		li := t.newTextListItem(item)
		t.textListItems[i].listItem = li
		listItems[i] = li
	}
	t.list().SetItems(listItems)
}

func (t *TextList) ItemsCount() int {
	return len(t.textListItems)
}

func (t *TextList) Tag(index int) any {
	return t.textListItems[index].Tag
}

func (t *TextList) newTextListItem(item TextListItem) ListItem {
	if item.Border {
		return ListItem{
			Content:    guigui.NewWidget(&textListBorderItem{}),
			Selectable: false,
		}
	}
	textItem := &textListTextItem{
		textList:     t,
		textListItem: item,
	}
	return ListItem{
		Content:    guigui.NewWidget(textItem),
		Selectable: item.selectable(),
		Wide:       item.Header,
		Draggable:  item.Draggable,
	}
}

func (t *TextList) SetSelectedItemIndex(index int) {
	t.list().SetSelectedItemIndex(index)
}

func (t *TextList) JumpToItemIndex(index int) {
	t.list().JumpToItemIndex(index)
}

func (t *TextList) SetStyle(style ListStyle) {
	t.list().SetStyle(style)
}

func (t *TextList) SetItemString(str string, index int) {
	t.textListItems[index].Text = str
}

func (t *TextList) AppendItem(item TextListItem) {
	t.AddItem(item, len(t.textListItems))
}

func (t *TextList) AddItem(item TextListItem, index int) {
	t.textListItems = slices.Insert(t.textListItems, index, item)
	li := t.newTextListItem(item)
	t.textListItems[index].listItem = li
	t.list().AddItem(li, index)
}

func (t *TextList) RemoveItem(index int) {
	t.textListItems = slices.Delete(t.textListItems, index, index+1)
	t.list().RemoveItem(index)
}

func (t *TextList) MoveItem(from, to int) {
	moveItemInSlice(t.textListItems, from, 1, to)
	t.list().MoveItem(from, to)
}

func (t *TextList) Update(context *guigui.Context, widget *guigui.Widget) error {
	for i, li := range t.textListItems {
		item, ok := li.listItem.Content.Behavior().(*textListTextItem)
		if !ok {
			continue
		}
		text := item.textWidget.Behavior().(*Text)
		text.SetBold(item.textListItem.Header)
		if widget.HasFocusedChildWidget() && t.list().SelectedItemIndex() == i ||
			(t.list().IsHoveringVisible() && t.list().HoveredItemIndex() == i) && li.listItem.Selectable {
			text.SetColor(DefaultActiveListItemTextColor(context))
		} else if !li.listItem.Selectable && !item.textListItem.Header {
			text.SetColor(DefaultDisabledListItemTextColor(context))
		} else {
			text.SetColor(li.Color)
		}
	}
	return nil
}

func (t *TextList) Size(context *guigui.Context, widget *guigui.Widget) (int, int) {
	return t.list().Size(context, widget)
}

/*func (t *TextList) MinimumWidth(scale float64) int {
	var w int
	for _, item := range t.textListItems {
		if ww, ok := item.listItem.WidgetWithHeight.(view.WidgetWithWidth); ok {
			w = max(w, ww.Width(scale))
		}
	}
	return w + 4*t.settings.SmallUnitSize(scale)
}

func (t *TextList) ContentHeight(context *guigui.Context) int {
	return t.list().ContentHeight(context)
}*/

type textListTextItem struct {
	guigui.DefaultWidgetBehavior

	textList     *TextList
	textListItem TextListItem
	textWidget   *guigui.Widget
}

/*func newTextListTextItem(settings *model.Settings, textList *TextList, textListItem TextListItem) *textListTextItem {
	t := &textListTextItem{
		textList:     textList,
		textListItem: textListItem,
		label:        NewLabel(settings),
	}
	t.label.SetText(t.labelText())
	t.AddChild(t.label, view.LayoutFunc(func(args view.WidgetArgs) image.Rectangle {
		bounds := args.Bounds
		if t.textListItem.Header {
			bounds.Min.X += t.settings.SmallUnitSize(args.Scale)
			bounds.Max.X -= t.settings.SmallUnitSize(args.Scale)
			bounds.Min.Y += t.settings.SmallUnitSize(args.Scale)
			bounds.Max.Y -= t.settings.SmallUnitSize(args.Scale)
		}
		return bounds
	}))
	return t
}*/

func (t *textListTextItem) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	if t.textWidget == nil {
		t.textWidget = guigui.NewWidget(&Text{})
	}
	text := t.textWidget.Behavior().(*Text)
	p := widget.Position()
	if t.textListItem.Header {
		p.X += UnitSize(context) / 2
		w, h := widget.Size(context)
		text.SetSize(w-UnitSize(context), h)
	}
	text.SetText(t.text())
	text.SetVerticalAlign(VerticalAlignMiddle)
	appender.AppendChildWidget(t.textWidget, p)
}

func (t *textListTextItem) text() string {
	if t.textListItem.DummyText != "" {
		return t.textListItem.DummyText
	}
	return t.textListItem.Text
}

func (t *textListTextItem) Draw(context *guigui.Context, widget *guigui.Widget, dst *ebiten.Image) {
	if !t.textListItem.Header {
		return
	}
	p := widget.Position()
	w, h := widget.Size(context)
	bounds := image.Rectangle{
		Min: p,
		Max: p.Add(image.Pt(w, h)),
	}
	DrawRoundedRect(context, dst, bounds, Color(context.ColorMode(), ColorTypeBase, 0.6), RoundedCornerRadius(context))
}

func (t *textListTextItem) Size(context *guigui.Context, widget *guigui.Widget) (int, int) {
	w, _ := widget.Parent().Size(context)
	return w, int(LineHeight(context))
}

func (t *textListTextItem) index() int {
	for i, li := range t.textList.textListItems {
		tt, ok := li.listItem.Content.Behavior().(*textListTextItem)
		if !ok {
			continue
		}
		if tt == t {
			return i
		}
	}
	return -1
}

/*func (t *textListTextItem) edit(from int) {
	t.label.Hide()
	t0 := t.textListItem.Text[:from]
	var l0 *Label
	if t0 != "" {
		l0 = NewLabel(t.settings)
		l0.SetText(t0)
		t.AddChild(l0, &view.IdentityLayouter{})
	}
	var tf *TextField
	tf = NewTextField(t.settings, &TextFieldCallback{
		OnTextUpdated: func(value string) {
			t.textListItem.Text = t0 + value
		},
		OnTextConfirmed: func(value string) {
			t.textListItem.Text = t0 + value

			t.label.SetText(t.labelText())
			if l0 != nil {
				l0.RemoveSelf()
			}
			tf.RemoveSelf()
			t.label.Show()

			if t.textList.callback != nil && t.textList.callback.OnItemEditEnded != nil {
				t.textList.callback.OnItemEditEnded(t.index(), t0+value)
			}
		},
	})
	tf.SetText(t.textListItem.Text[from:])
	tf.SetHorizontalAlign(HorizontalAlignStart)
	t.AddChild(tf, view.LayoutFunc(func(args view.WidgetArgs) image.Rectangle {
		bounds := args.Bounds
		if l0 != nil {
			bounds.Min.X += l0.Width(args.Scale)
		}
		return bounds
	}))
	tf.SelectAll()
	tf.Focus()
}*/

type textListBorderItem struct {
	guigui.DefaultWidgetBehavior
}

func (t *textListBorderItem) Draw(context *guigui.Context, widget *guigui.Widget, dst *ebiten.Image) {
	p := widget.Position()
	w, h := widget.Size(context)
	x0 := float32(p.X)
	x1 := float32(p.X + w)
	y := float32(p.Y) + float32(h)/2
	width := float32(1 * context.Scale())
	vector.StrokeLine(dst, x0, y, x1, y, width, Color(context.ColorMode(), ColorTypeBase, 0.8), false)
}

func (t *textListBorderItem) Size(context *guigui.Context, widget *guigui.Widget) (int, int) {
	w, _ := widget.Parent().Size(context)
	return w, UnitSize(context) / 2
}
