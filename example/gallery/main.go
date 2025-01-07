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

	sidebarWidget  *guigui.Widget
	settingsWidget *guigui.Widget
	basicWidget    *guigui.Widget
	listsWidget    *guigui.Widget
}

func (r *Root) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	if r.sidebarWidget == nil {
		r.sidebarWidget = guigui.NewWidget(&Sidebar{})
	}
	appender.AppendChildWidget(r.sidebarWidget, widget.Position())

	sw, _ := r.sidebarWidget.Size(context)
	p := widget.Position()
	p.X += sw

	sidebar := r.sidebarWidget.Behavior().(*Sidebar)
	switch sidebar.SelectedItemTag() {
	case "settings":
		if r.settingsWidget == nil {
			r.settingsWidget = guigui.NewWidget(&Settings{})
		}
		appender.AppendChildWidget(r.settingsWidget, p)
	case "basic":
		if r.basicWidget == nil {
			r.basicWidget = guigui.NewWidget(&Basic{})
		}
		appender.AppendChildWidget(r.basicWidget, p)
	case "lists":
		if r.listsWidget == nil {
			r.listsWidget = guigui.NewWidget(&Lists{})
		}
		appender.AppendChildWidget(r.listsWidget, p)
	}
}

func (r *Root) Update(context *guigui.Context, widget *guigui.Widget) error {
	return nil
}

func (r *Root) Draw(context *guigui.Context, widget *guigui.Widget, dst *ebiten.Image) {
	basicwidget.FillBackground(dst, context)
}

func main() {
	op := &guigui.RunOptions{
		Title: "Component Gallery",
	}
	if err := guigui.Run(guigui.NewWidget(&Root{}), op); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
