// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import (
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/hajimehoshi/guigui/internal/locale"
	"golang.org/x/text/language"
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

var envLocales []language.Tag

func init() {
	for _, tag := range strings.Split(os.Getenv("GUIGUI_LOCALES"), ",") {
		l, err := language.Parse(tag)
		if err != nil {
			slog.Warn(fmt.Sprintf("invalid GUIGUI_LOCALES: %s", tag))
			continue
		}
		envLocales = append(envLocales, l)
	}
}

var systemLocales []language.Tag

func init() {
	ls, err := locale.Locales()
	if err != nil {
		slog.Error(err.Error())
		return
	}
	systemLocales = ls
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
	locales        []language.Tag
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

func (c *Context) AppendLocales(locales []language.Tag) []language.Tag {
	origLen := len(locales)
	// App locales
	for _, l := range c.locales {
		if slices.Contains(locales[origLen:], l) {
			continue
		}
		locales = append(locales, l)
	}
	// Env locales
	for _, l := range envLocales {
		if slices.Contains(locales[origLen:], l) {
			continue
		}
		locales = append(locales, l)
	}
	// System locales
	for _, l := range systemLocales {
		if slices.Contains(locales[origLen:], l) {
			continue
		}
		locales = append(locales, l)
	}
	return locales
}

func (c *Context) AppendAppLocales(locales []language.Tag) []language.Tag {
	origLen := len(locales)
	for _, l := range c.locales {
		if slices.Contains(locales[origLen:], l) {
			continue
		}
		locales = append(locales, l)
	}
	return locales
}

func (c *Context) SetAppLocales(locales []language.Tag) {
	if slices.Equal(c.locales, locales) {
		return
	}

	c.locales = append([]language.Tag(nil), locales...)
	c.app.requestRedraw(c.app.bounds())
}

func (c *Context) WidgetFromBehavior(behavior WidgetBehavior) *Widget {
	return widgetFromBehavior(behavior)
}

func widgetFromBehavior(behavior WidgetBehavior) *Widget {
	w := behavior.internalWidget()
	w.behavior = behavior
	return w
}
