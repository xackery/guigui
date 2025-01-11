// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/guigui"
)

type Image struct {
	guigui.DefaultWidget

	image *ebiten.Image

	widthMinusDefault  int
	heightMinusDefault int

	needsRedraw bool
}

func (i *Image) Update(context *guigui.Context) error {
	if i.needsRedraw {
		guigui.RequestRedraw(i)
		i.needsRedraw = false
	}
	return nil
}

func (i *Image) Draw(context *guigui.Context, dst *ebiten.Image) {
	if i.image == nil {
		return
	}

	p := guigui.Position(i)
	w, h := i.Size(context)
	imgScale := min(float64(w)/float64(i.image.Bounds().Dx()), float64(h)/float64(i.image.Bounds().Dy()))
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(imgScale, imgScale)
	op.GeoM.Translate(float64(p.X), float64(p.Y))
	if !guigui.IsEnabled(i) {
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

func defaultImageSize(context *guigui.Context) (int, int) {
	return 6 * UnitSize(context), 6 * UnitSize(context)
}

func (i *Image) Size(context *guigui.Context) (int, int) {
	dw, dh := defaultImageSize(context)
	return i.widthMinusDefault + dw, i.heightMinusDefault + dh
}

func (i *Image) SetSize(context *guigui.Context, width, height int) {
	dw, dh := defaultImageSize(context)
	i.widthMinusDefault = width - dw
	i.heightMinusDefault = height - dh
}
