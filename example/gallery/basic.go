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

	groupWidget       *guigui.Widget
	buttonLabelWidget *guigui.Widget
	buttonWidget      *guigui.Widget
}

func (b *Basic) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	u := float64(basicwidget.UnitSize(context))

	if b.buttonLabelWidget == nil {
		var t basicwidget.Text
		t.SetText("Text Button")
		b.buttonLabelWidget = guigui.NewWidget(&t)
	}
	if b.buttonWidget == nil {
		var button basicwidget.TextButton
		button.SetText("Click me!")
		b.buttonWidget = guigui.NewWidget(&button)
	}

	if b.groupWidget == nil {
		b.groupWidget = guigui.NewWidget(&basicwidget.Group{})
	}
	group := b.groupWidget.Behavior().(*basicwidget.Group)
	w, _ := widget.Size(context)
	group.SetWidth(context, w-int(1*u))
	group.SetItems([]*basicwidget.GroupItem{
		{
			PrimaryWidget:   b.buttonLabelWidget,
			SecondaryWidget: b.buttonWidget,
		},
	})
	{
		p := widget.Position().Add(image.Pt(int(0.5*u), int(0.5*u)))
		appender.AppendChildWidget(b.groupWidget, p)
	}
}

func (b *Basic) Size(context *guigui.Context, widget *guigui.Widget) (int, int) {
	w, h := widget.Parent().Size(context)
	w -= sidebarWidth(context)
	return w, h
}
