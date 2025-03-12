// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package basicwidget

import (
	"iter"
	"slices"

	"github.com/xackery/guigui"
)

type ContainerChildWidgetAppender struct {
	childWidgets []guigui.Widget
}

func (c *ContainerChildWidgetAppender) AppendChildWidget(widget guigui.Widget) {
	c.childWidgets = append(c.childWidgets, widget)
}

func (c *ContainerChildWidgetAppender) iter() iter.Seq2[int, guigui.Widget] {
	return slices.All(c.childWidgets)
}

func (c *ContainerChildWidgetAppender) reset() {
	c.childWidgets = slices.Delete(c.childWidgets, 0, len(c.childWidgets))
}
