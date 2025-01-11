// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import "image"

type ChildWidgetAppender struct {
	app    *app
	widget Widget
}

func (c *ChildWidgetAppender) AppendChildWidget(widget Widget, position image.Point) {
	widgetState := widget.widgetState(widget)
	widgetState.parent = c.widget
	widgetState.position = position
	if widget.IsPopup() {
		widgetState.visibleBounds = widgetState.bounds()
	} else {
		widgetState.visibleBounds = c.widget.widgetState(c.widget).visibleBounds.Intersect(widgetState.bounds())
	}
	cWidgetState := c.widget.widgetState(c.widget)
	cWidgetState.children = append(cWidgetState.children, widget)
}
