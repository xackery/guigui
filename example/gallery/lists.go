// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"fmt"
	"image"

	"github.com/xackery/guigui"
	"github.com/xackery/guigui/basicwidget"
)

type Lists struct {
	guigui.DefaultWidget

	form         basicwidget.Form
	textListText basicwidget.Text
	textList     basicwidget.TextList
}

func (l *Lists) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	l.textListText.SetText("Text List")
	var items []basicwidget.TextListItem
	for i := 0; i < 100; i++ {
		items = append(items, basicwidget.TextListItem{
			Text: fmt.Sprintf("Item %d", i),
		})
	}
	l.textList.SetItems(items)
	l.textList.SetHeight(6 * basicwidget.UnitSize(context))

	u := float64(basicwidget.UnitSize(context))
	w, _ := l.Size(context)
	l.form.SetWidth(context, w-int(1*u))
	l.form.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &l.textListText,
			SecondaryWidget: &l.textList,
		},
	})
	{
		p := guigui.Position(l).Add(image.Pt(int(0.5*u), int(0.5*u)))
		guigui.SetPosition(&l.form, p)
		appender.AppendChildWidget(&l.form)
	}
}

func (l *Lists) Size(context *guigui.Context) (int, int) {
	w, h := guigui.Parent(l).Size(context)
	w -= sidebarWidth(context)
	return w, h
}
