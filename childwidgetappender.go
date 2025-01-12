// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

type ChildWidgetAppender struct {
	app    *app
	widget Widget
}

func (c *ChildWidgetAppender) AppendChildWidget(widget Widget) {
	widgetState := widget.widgetState()
	widgetState.parent = c.widget
	if widget.IsPopup() {
		widgetState.visibleBounds = Bounds(widget)
	} else {
		widgetState.visibleBounds = VisibleBounds(c.widget).Intersect(Bounds(widget))
	}
	cWidgetState := c.widget.widgetState()
	cWidgetState.children = append(cWidgetState.children, widget)
}
