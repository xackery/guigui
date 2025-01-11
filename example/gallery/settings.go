// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"image"
	"sync"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

type Settings struct {
	guigui.DefaultWidgetBehavior

	group               basicwidget.Group
	colorModeToggleText basicwidget.Text
	colorModeToggle     basicwidget.ToggleButton
	localeText          basicwidget.Text
	localeSelector      basicwidget.Text

	initOnce sync.Once
}

func (s *Settings) AppendChildWidgets(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	s.initOnce.Do(func() {
		if context.ColorMode() == guigui.ColorModeDark {
			s.colorModeToggle.SetValue(true)
		}
	})

	s.colorModeToggleText.SetText("Color Mode")
	s.localeText.SetText("Locale")
	// TODO: Make this a selector
	s.localeSelector.SetText("(TODO)")

	u := float64(basicwidget.UnitSize(context))
	w, _ := s.Size(context)
	s.group.SetWidth(context, w-int(1*u))
	s.group.SetItems([]*basicwidget.GroupItem{
		{
			PrimaryWidget:   &s.colorModeToggleText,
			SecondaryWidget: &s.colorModeToggle,
		},
		{
			PrimaryWidget:   &s.localeText,
			SecondaryWidget: &s.localeSelector,
		},
	})
	{
		p := context.Widget(s).Position().Add(image.Pt(int(0.5*u), int(0.5*u)))
		appender.AppendChildWidget(&s.group, p)
	}
}

func (s *Settings) Update(context *guigui.Context) error {
	if s.colorModeToggle.Value() {
		context.SetColorMode(guigui.ColorModeDark)
	} else {
		context.SetColorMode(guigui.ColorModeLight)
	}
	return nil
}

func (s *Settings) Size(context *guigui.Context) (int, int) {
	w, h := guigui.Parent(s).Size(context)
	w -= sidebarWidth(context)
	return w, h
}
