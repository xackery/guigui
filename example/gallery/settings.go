// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"image"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

type Settings struct {
	guigui.DefaultWidgetBehavior

	groupWidget                *guigui.Widget
	colorModeToggleLabelWidget *guigui.Widget
	colorModeToggleWidget      *guigui.Widget
	localeLabelWidget          *guigui.Widget
	localeSelectorWidget       *guigui.Widget
}

func (s *Settings) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	u := float64(basicwidget.UnitSize(context))

	if s.colorModeToggleLabelWidget == nil {
		var t basicwidget.Text
		t.SetText("Color Mode")
		s.colorModeToggleLabelWidget = guigui.NewWidget(&t)
	}
	if s.colorModeToggleWidget == nil {
		var t basicwidget.ToggleButton
		if context.ColorMode() == guigui.ColorModeDark {
			t.SetValue(true)
		}
		s.colorModeToggleWidget = guigui.NewWidget(&t)
	}
	if s.localeLabelWidget == nil {
		var t basicwidget.Text
		t.SetText("Locale")
		s.localeLabelWidget = guigui.NewWidget(&t)
	}
	if s.localeSelectorWidget == nil {
		// TODO: Make this a selector
		var t basicwidget.Text
		t.SetText("(TODO)")
		s.localeSelectorWidget = guigui.NewWidget(&t)
	}

	if s.groupWidget == nil {
		s.groupWidget = guigui.NewWidget(&basicwidget.Group{})
	}
	group := s.groupWidget.Behavior().(*basicwidget.Group)
	w, _ := widget.Size(context)
	group.SetWidth(context, w-int(1*u))
	group.SetItems([]*basicwidget.GroupItem{
		{
			PrimaryWidget:   s.colorModeToggleLabelWidget,
			SecondaryWidget: s.colorModeToggleWidget,
		},
		{
			PrimaryWidget:   s.localeLabelWidget,
			SecondaryWidget: s.localeSelectorWidget,
		},
	})
	{
		p := widget.Position().Add(image.Pt(int(0.5*u), int(0.5*u)))
		appender.AppendChildWidget(s.groupWidget, p)
	}
}

func (s *Settings) Update(context *guigui.Context, widget *guigui.Widget) error {
	if s.colorModeToggleWidget.Behavior().(*basicwidget.ToggleButton).Value() {
		context.SetColorMode(guigui.ColorModeDark)
	} else {
		context.SetColorMode(guigui.ColorModeLight)
	}
	return nil
}

/*func (s *Settings) SetWidth(width int) {
	s.width = width
}*/

func (s *Settings) Size(context *guigui.Context, widget *guigui.Widget) (int, int) {
	w, h := widget.Parent().Size(context)
	w -= sidebarWidth(context)
	return w, h
}
