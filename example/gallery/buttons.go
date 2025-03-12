// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"image"

	"github.com/xackery/guigui"
	"github.com/xackery/guigui/basicwidget"
)

type Buttons struct {
	guigui.DefaultWidget

	form                basicwidget.Form
	textButtonText      basicwidget.Text
	textButton          basicwidget.TextButton
	textImageButtonText basicwidget.Text
	textImageButton     basicwidget.TextButton
	toggleButtonText    basicwidget.Text
	toggleButton        basicwidget.ToggleButton

	err error
}

func (b *Buttons) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	b.textButtonText.SetText("Text Button")
	b.textButton.SetText("Button")
	b.textImageButtonText.SetText("Text w/ Image Button")
	b.textImageButton.SetText("Button")
	img, err := theImageCache.Get("check", context.ColorMode())
	if err != nil {
		b.err = err
		return
	}
	b.textImageButton.SetImage(img)
	b.toggleButtonText.SetText("Toggle Button")

	u := float64(basicwidget.UnitSize(context))
	w, _ := b.Size(context)
	b.form.SetWidth(context, w-int(1*u))
	p := guigui.Position(b).Add(image.Pt(int(0.5*u), int(0.5*u)))
	guigui.SetPosition(&b.form, p)

	b.form.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &b.textButtonText,
			SecondaryWidget: &b.textButton,
		},
		{
			PrimaryWidget:   &b.textImageButtonText,
			SecondaryWidget: &b.textImageButton,
		},
		{
			PrimaryWidget:   &b.toggleButtonText,
			SecondaryWidget: &b.toggleButton,
		},
	})
	appender.AppendChildWidget(&b.form)
}

func (b *Buttons) Update(context *guigui.Context) error {
	if b.err != nil {
		return b.err
	}
	return nil
}

func (b *Buttons) Size(context *guigui.Context) (int, int) {
	w, h := guigui.Parent(b).Size(context)
	w -= sidebarWidth(context)
	return w, h
}
