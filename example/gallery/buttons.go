// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"image"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

type Buttons struct {
	guigui.DefaultWidget

	group               basicwidget.Group
	textButtonText      basicwidget.Text
	textButton          basicwidget.TextButton
	textImageButtonText basicwidget.Text
	textImageButton     basicwidget.TextButton
}

func (b *Buttons) AppendChildWidgets(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	u := float64(basicwidget.UnitSize(context))
	w, _ := b.Size(context)
	b.group.SetWidth(context, w-int(1*u))
	p := guigui.Position(b).Add(image.Pt(int(0.5*u), int(0.5*u)))
	guigui.SetPosition(&b.group, p)

	b.group.SetItems([]*basicwidget.GroupItem{
		{
			PrimaryWidget:   &b.textButtonText,
			SecondaryWidget: &b.textButton,
		},
		{
			PrimaryWidget:   &b.textImageButtonText,
			SecondaryWidget: &b.textImageButton,
		},
	})
	appender.AppendChildWidget(&b.group)
}

func (b *Buttons) Update(context *guigui.Context) error {
	b.textButtonText.SetText("Text Button")
	b.textButton.SetText("Button")
	b.textImageButtonText.SetText("Text w/ Image Button")
	b.textImageButton.SetText("Button")
	img, err := theImageCache.Get("check", context.ColorMode())
	if err != nil {
		return err
	}
	b.textImageButton.SetImage(img)

	return nil
}

func (b *Buttons) Size(context *guigui.Context) (int, int) {
	w, h := guigui.Parent(b).Size(context)
	w -= sidebarWidth(context)
	return w, h
}
