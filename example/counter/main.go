// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package main

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

type Root struct {
	guigui.RootWidgetBehavior

	resetButtonWidget *guigui.Widget
	incButtonWidget   *guigui.Widget
	decButtonWidget   *guigui.Widget
	counterTextWidget *guigui.Widget

	counter int
}

func (r *Root) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	if r.counterTextWidget == nil {
		r.counterTextWidget = guigui.NewWidget(&basicwidget.Text{})
	}
	{
		w, h := widget.Size(context)
		w -= 2 * basicwidget.UnitSize(context)
		h -= 4 * basicwidget.UnitSize(context)
		r.counterTextWidget.Behavior().(*basicwidget.Text).SetSize(w, h)
		p := widget.Position()
		p.X += basicwidget.UnitSize(context)
		p.Y += basicwidget.UnitSize(context)
		appender.AppendChildWidget(r.counterTextWidget, p)
	}

	if r.resetButtonWidget == nil {
		var b basicwidget.TextButton
		b.SetText("Reset")
		b.SetWidth(6 * basicwidget.UnitSize(context))
		r.resetButtonWidget = guigui.NewWidget(&b)
	}
	{
		p := widget.Position()
		_, h := widget.Size(context)
		p.X += basicwidget.UnitSize(context)
		p.Y += h - 2*basicwidget.UnitSize(context)
		appender.AppendChildWidget(r.resetButtonWidget, p)
	}

	if r.incButtonWidget == nil {
		var b basicwidget.TextButton
		b.SetText("Increment")
		b.SetWidth(6 * basicwidget.UnitSize(context))
		r.incButtonWidget = guigui.NewWidget(&b)
	}
	{
		p := widget.Position()
		w, h := widget.Size(context)
		p.X += w - 7*basicwidget.UnitSize(context)
		p.Y += h - 2*basicwidget.UnitSize(context)
		appender.AppendChildWidget(r.incButtonWidget, p)
	}

	if r.decButtonWidget == nil {
		var b basicwidget.TextButton
		b.SetText("Decrement")
		b.SetWidth(6 * basicwidget.UnitSize(context))
		r.decButtonWidget = guigui.NewWidget(&b)
	}
	{
		p := widget.Position()
		w, h := widget.Size(context)
		p.X += w - int(13.5*float64(basicwidget.UnitSize(context)))
		p.Y += h - 2*basicwidget.UnitSize(context)
		appender.AppendChildWidget(r.decButtonWidget, p)
	}
}

func (r *Root) Update(context *guigui.Context, widget *guigui.Widget) error {
	for e := range r.incButtonWidget.DequeueEvents() {
		args := e.(basicwidget.ButtonEvent)
		if args.Type == basicwidget.ButtonEventTypeUp {
			r.counter++
		}
	}
	for e := range r.decButtonWidget.DequeueEvents() {
		args := e.(basicwidget.ButtonEvent)
		if args.Type == basicwidget.ButtonEventTypeUp {
			r.counter--
		}
	}
	for e := range r.resetButtonWidget.DequeueEvents() {
		args := e.(basicwidget.ButtonEvent)
		if args.Type == basicwidget.ButtonEventTypeUp {
			r.counter = 0
		}
	}

	if r.counter == 0 {
		r.resetButtonWidget.Disable()
	} else {
		r.resetButtonWidget.Enable()
	}
	t := r.counterTextWidget.Behavior().(*basicwidget.Text)
	t.SetSelectable(true)
	t.SetBold(true)
	t.SetHorizontalAlign(basicwidget.HorizontalAlignCenter)
	t.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
	t.SetScale(4)
	t.SetText(fmt.Sprintf("%d", r.counter))

	return nil
}

func (r *Root) Draw(context *guigui.Context, widget *guigui.Widget, dst *ebiten.Image) {
	basicwidget.FillBackground(dst, context)
}

func main() {
	op := &guigui.RunOptions{
		Title:           "Counter",
		WindowMinWidth:  640,
		WindowMinHeight: 480,
	}
	if err := guigui.Run(guigui.NewWidget(&Root{}), op); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
