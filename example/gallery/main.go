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
	guigui.RootWidgetBehavior

	sidebar  Sidebar
	settings Settings
	basic    Basic
	lists    Lists
}

func (r *Root) AppendChildWidgets(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	appender.AppendChildWidget(&r.sidebar, context.Widget(r).Position())

	sw, _ := r.sidebar.Size(context)
	p := context.Widget(r).Position()
	p.X += sw

	switch r.sidebar.SelectedItemTag() {
	case "settings":
		appender.AppendChildWidget(&r.settings, p)
	case "basic":
		appender.AppendChildWidget(&r.basic, p)
	case "lists":
		appender.AppendChildWidget(&r.lists, p)
	}
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
