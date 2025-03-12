// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"embed"
	"image/color"
	"image/draw"
	"image/png"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/xackery/guigui"
)

//go:embed *.png
var pngImages embed.FS

type imageCacheKey struct {
	name      string
	colorMode guigui.ColorMode
}

type imageCache struct {
	m map[imageCacheKey]*ebiten.Image
}

var theImageCache = &imageCache{}

func (i *imageCache) Get(name string, colorMode guigui.ColorMode) (*ebiten.Image, error) {
	key := imageCacheKey{
		name:      name,
		colorMode: colorMode,
	}
	if img, ok := i.m[key]; ok {
		return img, nil
	}

	f, err := pngImages.Open(name + ".png")
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
