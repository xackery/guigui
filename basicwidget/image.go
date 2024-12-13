// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/guigui"
)

type Image struct {
	guigui.DefaultWidgetBehavior

	image *ebiten.Image

	needsRedraw bool
}

func (i *Image) Update(context *guigui.Context, widget *guigui.Widget) error {
	if i.needsRedraw {
		widget.RequestRedraw()
		i.needsRedraw = false
	}
	return nil
}

func (i *Image) Draw(context *guigui.Context, widget *guigui.Widget, dst *ebiten.Image) {
	if i.image == nil {
		return
	}

	imgScale := min(float64(widget.Bounds().Dx())/float64(i.image.Bounds().Dx()), float64(widget.Bounds().Dy())/float64(i.image.Bounds().Dy()))
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(imgScale, imgScale)
	op.GeoM.Translate(float64(widget.Bounds().Min.X), float64(widget.Bounds().Min.Y))
	if !widget.IsEnabled() {
		// TODO: Reduce the saturation?
		op.ColorScale.ScaleAlpha(0.25)
	}
	// TODO: Use a better filter.
	op.Filter = ebiten.FilterLinear
	dst.DrawImage(i.image, op)
}

func (i *Image) HasImage() bool {
	return i.image != nil
}

func (i *Image) SetImage(image *ebiten.Image) {
	if i.image == image {
		return
	}
	i.image = image
	i.needsRedraw = true
}
