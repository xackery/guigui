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

	groupWidget             *guigui.Widget
	textButtonLabelWidget   *guigui.Widget
	textButtonWidget        *guigui.Widget
	toggleButtonLabelWidget *guigui.Widget
	toggleButtonWidget      *guigui.Widget
	textFieldLabelWidget    *guigui.Widget
	textFieldWidget         *guigui.Widget
}

func (b *Basic) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	u := float64(basicwidget.UnitSize(context))

	if b.textButtonLabelWidget == nil {
		var t basicwidget.Text
		t.SetText("Text Button")
		b.textButtonLabelWidget = guigui.NewWidget(&t)
	}
	if b.textButtonWidget == nil {
		var button basicwidget.TextButton
		button.SetText("Click me!")
		b.textButtonWidget = guigui.NewWidget(&button)
	}
	if b.toggleButtonLabelWidget == nil {
		var t basicwidget.Text
		t.SetText("Toggle Button")
		b.toggleButtonLabelWidget = guigui.NewWidget(&t)
	}
	if b.toggleButtonWidget == nil {
		var button basicwidget.ToggleButton
		b.toggleButtonWidget = guigui.NewWidget(&button)
	}
	if b.textFieldLabelWidget == nil {
		var t basicwidget.Text
		t.SetText("Text Field")
		b.textFieldLabelWidget = guigui.NewWidget(&t)
	}
	if b.textFieldWidget == nil {
		var t basicwidget.TextField
		t.SetHorizontalAlign(basicwidget.HorizontalAlignEnd)
		b.textFieldWidget = guigui.NewWidget(&t)
	}

	if b.groupWidget == nil {
		b.groupWidget = guigui.NewWidget(&basicwidget.Group{})
	}
	group := b.groupWidget.Behavior().(*basicwidget.Group)
	w, _ := widget.Size(context)
	group.SetWidth(context, w-int(1*u))
	group.SetItems([]*basicwidget.GroupItem{
		{
			PrimaryWidget:   b.textButtonLabelWidget,
			SecondaryWidget: b.textButtonWidget,
		},
		{
			PrimaryWidget:   b.toggleButtonLabelWidget,
			SecondaryWidget: b.toggleButtonWidget,
		},
		{
			PrimaryWidget:   b.textFieldLabelWidget,
			SecondaryWidget: b.textFieldWidget,
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
