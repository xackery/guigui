// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package basicwidget

import (
	"github.com/hajimehoshi/guigui"
)

type DropdownList struct {
	guigui.DefaultWidget

	textButton TextButton
	popupMenu  PopupMenu
}

func (d *DropdownList) SetOnValueChanged(f func(index int)) {
	d.popupMenu.SetOnClosed(func(index int) {
		f(index)
	})
}

func (d *DropdownList) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	img, err := theResourceImages.Get("unfold_more", context.ColorMode())
	if err != nil {
		panic(err)
	}
	d.textButton.SetImage(img)
	if item, ok := d.popupMenu.SelectedItem(); ok {
		d.textButton.SetText(item.Text)
	} else {
		d.textButton.SetText("")
	}

	d.textButton.SetOnDown(func() {
		pt := guigui.Position(d)
		pt.Y -= int(float64(d.popupMenu.SelectedItemIndex()) * LineHeight(context))
		pt.Y = max(pt.Y, 0)
		// TODO: Chaning the position here might be too late here.
		// A glitch is visible when the dropdown list is reopened.
		// TODO: Add a check mark for a selected item.
		guigui.SetPosition(&d.popupMenu, pt)
		d.popupMenu.Open(context)
	})

	guigui.SetPosition(&d.textButton, guigui.Position(d))
	appender.AppendChildWidget(&d.textButton)
	appender.AppendChildWidget(&d.popupMenu)
}

func (d *DropdownList) SetItemsByStrings(items []string) {
	d.popupMenu.SetItemsByStrings(items)
}

func (d *DropdownList) SelectedItemIndex() int {
	return d.popupMenu.SelectedItemIndex()
}

func (d *DropdownList) SetSelectedItemIndex(index int) {
	d.popupMenu.SetSelectedItemIndex(index)
}

func (d *DropdownList) Size(context *guigui.Context) (int, int) {
	return d.textButton.Size(context)
}

/*func (d *DropdownList) Enable() {
	d.textButton().Enable()
}

func (d *DropdownList) Disable() {
	d.button.Disable()
	if d.popup != nil {
		d.popup.Close()
	}
}
*/
