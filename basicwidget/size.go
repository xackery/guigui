// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import "github.com/xackery/guigui"

const baseUnitSize = 24

func FontSize(context *guigui.Context) float64 {
	return baseUnitSize * context.Scale() * 1 / 2
}

func LineHeight(context *guigui.Context) float64 {
	return baseUnitSize * context.Scale() * 3 / 4
}

func UnitSize(context *guigui.Context) int {
	return int(baseUnitSize * context.Scale())
}

func RoundedCornerRadius(context *guigui.Context) int {
	return int(baseUnitSize * context.Scale() / 4)
}
