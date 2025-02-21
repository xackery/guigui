// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"image"
	"sync"

	"golang.org/x/text/language"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

type Settings struct {
	guigui.DefaultWidget

	form                  basicwidget.Form
	colorModeText         basicwidget.Text
	colorModeDropdownList basicwidget.DropdownList
	localeText            basicwidget.Text
	localeDropdownList    basicwidget.DropdownList

	initOnce sync.Once
}

func (s *Settings) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	s.colorModeText.SetText("Color Mode")
	s.colorModeDropdownList.SetItemsByStrings([]string{"Light", "Dark"})
	s.colorModeDropdownList.SetOnValueChanged(func(index int) {
		switch index {
		case 0:
			context.SetColorMode(guigui.ColorModeLight)
		case 1:
			context.SetColorMode(guigui.ColorModeDark)
		}
	})

	s.localeText.SetText("Locale")
	langs := []string{"(Default)", "en", "ja", "ko", "zh-Hans", "zh-Hant"}
	s.localeDropdownList.SetItemsByStrings(langs)
	s.localeDropdownList.SetOnValueChanged(func(index int) {
		if index == 0 {
			context.SetAppLocales(nil)
			return
		}
		lang := language.MustParse(langs[index])
		context.SetAppLocales([]language.Tag{lang})
	})

	s.initOnce.Do(func() {
		switch context.ColorMode() {
		case guigui.ColorModeLight:
			s.colorModeDropdownList.SetSelectedItemIndex(0)
		case guigui.ColorModeDark:
			s.colorModeDropdownList.SetSelectedItemIndex(1)
		}

		s.localeDropdownList.SetSelectedItemIndex(0)
	})

	u := float64(basicwidget.UnitSize(context))
	w, _ := s.Size(context)
	s.form.SetWidth(context, w-int(1*u))
	s.form.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &s.colorModeText,
			SecondaryWidget: &s.colorModeDropdownList,
		},
		{
			PrimaryWidget:   &s.localeText,
			SecondaryWidget: &s.localeDropdownList,
		},
	})
	{
		p := guigui.Position(s).Add(image.Pt(int(0.5*u), int(0.5*u)))
		guigui.SetPosition(&s.form, p)
		appender.AppendChildWidget(&s.form)
	}
}

func (s *Settings) Size(context *guigui.Context) (int, int) {
	w, h := guigui.Parent(s).Size(context)
	w -= sidebarWidth(context)
	return w, h
}
