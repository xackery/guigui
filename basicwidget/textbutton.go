// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package basicwidget

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/guigui"
)

type TextButton struct {
	guigui.DefaultWidget

	button Button
	text   Text
	image  Image

	textColor color.Color

	width    int
	widthSet bool
}

func (t *TextButton) SetOnDown(f func()) {
	t.button.SetOnDown(f)
}

func (t *TextButton) SetOnUp(f func()) {
	t.button.SetOnUp(f)
}

func (t *TextButton) SetText(text string) {
	t.text.SetText(text)
}

func (t *TextButton) SetImage(image *ebiten.Image) {
	t.image.SetImage(image)
}

func (t *TextButton) SetTextColor(clr color.Color) {
	if equalColor(t.textColor, clr) {
		return
	}
	t.textColor = clr
	guigui.RequestRedraw(t)
}

func (t *TextButton) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	w, h := t.Size(context)
	t.button.SetSize(context, w, h)
	guigui.SetPosition(&t.button, guigui.Position(t))
	appender.AppendChildWidget(&t.button)

	imgSize := int(LineHeight(context))

	tw, _ := t.text.TextSize(context)
	t.text.SetSize(tw, h)
	textP := guigui.Position(t)
	if t.image.HasImage() {
		textP.X += (w - tw + UnitSize(context)/4) / 2
		textP.X -= (t.textImagePadding(context) + imgSize) / 2
	} else {
		textP.X += (w - tw) / 2
	}
	if t.button.isActive() {
		textP.Y += int(1 * context.Scale())
	}
	guigui.SetPosition(&t.text, textP)
	appender.AppendChildWidget(&t.text)

	t.image.SetSize(context, imgSize, imgSize)
	imgP := guigui.Position(t)
	imgP.X = textP.X + tw + t.textImagePadding(context)
	imgP.Y += (h - imgSize) / 2
	if t.button.isActive() {
		imgP.Y += int(1 * context.Scale())
	}
	guigui.SetPosition(&t.image, imgP)
	appender.AppendChildWidget(&t.image)
}

func (t *TextButton) Update(context *guigui.Context) error {
	if !guigui.IsEnabled(&t.button) {
		t.text.SetColor(Color(context.ColorMode(), ColorTypeBase, 0.5))
	} else {
		t.text.SetColor(t.textColor)
	}

	t.text.SetHorizontalAlign(HorizontalAlignCenter)
	t.text.SetVerticalAlign(VerticalAlignMiddle)

	return nil
}

func (t *TextButton) Size(context *guigui.Context) (int, int) {
	_, dh := defaultButtonSize(context)
	if t.widthSet {
		return t.width, dh
	}
	w, _ := t.text.TextSize(context)
	if t.image.HasImage() {
		imgSize := int(LineHeight(context))
		return w + t.textImagePadding(context) + imgSize + UnitSize(context)*3/4, dh
	}
	return w + UnitSize(context), dh
}

func (t *TextButton) SetWidth(width int) {
	t.width = width
	t.widthSet = true
}

func (t *TextButton) ResetWidth() {
	t.width = 0
	t.widthSet = false
}

func (t *TextButton) textImagePadding(context *guigui.Context) int {
	return UnitSize(context) / 4
}
