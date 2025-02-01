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
	guigui.RootWidget

	resetButton basicwidget.TextButton
	incButton   basicwidget.TextButton
	decButton   basicwidget.TextButton
	counterText basicwidget.Text

	counter int
}

func (r *Root) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	{
		w, h := r.Size(context)
		w -= 2 * basicwidget.UnitSize(context)
		h -= 4 * basicwidget.UnitSize(context)
		r.counterText.SetSize(w, h)
		p := guigui.Position(r)
		p.X += basicwidget.UnitSize(context)
		p.Y += basicwidget.UnitSize(context)
		guigui.SetPosition(&r.counterText, p)
		appender.AppendChildWidget(&r.counterText)
	}

	r.resetButton.SetText("Reset")
	r.resetButton.SetWidth(6 * basicwidget.UnitSize(context))
	r.resetButton.SetOnUp(func() {
		r.counter = 0
	})
	{
		p := guigui.Position(r)
		_, h := r.Size(context)
		p.X += basicwidget.UnitSize(context)
		p.Y += h - 2*basicwidget.UnitSize(context)
		guigui.SetPosition(&r.resetButton, p)
		appender.AppendChildWidget(&r.resetButton)
	}

	r.incButton.SetText("Increment")
	r.incButton.SetWidth(6 * basicwidget.UnitSize(context))
	r.incButton.SetOnUp(func() {
		r.counter++
	})
	{
		p := guigui.Position(r)
		w, h := r.Size(context)
		p.X += w - 7*basicwidget.UnitSize(context)
		p.Y += h - 2*basicwidget.UnitSize(context)
		guigui.SetPosition(&r.incButton, p)
		appender.AppendChildWidget(&r.incButton)
	}

	r.decButton.SetText("Decrement")
	r.decButton.SetWidth(6 * basicwidget.UnitSize(context))
	r.decButton.SetOnUp(func() {
		r.counter--
	})
	{
		p := guigui.Position(r)
		w, h := r.Size(context)
		p.X += w - int(13.5*float64(basicwidget.UnitSize(context)))
		p.Y += h - 2*basicwidget.UnitSize(context)
		guigui.SetPosition(&r.decButton, p)
		appender.AppendChildWidget(&r.decButton)
	}
}

func (r *Root) Update(context *guigui.Context) error {
	if r.counter == 0 {
		guigui.Disable(&r.resetButton)
	} else {
		guigui.Enable(&r.resetButton)
	}
	r.counterText.SetSelectable(true)
	r.counterText.SetBold(true)
	r.counterText.SetHorizontalAlign(basicwidget.HorizontalAlignCenter)
	r.counterText.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
	r.counterText.SetScale(4)
	r.counterText.SetText(fmt.Sprintf("%d", r.counter))

	return nil
}

func (r *Root) Draw(context *guigui.Context, dst *ebiten.Image) {
	basicwidget.FillBackground(dst, context)
}

func main() {
	op := &guigui.RunOptions{
		Title:           "Counter",
		WindowMinWidth:  640,
		WindowMinHeight: 480,
	}
	if err := guigui.Run(&Root{}, op); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
