// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"image"
	"sync"

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

	initOnce sync.Once
}

func (b *Popups) AppendChildWidgets(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	b.blurBackgroundText.SetText("Blur Background")
	b.showButton.SetText("Show")
	u := float64(basicwidget.UnitSize(context))
	w, _ := b.Size(context)
	b.group.SetWidth(context, w-int(1*u))
	b.group.SetItems([]*basicwidget.GroupItem{
		{
			PrimaryWidget:   &b.blurBackgroundText,
			SecondaryWidget: &b.blurBackgroundToggleButton,
		},
		{
			SecondaryWidget: &b.showButton,
		},
	})
	p := guigui.Position(b).Add(image.Pt(int(0.5*u), int(0.5*u)))
	guigui.SetPosition(&b.group, p)
	appender.AppendChildWidget(&b.group)

	{
		contentWidth := int(12 * u)
		contentHeight := int(6 * u)
		bounds := guigui.Bounds(&b.simplePopup)
		contentPosition := image.Point{
			X: bounds.Min.X + (bounds.Dx()-contentWidth)/2,
			Y: bounds.Min.Y + (bounds.Dy()-contentHeight)/2,
		}
		contentBounds := image.Rectangle{
			Min: contentPosition,
			Max: contentPosition.Add(image.Pt(contentWidth, contentHeight)),
		}
		b.simplePopup.SetContent(func(context *guigui.Context, appender *basicwidget.ContainerChildWidgetAppender) {
			b.simplePopupTitleText.SetText("Hello!")
			b.simplePopupTitleText.SetBold(true)
			p := contentBounds.Min.Add(image.Pt(int(0.5*u), int(0.5*u)))
			guigui.SetPosition(&b.simplePopupTitleText, p)
			appender.AppendChildWidget(&b.simplePopupTitleText)

			b.simplePopupCloseButton.SetText("Close")
			w, h := b.simplePopupCloseButton.Size(context)
			p = contentBounds.Max.Add(image.Pt(-int(0.5*u)-w, -int(0.5*u)-h))
			guigui.SetPosition(&b.simplePopupCloseButton, p)
			appender.AppendChildWidget(&b.simplePopupCloseButton)
		})
		b.simplePopup.SetContentBounds(contentBounds)
		b.simplePopup.SetBackgroundBlurred(b.blurBackgroundToggleButton.Value())
		appender.AppendChildWidget(&b.simplePopup)
	}
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
			// Use Close() instead of Hide() to gradually close the popup.
			p.simplePopup.Close()
		}
	}
	return nil
}

func (p *Popups) Size(context *guigui.Context) (int, int) {
	w, h := guigui.Parent(p).Size(context)
	w -= sidebarWidth(context)
	return w, h
}
