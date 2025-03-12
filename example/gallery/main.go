// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package main

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/xackery/guigui"
	_ "github.com/xackery/guigui/basicwidget/cjkfont"
)

type Root struct {
	guigui.RootWidget

	sidebar  Sidebar
	settings Settings
	basic    Basic
	buttons  Buttons
	lists    Lists
	popups   Popups
}

func (r *Root) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	appender.AppendChildWidget(&r.sidebar)

	guigui.SetPosition(&r.sidebar, guigui.Position(r))
	sw, _ := r.sidebar.Size(context)
	p := guigui.Position(r)
	p.X += sw
	guigui.SetPosition(&r.settings, p)
	guigui.SetPosition(&r.basic, p)
	guigui.SetPosition(&r.buttons, p)
	guigui.SetPosition(&r.lists, p)
	guigui.SetPosition(&r.popups, p)

	switch r.sidebar.SelectedItemTag() {
	case "settings":
		appender.AppendChildWidget(&r.settings)
	case "basic":
		appender.AppendChildWidget(&r.basic)
	case "buttons":
		appender.AppendChildWidget(&r.buttons)
	case "lists":
		appender.AppendChildWidget(&r.lists)
	case "popups":
		appender.AppendChildWidget(&r.popups)
	}
}

func (r *Root) Draw(context *guigui.Context, dst *ebiten.Image) {
	//basicwidget.FillBackground(dst, context)
}

func main() {
	op := &guigui.RunOptions{
		Title:             "Component Gallery",
		ScreenTransparent: true,
	}
	if err := guigui.Run(&Root{}, op); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
