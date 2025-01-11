// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import "image"

type ChildWidgetAppender struct {
	app    *app
	widget *Widget
}

func (c *ChildWidgetAppender) AppendChildWidget(widgetBehavior WidgetBehavior, position image.Point) {
	widget := widgetBehavior.internalWidget(widgetBehavior)
	widget.parent = c.widget
	widget.position = position
	if widget.behavior.IsPopup() {
		widget.visibleBounds = widget.bounds()
	} else {
		widget.visibleBounds = c.widget.visibleBounds.Intersect(widget.bounds())
	}
	c.widget.children = append(c.widget.children, widget)
}
