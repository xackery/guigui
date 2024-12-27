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
	guigui.DefaultWidgetBehavior

	resetButtonWidget *guigui.Widget
	incButtonWidget   *guigui.Widget
	decButtonWidget   *guigui.Widget
	counterTextWidget *guigui.Widget

	counter int
}

func (r *Root) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	if r.resetButtonWidget == nil {
		var b basicwidget.TextButton
		b.SetText("Reset")
		r.resetButtonWidget = guigui.NewWidget(&b)
	}
	if r.incButtonWidget == nil {
		var b basicwidget.TextButton
		b.SetText("Increment")
		r.incButtonWidget = guigui.NewWidget(&b)
	}
	if r.decButtonWidget == nil {
		var b basicwidget.TextButton
		b.SetText("Decrement")
		r.decButtonWidget = guigui.NewWidget(&b)
	}
	if r.counterTextWidget == nil {
		r.counterTextWidget = guigui.NewWidget(&basicwidget.Text{})
	}

	c := &basicwidget.LinearGrid{
		Items: []basicwidget.LinearGridItem{
			{
				Widget: guigui.NewWidget(&basicwidget.LinearGrid{
					Direction: basicwidget.LinearGridDirectionVertical,
					Items: []basicwidget.LinearGridItem{
						{
							Widget:   r.counterTextWidget,
							Size:     1,
							SizeUnit: basicwidget.SizeUnitFraction,
						},
						{
							Widget: guigui.NewWidget(&basicwidget.LinearGrid{
								Direction: basicwidget.LinearGridDirectionHorizontal,
								Items: []basicwidget.LinearGridItem{
									{
										Widget: r.resetButtonWidget,
										Size:   6,
									},
									{
										Size:     1,
										SizeUnit: basicwidget.SizeUnitFraction,
									},
									{
										Widget: r.decButtonWidget,
										Size:   6,
									},
									{
										Size: 0.5,
									},
									{
										Widget: r.incButtonWidget,
										Size:   6,
									},
								},
							}),
							Size: 1,
						},
					},
				}),
				Size:          1,
				SizeUnit:      basicwidget.SizeUnitFraction,
				PaddingLeft:   1,
				PaddingRight:  1,
				PaddingTop:    1,
				PaddingBottom: 1,
			},
		},
	}
	c.AppendChildWidgets(context, widget, appender)
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
