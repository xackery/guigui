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
	guigui.DefaultWidget

	list                List
	textListItemWidgets []*textListItemWidget
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

func (t *TextList) AppendChildWidgets(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	appender.AppendChildWidget(&t.list, guigui.Position(t))
}

func (t *TextList) SelectedItemIndex() int {
	return t.list.SelectedItemIndex()
}

func (t *TextList) SelectedItem() (TextListItem, bool) {
	if t.list.SelectedItemIndex() < 0 || t.list.SelectedItemIndex() >= len(t.textListItemWidgets) {
		return TextListItem{}, false
	}
	return t.textListItemWidgets[t.list.SelectedItemIndex()].textListItem, true
}

func (t *TextList) ItemAt(index int) (TextListItem, bool) {
	if index < 0 || index >= len(t.textListItemWidgets) {
		return TextListItem{}, false
	}
	return t.textListItemWidgets[index].textListItem, true
}

func (t *TextList) SetItemsByStrings(strs []string) {
	items := make([]TextListItem, len(strs))
	for i, str := range strs {
		items[i].Text = str
	}
	t.SetItems(items)
}

func (t *TextList) SetItems(items []TextListItem) {
	if cap(t.textListItemWidgets) < len(items) {
		t.textListItemWidgets = append(t.textListItemWidgets, make([]*textListItemWidget, len(items)-cap(t.textListItemWidgets))...)
	}
	t.textListItemWidgets = t.textListItemWidgets[:len(items)]

	listItems := make([]ListItem, len(items))
	for i, item := range items {
		if t.textListItemWidgets[i] == nil {
			t.textListItemWidgets[i] = &textListItemWidget{
				textList:     t,
				textListItem: item,
			}
		} else {
			t.textListItemWidgets[i].textListItem = item
		}
		listItems[i] = t.textListItemWidgets[i].listItem()
	}
	t.list.SetItems(listItems)
}

func (t *TextList) ItemsCount() int {
	return len(t.textListItemWidgets)
}

func (t *TextList) Tag(index int) any {
	return t.textListItemWidgets[index].textListItem.Tag
}

func (t *TextList) SetSelectedItemIndex(index int) {
	t.list.SetSelectedItemIndex(index)
}

func (t *TextList) JumpToItemIndex(index int) {
	t.list.JumpToItemIndex(index)
}

func (t *TextList) SetStyle(style ListStyle) {
	t.list.SetStyle(style)
}

func (t *TextList) SetItemString(str string, index int) {
	t.textListItemWidgets[index].textListItem.Text = str
}

func (t *TextList) AppendItem(item TextListItem) {
	t.AddItem(item, len(t.textListItemWidgets))
}

func (t *TextList) AddItem(item TextListItem, index int) {
	t.textListItemWidgets = slices.Insert(t.textListItemWidgets, index, &textListItemWidget{
		textList:     t,
		textListItem: item,
	})
	t.list.AddItem(t.textListItemWidgets[index].listItem(), index)
}

func (t *TextList) RemoveItem(index int) {
	t.textListItemWidgets = slices.Delete(t.textListItemWidgets, index, index+1)
	t.list.RemoveItem(index)
}

func (t *TextList) MoveItem(from, to int) {
	moveItemInSlice(t.textListItemWidgets, from, 1, to)
	t.list.MoveItem(from, to)
}

func (t *TextList) Update(context *guigui.Context) error {
	for i, item := range t.textListItemWidgets {
		item.text.SetBold(item.textListItem.Header)
		if guigui.HasFocusedChildWidget(t) && t.list.SelectedItemIndex() == i ||
			(t.list.IsHoveringVisible() && t.list.HoveredItemIndex() == i) && item.selectable() {
			item.text.SetColor(DefaultActiveListItemTextColor(context))
		} else if !item.selectable() && !item.textListItem.Header {
			item.text.SetColor(DefaultDisabledListItemTextColor(context))
		} else {
			item.text.SetColor(item.textListItem.Color)
		}
	}
	return nil
}

func (t *TextList) Size(context *guigui.Context) (int, int) {
	return t.list.Size(context)
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

type textListItemWidget struct {
	guigui.DefaultWidget

	textList     *TextList
	textListItem TextListItem

	text Text
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

func (t *textListItemWidget) AppendChildWidgets(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	p := guigui.Position(t)
	if t.textListItem.Header {
		p.X += UnitSize(context) / 2
		w, h := t.Size(context)
		t.text.SetSize(w-UnitSize(context), h)
	}
	t.text.SetText(t.textString())
	t.text.SetVerticalAlign(VerticalAlignMiddle)
	appender.AppendChildWidget(&t.text, p)
}

func (t *textListItemWidget) textString() string {
	if t.textListItem.DummyText != "" {
		return t.textListItem.DummyText
	}
	return t.textListItem.Text
}

func (t *textListItemWidget) Draw(context *guigui.Context, dst *ebiten.Image) {
	if t.textListItem.Border {
		p := guigui.Position(t)
		w, h := t.Size(context)
		x0 := float32(p.X)
		x1 := float32(p.X + w)
		y := float32(p.Y) + float32(h)/2
		width := float32(1 * context.Scale())
		vector.StrokeLine(dst, x0, y, x1, y, width, Color(context.ColorMode(), ColorTypeBase, 0.8), false)
		return
	}
	if t.textListItem.Header {
		p := guigui.Position(t)
		w, h := t.Size(context)
		bounds := image.Rectangle{
			Min: p,
			Max: p.Add(image.Pt(w, h)),
		}
		DrawRoundedRect(context, dst, bounds, Color(context.ColorMode(), ColorTypeBase, 0.6), RoundedCornerRadius(context))
	}
}

func (t *textListItemWidget) Size(context *guigui.Context) (int, int) {
	w, _ := guigui.Parent(t).Size(context)
	if t.textListItem.Border {
		return w, UnitSize(context) / 2
	}
	return w, int(LineHeight(context))
}

func (t *textListItemWidget) index() int {
	for i, tt := range t.textList.textListItemWidgets {
		if tt == t {
			return i
		}
	}
	return -1
}

func (t *textListItemWidget) selectable() bool {
	return t.textListItem.selectable() && !t.textListItem.Border
}

func (t *textListItemWidget) listItem() ListItem {
	return ListItem{
		Content:    t,
		Selectable: t.selectable(),
		Wide:       t.textListItem.Header,
		Draggable:  t.textListItem.Draggable,
	}
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
