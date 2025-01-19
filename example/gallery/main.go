// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package main

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	_ "github.com/hajimehoshi/guigui/basicwidget/cjkfont"
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

func (r *Root) AppendChildWidgets(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	appender.AppendChildWidget(&r.sidebar)
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

func (r *Root) Update(context *guigui.Context) error {
	guigui.SetPosition(&r.sidebar, guigui.Position(r))
	sw, _ := r.sidebar.Size(context)
	p := guigui.Position(r)
	p.X += sw
	guigui.SetPosition(&r.settings, p)
	guigui.SetPosition(&r.basic, p)
	guigui.SetPosition(&r.buttons, p)
	guigui.SetPosition(&r.lists, p)
	guigui.SetPosition(&r.popups, p)
	return nil
}

func (r *Root) Draw(context *guigui.Context, dst *ebiten.Image) {
	basicwidget.FillBackground(dst, context)
}

func main() {
	op := &guigui.RunOptions{
		Title: "Component Gallery",
	}
	if err := guigui.Run(&Root{}, op); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
