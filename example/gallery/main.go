// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package main

import (
	"fmt"
	"image"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

type Root struct {
	guigui.RootWidgetBehavior

	sideBarWidget *guigui.Widget

	// General
	generalGroup               basicwidget.Group
	generalGroupWidget         *guigui.Widget
	colorModeToggleLabelWidget *guigui.Widget
	colorModeToggleWidget      *guigui.Widget
	localeLabelWidget          *guigui.Widget
	localeSelectorWidget       *guigui.Widget
}

func (r *Root) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	u := float64(basicwidget.UnitSize(context))

	if r.sideBarWidget == nil {
		r.sideBarWidget = guigui.NewWidget(&Sidebar{})
	}
	appender.AppendChildWidget(r.sideBarWidget, widget.Position())

	if r.generalGroupWidget == nil {
		r.generalGroupWidget = guigui.NewWidget(&r.generalGroup)
	}

	if r.colorModeToggleLabelWidget == nil {
		var t basicwidget.Text
		t.SetText("Color Mode")
		r.colorModeToggleLabelWidget = guigui.NewWidget(&t)
	}
	if r.colorModeToggleWidget == nil {
		var t basicwidget.ToggleButton
		if context.ColorMode() == guigui.ColorModeDark {
			t.SetValue(true)
		}
		r.colorModeToggleWidget = guigui.NewWidget(&t)
	}
	if r.localeLabelWidget == nil {
		var t basicwidget.Text
		t.SetText("Locale")
		r.localeLabelWidget = guigui.NewWidget(&t)
	}
	if r.localeSelectorWidget == nil {
		// TODO: Make this a selector
		var t basicwidget.Text
		t.SetText("(TODO)")
		r.localeSelectorWidget = guigui.NewWidget(&t)
	}
	w, _ := widget.Size(context)
	r.generalGroup.SetWidth(context, w-int(9*u))
	r.generalGroup.SetItems([]*basicwidget.GroupItem{
		{
			PrimaryWidget:   r.colorModeToggleLabelWidget,
			SecondaryWidget: r.colorModeToggleWidget,
		},
		{
			PrimaryWidget:   r.localeLabelWidget,
			SecondaryWidget: r.localeSelectorWidget,
		},
	})
	{
		p := widget.Position().Add(image.Pt(int(8.5*u), int(0.5*u)))
		appender.AppendChildWidget(r.generalGroupWidget, p)
	}
}

func (r *Root) Update(context *guigui.Context, widget *guigui.Widget) error {
	if r.colorModeToggleWidget.Behavior().(*basicwidget.ToggleButton).Value() {
		context.SetColorMode(guigui.ColorModeDark)
	} else {
		context.SetColorMode(guigui.ColorModeLight)
	}
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
