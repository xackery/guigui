// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import (
	"fmt"
	"log/slog"
	"os"
)

type ColorMode int

var defaultColorMode ColorMode = ColorModeLight

func init() {
	// TODO: Consider the system color mode.
	switch mode := os.Getenv("GUIGUI_COLOR_MODE"); mode {
	case "light", "":
		defaultColorMode = ColorModeLight
	case "dark":
		defaultColorMode = ColorModeDark
	default:
		slog.Warn(fmt.Sprintf("invalid GUIGUI_COLOR_MODE: %s", mode))
	}
}

const (
	ColorModeLight ColorMode = iota
	ColorModeDark
)

type Context struct {
	app *app

	deviceScale    float64
	appScaleMinus1 float64
	colorMode      ColorMode
	hasColorMode   bool
}

func (c *Context) Scale() float64 {
	return c.deviceScale * c.AppScale()
}

func (c *Context) DeviceScale() float64 {
	return c.deviceScale
}

func (c *Context) setDeviceScale(deviceScale float64) {
	if c.deviceScale == deviceScale {
		return
	}
	c.deviceScale = deviceScale
	c.app.requestRedraw(c.app.bounds())
}

func (c *Context) AppScale() float64 {
	return c.appScaleMinus1 + 1
}

func (c *Context) SetAppScale(scale float64) {
	if c.appScaleMinus1 == scale-1 {
		return
	}
	c.appScaleMinus1 = scale - 1
	c.app.requestRedraw(c.app.bounds())
}

func (c *Context) ColorMode() ColorMode {
	if c.hasColorMode {
		return c.colorMode
	}
	return defaultColorMode
}

func (c *Context) SetColorMode(mode ColorMode) {
	if c.hasColorMode && mode == c.colorMode {
		return
	}

	c.colorMode = mode
	c.hasColorMode = true
	c.app.requestRedraw(c.app.bounds())
}

func (c *Context) ResetColorMode() {
	c.hasColorMode = false
}
