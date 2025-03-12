// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package basicwidget

import (
	"embed"
	"image/color"
	"image/draw"
	"image/png"
	"path"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/xackery/guigui"
)

//go:embed resource/*.png
var imageResource embed.FS

type imageCacheKey struct {
	name      string
	colorMode guigui.ColorMode
}

type resourceImages struct {
	m map[imageCacheKey]*ebiten.Image
}

var theResourceImages = &resourceImages{}

func (i *resourceImages) Get(name string, colorMode guigui.ColorMode) (*ebiten.Image, error) {
	key := imageCacheKey{
		name:      name,
		colorMode: colorMode,
	}
	if img, ok := i.m[key]; ok {
		return img, nil
	}

	f, err := imageResource.Open(path.Join("resource", name+".png"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	pImg, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	if colorMode == guigui.ColorModeDark {
		// Create a white image for dark mode.
		rgbaImg := pImg.(draw.Image)
		b := rgbaImg.Bounds()
		for j := b.Min.Y; j < b.Max.Y; j++ {
			for i := b.Min.X; i < b.Max.X; i++ {
				if _, _, _, a := rgbaImg.At(i, j).RGBA(); a > 0 {
					a16 := uint16(a)
					rgbaImg.Set(i, j, color.RGBA64{a16, a16, a16, a16})
				}
			}
		}
		pImg = rgbaImg
	}
	img := ebiten.NewImageFromImage(pImg)
	if i.m == nil {
		i.m = map[imageCacheKey]*ebiten.Image{}
	}
	i.m[key] = img
	return img, nil
}
