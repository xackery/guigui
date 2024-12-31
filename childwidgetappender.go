// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import "image"

type ChildWidgetAppender struct {
	app    *app
	widget *Widget
}

func (c *ChildWidgetAppender) AppendChildWidget(widget *Widget, position image.Point) {
	if _, ok := c.app.currentWidgets[widget]; ok {
		panic("guigui: the widget is already in the widget tree")
	}
	if c.app.currentWidgets == nil {
		c.app.currentWidgets = map[*Widget]struct{}{}
	}
	c.app.currentWidgets[widget] = struct{}{}

	// Redraw if the child is a new one, or the bounds are changed.
	// Size might require the parent info, so set this earlier.
	widget.parent = c.widget
	w, h := widget.Size(c.app.context)
	bounds := image.Rectangle{
		Min: position,
		Max: position.Add(image.Point{w, h}),
	}
	if _, ok := c.app.prevWidgets[widget]; !ok {
		if widget.popup {
			c.app.requestRedraw(bounds)
		} else {
			c.app.requestRedraw(bounds.Intersect(c.widget.visibleBounds))
		}
	} else if !widget.bounds().Eq(bounds) {
		if widget.popup {
			c.app.requestRedraw(bounds)
			c.app.requestRedraw(widget.bounds())
		} else {
			c.app.requestRedraw(bounds.Intersect(c.widget.visibleBounds))
			c.app.requestRedraw(widget.bounds().Intersect(c.widget.visibleBounds))
		}
	}

	widget.position = position
	if widget.popup {
		widget.visibleBounds = widget.bounds()
	} else {
		widget.visibleBounds = c.widget.visibleBounds.Intersect(widget.bounds())
	}

	c.widget.children = append(c.widget.children, widget)
}
