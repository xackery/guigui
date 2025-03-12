// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/xackery/guigui"
	"github.com/xackery/guigui/basicwidget"
)

type Popups struct {
	guigui.DefaultWidget

	forms                              [2]basicwidget.Form
	blurBackgroundText                 basicwidget.Text
	blurBackgroundToggleButton         basicwidget.ToggleButton
	closeByClickingOutsideText         basicwidget.Text
	closeByClickingOutsideToggleButton basicwidget.ToggleButton
	showButton                         basicwidget.TextButton

	contextMenuPopupText          basicwidget.Text
	contextMenuPopupClickHereText basicwidget.Text

	simplePopup            basicwidget.Popup
	simplePopupTitleText   basicwidget.Text
	simplePopupCloseButton basicwidget.TextButton

	contextMenuPopup basicwidget.PopupMenu
}

func (p *Popups) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	p.blurBackgroundText.SetText("Blur Background")
	p.closeByClickingOutsideText.SetText("Close by Clicking Outside")
	p.showButton.SetText("Show")
	p.showButton.SetOnUp(func() {
		p.simplePopup.Open()
	})

	u := float64(basicwidget.UnitSize(context))

	w, _ := p.Size(context)
	p.forms[0].SetWidth(context, w-int(1*u))
	p.forms[0].SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &p.blurBackgroundText,
			SecondaryWidget: &p.blurBackgroundToggleButton,
		},
		{
			PrimaryWidget:   &p.closeByClickingOutsideText,
			SecondaryWidget: &p.closeByClickingOutsideToggleButton,
		},
		{
			SecondaryWidget: &p.showButton,
		},
	})
	pt := guigui.Position(p).Add(image.Pt(int(0.5*u), int(0.5*u)))
	guigui.SetPosition(&p.forms[0], pt)
	appender.AppendChildWidget(&p.forms[0])

	p.contextMenuPopupText.SetText("Context Menu")
	p.contextMenuPopupClickHereText.SetText("Click Here by the Right Button")

	p.forms[1].SetWidth(context, w-int(1*u))
	p.forms[1].SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &p.contextMenuPopupText,
			SecondaryWidget: &p.contextMenuPopupClickHereText,
		},
	})
	_, h := p.forms[0].Size(context)
	pt.Y += h + int(0.5*u)
	guigui.SetPosition(&p.forms[1], pt)
	appender.AppendChildWidget(&p.forms[1])

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
		p.simplePopupCloseButton.SetOnUp(func() {
			p.simplePopup.Close()
		})
		w, h := p.simplePopupCloseButton.Size(context)
		pt = contentBounds.Max.Add(image.Pt(-int(0.5*u)-w, -int(0.5*u)-h))
		guigui.SetPosition(&p.simplePopupCloseButton, pt)
		appender.AppendChildWidget(&p.simplePopupCloseButton)
	})
	p.simplePopup.SetContentBounds(contentBounds)
	p.simplePopup.SetBackgroundBlurred(p.blurBackgroundToggleButton.Value())
	p.simplePopup.SetCloseByClickingOutside(p.closeByClickingOutsideToggleButton.Value())

	appender.AppendChildWidget(&p.simplePopup)

	p.contextMenuPopup.SetItemsByStrings([]string{"Item 1", "Item 2", "Item 3"})
	appender.AppendChildWidget(&p.contextMenuPopup)
}

func (p *Popups) HandleInput(context *guigui.Context) guigui.HandleInputResult {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		pt := image.Pt(ebiten.CursorPosition())
		if pt.In(guigui.VisibleBounds(&p.contextMenuPopupClickHereText)) {
			guigui.SetPosition(&p.contextMenuPopup, pt)
			p.contextMenuPopup.Open(context)
		}
	}
	return guigui.HandleInputResult{}
}

func (p *Popups) Size(context *guigui.Context) (int, int) {
	w, h := guigui.Parent(p).Size(context)
	w -= sidebarWidth(context)
	return w, h
}
