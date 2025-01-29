// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"image"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

type Popups struct {
	guigui.DefaultWidget

	group                      basicwidget.Group
	blurBackgroundText         basicwidget.Text
	blurBackgroundToggleButton basicwidget.ToggleButton
	showButton                 basicwidget.TextButton

	simplePopup            basicwidget.Popup
	simplePopupTitleText   basicwidget.Text
	simplePopupCloseButton basicwidget.TextButton
}

func (p *Popups) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	u := float64(basicwidget.UnitSize(context))
	w, _ := p.Size(context)
	p.group.SetWidth(context, w-int(1*u))
	p.group.SetItems([]*basicwidget.GroupItem{
		{
			PrimaryWidget:   &p.blurBackgroundText,
			SecondaryWidget: &p.blurBackgroundToggleButton,
		},
		{
			SecondaryWidget: &p.showButton,
		},
	})
	pt := guigui.Position(p).Add(image.Pt(int(0.5*u), int(0.5*u)))
	guigui.SetPosition(&p.group, pt)
	appender.AppendChildWidget(&p.group)

	appender.AppendChildWidget(&p.simplePopup)
}

func (p *Popups) Update(context *guigui.Context) error {
	for e := range guigui.DequeueEvents(&p.showButton) {
		args := e.(basicwidget.ButtonEvent)
		if args.Type == basicwidget.ButtonEventTypeUp {
			p.simplePopup.Open()
		}
	}
	for e := range guigui.DequeueEvents(&p.simplePopupCloseButton) {
		args := e.(basicwidget.ButtonEvent)
		if args.Type == basicwidget.ButtonEventTypeUp {
			p.simplePopup.Close()
		}
	}

	p.blurBackgroundText.SetText("Blur Background")
	p.showButton.SetText("Show")

	u := float64(basicwidget.UnitSize(context))
	contentWidth := int(12 * u)
	contentHeight := int(6 * u)
	bounds := guigui.Bounds(&p.simplePopup)
	contentPosition := image.Point{
		X: bounds.Min.X + (bounds.Dx()-contentWidth)/2,
		Y: bounds.Min.Y + (bounds.Dy()-contentHeight)/2,
	}
	contentBounds := image.Rectangle{
		Min: contentPosition,
		Max: contentPosition.Add(image.Pt(contentWidth, contentHeight)),
	}
	p.simplePopup.SetContent(func(context *guigui.Context, appender *basicwidget.ContainerChildWidgetAppender) {
		p.simplePopupTitleText.SetText("Hello!")
		p.simplePopupTitleText.SetBold(true)
		pt := contentBounds.Min.Add(image.Pt(int(0.5*u), int(0.5*u)))
		guigui.SetPosition(&p.simplePopupTitleText, pt)
		appender.AppendChildWidget(&p.simplePopupTitleText)

		p.simplePopupCloseButton.SetText("Close")
		w, h := p.simplePopupCloseButton.Size(context)
		pt = contentBounds.Max.Add(image.Pt(-int(0.5*u)-w, -int(0.5*u)-h))
		guigui.SetPosition(&p.simplePopupCloseButton, pt)
		appender.AppendChildWidget(&p.simplePopupCloseButton)
	})
	p.simplePopup.SetContentBounds(contentBounds)
	p.simplePopup.SetBackgroundBlurred(p.blurBackgroundToggleButton.Value())

	return nil
}

func (p *Popups) Size(context *guigui.Context) (int, int) {
	w, h := guigui.Parent(p).Size(context)
	w -= sidebarWidth(context)
	return w, h
}
