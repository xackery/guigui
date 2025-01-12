// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import "github.com/hajimehoshi/ebiten/v2"

var blurShader *ebiten.Shader

const blurShaderSource = `//kage:unit pixels

package main

const blurSize = 5

var Rate float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	minPos := imageSrc0Origin()
	maxPos := minPos + imageSrc0Size()
	var sum vec4
	var count int
	for j := -blurSize; j <= blurSize; j++ {
		for i := -blurSize; i <= blurSize; i++ {
			pos := vec2(float(i), float(j)) + srcPos
			if minPos.x <= pos.x && pos.x <= maxPos.x && minPos.y <= pos.y && pos.y <= maxPos.y {
				sum += imageSrc0UnsafeAt(pos)
				count++
			}
		}
	}
	return imageSrc0UnsafeAt(srcPos) * (1-Rate) + sum / float(count) * Rate
}
`

func init() {
	s, err := ebiten.NewShader([]byte(blurShaderSource))
	if err != nil {
		panic(err)
	}
	blurShader = s
}

func DrawBlurredImage(dst *ebiten.Image, src *ebiten.Image, rate float64) {
	if rate == 0 {
		return
	}

	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = src
	op.GeoM.Translate(float64(src.Bounds().Min.X), float64(src.Bounds().Min.Y))
	op.Uniforms = map[string]any{
		"Rate": rate,
	}
	op.Blend = ebiten.BlendCopy
	dst.DrawRectShader(src.Bounds().Dx(), src.Bounds().Dy(), blurShader, op)
}
