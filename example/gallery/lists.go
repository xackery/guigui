// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

type Lists struct {
	guigui.DefaultWidgetBehavior

	groupWidget         *guigui.Widget
	textListLabelWidget *guigui.Widget
	textListWidget      *guigui.Widget
}

func (l *Lists) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	if l.textListLabelWidget == nil {
		var t basicwidget.Text
		t.SetText("Text List")
		l.textListLabelWidget = guigui.NewWidget(&t)
	}
	if l.textListWidget == nil {
		l.textListWidget = guigui.NewWidget(&basicwidget.TextList{})
	}
	var items []basicwidget.TextListItem
	for i := 0; i < 100; i++ {
		items = append(items, basicwidget.TextListItem{
			Text: fmt.Sprintf("Item %d", i),
		})
	}
	l.textListWidget.Behavior().(*basicwidget.TextList).SetItems(items)

	if l.groupWidget == nil {
		l.groupWidget = guigui.NewWidget(&basicwidget.Group{})
	}
	u := float64(basicwidget.UnitSize(context))
	group := l.groupWidget.Behavior().(*basicwidget.Group)
	w, _ := widget.Size(context)
	group.SetWidth(context, w-int(1*u))
	group.SetItems([]*basicwidget.GroupItem{
		{
			PrimaryWidget:   l.textListLabelWidget,
			SecondaryWidget: l.textListWidget,
		},
	})
	{
		p := widget.Position().Add(image.Pt(int(0.5*u), int(0.5*u)))
		appender.AppendChildWidget(l.groupWidget, p)
	}
}

func (l *Lists) Size(context *guigui.Context, widget *guigui.Widget) (int, int) {
	w, h := widget.Parent().Size(context)
	w -= sidebarWidth(context)
	return w, h
}
