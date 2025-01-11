// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import "image"

type ChildWidgetAppender struct {
	app    *app
	widget *widgetState
}

func (c *ChildWidgetAppender) AppendChildWidget(widget Widget, position image.Point) {
	widgetState := widget.widgetState(widget)
	widgetState.parent = c.widget
	widgetState.position = position
	if widget.IsPopup() {
		widgetState.visibleBounds = widgetState.bounds()
	} else {
		widgetState.visibleBounds = c.widget.visibleBounds.Intersect(widgetState.bounds())
	}
	c.widget.children = append(c.widget.children, widgetState)
}
