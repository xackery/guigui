// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"image"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

type Basic struct {
	guigui.DefaultWidgetBehavior

	group            basicwidget.Group
	textButtonText   basicwidget.Text
	textButton       basicwidget.TextButton
	toggleButtonText basicwidget.Text
	toggleButton     basicwidget.ToggleButton
	textFieldText    basicwidget.Text
	textField        basicwidget.TextField
}

func (b *Basic) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	b.textButtonText.SetText("Text Button")
	b.textButton.SetText("Click Me!")
	b.toggleButtonText.SetText("Toggle Button")
	b.textFieldText.SetText("Text Field")
	b.textField.SetHorizontalAlign(basicwidget.HorizontalAlignEnd)

	u := float64(basicwidget.UnitSize(context))
	w, _ := widget.Size(context)
	b.group.SetWidth(context, w-int(1*u))
	b.group.SetItems([]*basicwidget.GroupItem{
		{
			PrimaryWidget:   &b.textButtonText,
			SecondaryWidget: &b.textButton,
		},
		{
			PrimaryWidget:   &b.toggleButtonText,
			SecondaryWidget: &b.toggleButton,
		},
		{
			PrimaryWidget:   &b.textFieldText,
			SecondaryWidget: &b.textField,
		},
	})
	{
		p := widget.Position().Add(image.Pt(int(0.5*u), int(0.5*u)))
		appender.AppendChildWidget(&b.group, p)
	}
}

func (b *Basic) Size(context *guigui.Context) (int, int) {
	w, h := context.WidgetFromBehavior(b).Parent().Size(context)
	w -= sidebarWidth(context)
	return w, h
}
