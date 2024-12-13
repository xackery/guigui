// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/oklab"

	"github.com/hajimehoshi/guigui"
)

var (
	blue   = oklab.OklchModel.Convert(color.RGBA{R: 0x00, G: 0x5a, B: 0xff, A: 0xff}).(oklab.Oklch)
	green  = oklab.OklchModel.Convert(color.RGBA{R: 0x03, G: 0xaf, B: 0x7a, A: 0xff}).(oklab.Oklch)
	yellow = oklab.OklchModel.Convert(color.RGBA{R: 0xff, G: 0xf1, B: 0x00, A: 0xff}).(oklab.Oklch)
	red    = oklab.OklchModel.Convert(color.RGBA{R: 0xff, G: 0x4b, B: 0x00, A: 0xff}).(oklab.Oklch)
)

var (
	white = oklab.OklchModel.Convert(color.White).(oklab.Oklch)
	black = oklab.OklchModel.Convert(oklab.Oklab{L: 0.2, A: 0, B: 0, Alpha: 1}).(oklab.Oklch)
	gray  = oklab.OklchModel.Convert(oklab.Oklab{L: 0.6, A: 0, B: 0, Alpha: 1}).(oklab.Oklch)
)

type ColorType int

const (
	ColorTypeBase ColorType = iota
	ColorTypeAccent
	ColorTypeInfo
	ColorTypeSuccess
	ColorTypeWarning
	ColorTypeDanger
)

func Color(colorMode guigui.ColorMode, typ ColorType, lightnessInLightMode float64) color.Color {
	return Color2(colorMode, typ, lightnessInLightMode, 1-lightnessInLightMode)
}

func Color2(colorMode guigui.ColorMode, typ ColorType, lightnessInLightMode, lightnessInDarkMode float64) color.Color {
	var base color.Color
	switch typ {
	case ColorTypeBase:
		base = gray
	case ColorTypeAccent:
		base = blue
	case ColorTypeInfo:
		base = blue
	case ColorTypeSuccess:
		base = green
	case ColorTypeWarning:
		base = yellow
	case ColorTypeDanger:
		base = red
	default:
		panic(fmt.Sprintf("basicwidget: invalid color type: %d", typ))
	}
	switch colorMode {
	case guigui.ColorModeLight:
		return getColor(base, lightnessInLightMode, black, white)
	case guigui.ColorModeDark:
		return getColor(base, lightnessInDarkMode, black, white)
	default:
		panic(fmt.Sprintf("basicwidget: invalid color mode: %d", colorMode))
	}
}

func getColor(base color.Color, lightness float64, back, front color.Color) color.Color {
	c0 := oklab.OklchModel.Convert(back).(oklab.Oklch)
	c1 := oklab.OklchModel.Convert(front).(oklab.Oklch)
	l := oklab.OklchModel.Convert(base).(oklab.Oklch).L
	l = max(min(l, c1.L), c0.L)
	l2 := c0.L*(1-lightness) + c1.L*lightness
	if l2 < l {
		rate := (l2 - c0.L) / (l - c0.L)
		return mixColor(c0, base, rate)
	}
	rate := (l2 - l) / (c1.L - l)
	return mixColor(base, c1, rate)
}

func mixColor(clr0, clr1 color.Color, rate float64) color.Color {
	if rate == 0 {
		return clr0
	}
	if rate == 1 {
		return clr1
	}
	okClr0 := oklab.OklabModel.Convert(clr0).(oklab.Oklab)
	okClr1 := oklab.OklabModel.Convert(clr1).(oklab.Oklab)
	return oklab.Oklab{
		L:     okClr0.L*(1-rate) + okClr1.L*rate,
		A:     okClr0.A*(1-rate) + okClr1.A*rate,
		B:     okClr0.B*(1-rate) + okClr1.B*rate,
		Alpha: okClr0.Alpha*(1-rate) + okClr1.Alpha*rate,
	}
}

func FillBackground(dst *ebiten.Image, context *guigui.Context) {
	dst.Fill(Color(context.ColorMode(), ColorTypeBase, 0.95))
}
