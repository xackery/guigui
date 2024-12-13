// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/hajimehoshi/guigui"
)

func equalColor(c0, c1 color.Color) bool {
	if c0 == c1 {
		return true
	}
	if c0 == nil || c1 == nil {
		return false
	}
	r0, g0, b0, a0 := c0.RGBA()
	r1, g1, b1, a1 := c1.RGBA()
	return r0 == r1 && g0 == g1 && b0 == b1 && a0 == a1
}

type RoundedRectBorderType int

const (
	RoundedRectBorderTypeRegular RoundedRectBorderType = iota
	RoundedRectBorderTypeInset
	RoundedRectBorderTypeOutset
)

var (
	whiteImage    = ebiten.NewImage(3, 3)
	whiteSubImage = whiteImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
)

func init() {
	b := whiteImage.Bounds()
	pix := make([]byte, 4*b.Dx()*b.Dy())
	for i := range pix {
		pix[i] = 0xff
	}
	// This is hacky, but WritePixels is better than Fill in term of automatic texture packing.
	whiteImage.WritePixels(pix)
}

func appendRectVectorPath(path *vector.Path, rx0, ry0, rx1, ry1 float32, radius float32) {
	x0 := rx0
	x1 := rx0 + radius
	x2 := rx1 - radius
	x3 := rx1
	y0 := ry0
	y1 := ry0 + radius
	y2 := ry1 - radius
	y3 := ry1

	path.MoveTo(x1, y0)
	path.LineTo(x2, y0)
	path.ArcTo(x3, y0, x3, y1, radius)
	path.LineTo(x3, y2)
	path.ArcTo(x3, y3, x2, y3, radius)
	path.LineTo(x1, y3)
	path.ArcTo(x0, y3, x0, y2, radius)
	path.LineTo(x0, y1)
	path.ArcTo(x0, y0, x1, y0, radius)
}

type imageKey struct {
	radius      int
	borderWidth float32
	scale       float64
	borderType  RoundedRectBorderType
}

var (
	whiteRoundedRects       = map[imageKey]*ebiten.Image{}
	whiteRoundedRectBorders = map[imageKey]*ebiten.Image{}
)

func ensureWhiteRoundedRect(radius int, scale float64) *ebiten.Image {
	key := imageKey{
		radius: radius,
		scale:  scale,
	}
	if img, ok := whiteRoundedRects[key]; ok {
		return img
	}

	s := radius * 3
	img := ebiten.NewImage(s, s)

	var path vector.Path
	appendRectVectorPath(&path, 0, 0, float32(s), float32(s), float32(radius))
	path.Close()

	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
	}
	op := &ebiten.DrawTrianglesOptions{}
	op.AntiAlias = true
	img.DrawTriangles(vs, is, whiteSubImage, op)

	whiteRoundedRects[key] = img

	return img
}

func ensureWhiteRoundedRectBorder(radius int, borderWidth float32, scale float64, borderType RoundedRectBorderType) *ebiten.Image {
	key := imageKey{
		radius:      radius,
		borderWidth: borderWidth,
		scale:       scale,
		borderType:  borderType,
	}
	if img, ok := whiteRoundedRectBorders[key]; ok {
		return img
	}

	// Use it's own anti-aliasing instead of Ebitengine's anti-aliasing for higher quality result.
	// Ebitengine's anti-aliasing just scales vertice and doesn't create finer paths for anti-aliasing scale.
	const aaScale = 2
	radius *= aaScale
	s := radius * 3
	inset := borderWidth * aaScale

	var path vector.Path
	appendRectVectorPath(&path, 0, 0, float32(s), float32(s), float32(radius))
	switch borderType {
	case RoundedRectBorderTypeRegular:
		appendRectVectorPath(&path, inset, inset, float32(s)-inset, float32(s)-inset, float32(radius)-inset)
	case RoundedRectBorderTypeInset:
		appendRectVectorPath(&path, inset, inset*3/2, float32(s)-inset, float32(s)-inset/2, float32(radius)-inset/2)
	case RoundedRectBorderTypeOutset:
		appendRectVectorPath(&path, inset, inset/2, float32(s)-inset, float32(s)-inset*3/2, float32(radius)-inset/2)
	}
	path.Close()

	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
	}
	op := &ebiten.DrawTrianglesOptions{}
	op.AntiAlias = true
	op.FillRule = ebiten.FillRuleEvenOdd

	offscreen := ebiten.NewImage(s, s)
	offscreen.DrawTriangles(vs, is, whiteSubImage, op)
	defer offscreen.Deallocate()

	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Scale(1.0/aaScale, 1.0/aaScale)
	op2.Filter = ebiten.FilterLinear
	img := ebiten.NewImage(s/aaScale, s/aaScale)
	img.DrawImage(offscreen, op2)

	whiteRoundedRectBorders[key] = img

	return img
}

func DrawRoundedRect(context *guigui.Context, dst *ebiten.Image, rect image.Rectangle, clr color.Color, radius int) {
	if !dst.Bounds().Overlaps(rect) {
		return
	}
	if rect.Dx()/2-1 < radius {
		radius = rect.Dx()/2 - 1
	}
	if rect.Dy()/2-1 < radius {
		radius = rect.Dy()/2 - 1
	}
	drawNinePatch(dst, rect, ensureWhiteRoundedRect(radius, context.Scale()), clr)
}

func DrawRoundedRectBorder(context *guigui.Context, dst *ebiten.Image, rect image.Rectangle, clr color.Color, radius int, borderWidth float32, borderType RoundedRectBorderType) {
	if !dst.Bounds().Overlaps(rect) {
		return
	}
	if rect.Dx()/2-1 < radius {
		radius = rect.Dx()/2 - 1
	}
	if rect.Dy()/2-1 < radius {
		radius = rect.Dy()/2 - 1
	}
	drawNinePatch(dst, rect, ensureWhiteRoundedRectBorder(radius, borderWidth, context.Scale(), borderType), clr)
}

func drawNinePatch(dst *ebiten.Image, rect image.Rectangle, src *ebiten.Image, clr color.Color) {
	if dst.Bounds().Intersect(rect).Empty() {
		return
	}
	partW, partH := src.Bounds().Dx()/3, src.Bounds().Dy()/3

	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleWithColor(clr)

	for j := 0; j < 3; j++ {
		for i := 0; i < 3; i++ {
			sx := 1.0
			sy := 1.0
			var tx, ty int

			switch i {
			case 0:
				tx = rect.Min.X
			case 1:
				sx = float64(rect.Dx()-2*partW) / float64(partW)
				tx = rect.Min.X + partW
			case 2:
				tx = rect.Max.X - partW
			}
			switch j {
			case 0:
				ty = rect.Min.Y
			case 1:
				sy = float64(rect.Dy()-2*partH) / float64(partH)
				ty = rect.Min.Y + partH
			case 2:
				ty = rect.Max.Y - partH
			}

			op.GeoM.Reset()
			op.GeoM.Scale(sx, sy)
			op.GeoM.Translate(float64(tx), float64(ty))
			dst.DrawImage(src.SubImage(image.Rect(i*partW, j*partH, (i+1)*partW, (j+1)*partH)).(*ebiten.Image), op)
		}
	}
}
