// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

type Sidebar struct {
	guigui.DefaultWidgetBehavior

	sidebarWidget   *guigui.Widget
	listWidget      *guigui.Widget
	listItemWidgets []*guigui.Widget
}

func sidebarWidth(context *guigui.Context) int {
	return 8 * basicwidget.UnitSize(context)
}

func (s *Sidebar) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	if s.sidebarWidget == nil {
		s.sidebarWidget = guigui.NewWidget(&basicwidget.Sidebar{})
	}
	if s.listWidget == nil {
		s.listWidget = guigui.NewWidget(&basicwidget.List{})
	}
	_, h := widget.Size(context)
	sidebar := s.sidebarWidget.Behavior().(*basicwidget.Sidebar)
	sidebar.SetSize(context, sidebarWidth(context), h)
	sidebar.SetContent(context, func(context *guigui.Context, widget *guigui.Widget, childAppender *basicwidget.ScrollablePanelChildWidgetAppender) {
		list := s.listWidget.Behavior().(*basicwidget.List)
		list.SetSize(context, sidebarWidth(context), h)
		childAppender.AppendChildWidget(s.listWidget, widget.Position())
	})
	appender.AppendChildWidget(s.sidebarWidget, widget.Position())

	if len(s.listItemWidgets) == 0 {
		{
			var t basicwidget.Text
			t.SetText("Settings")
			t.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
			s.listItemWidgets = append(s.listItemWidgets, guigui.NewWidget(&t))
		}
		{
			var t basicwidget.Text
			t.SetText("Buttons")
			t.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
			s.listItemWidgets = append(s.listItemWidgets, guigui.NewWidget(&t))
		}
	}

	list := s.listWidget.Behavior().(*basicwidget.List)
	list.SetStyle(basicwidget.ListStyleSidebar)
	var items []basicwidget.ListItem
	for _, w := range s.listItemWidgets {
		t := w.Behavior().(*basicwidget.Text)
		t.SetSize(list.ItemWidth(context, s.listWidget), basicwidget.UnitSize(context))
		items = append(items, basicwidget.ListItem{
			Content:    w,
			Selectable: true,
		})
	}
	list.SetItems(items)
}

func (s *Sidebar) Update(context *guigui.Context, widget *guigui.Widget) error {
	list := s.listWidget.Behavior().(*basicwidget.List)
	for i, w := range s.listItemWidgets {
		if list.SelectedItemIndex() == i {
			w.Behavior().(*basicwidget.Text).SetColor(basicwidget.DefaultActiveListItemTextColor(context))
		} else {
			w.Behavior().(*basicwidget.Text).SetColor(basicwidget.DefaultTextColor(context))
		}
	}
	return nil
}

func (s *Sidebar) Size(context *guigui.Context, widget *guigui.Widget) (int, int) {
	_, h := widget.Parent().Size(context)
	return sidebarWidth(context), h
}
