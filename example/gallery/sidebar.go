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
	listItemWidgets []basicwidget.ListItem
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

	list := s.listWidget.Behavior().(*basicwidget.List)
	list.SetStyle(basicwidget.ListStyleSidebar)
	if len(s.listItemWidgets) == 0 {
		{
			var t basicwidget.Text
			t.SetText("Settings")
			t.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
			t.SetSize(list.ItemWidth(context, s.listWidget), basicwidget.UnitSize(context))
			s.listItemWidgets = append(s.listItemWidgets, basicwidget.ListItem{
				Content:    guigui.NewWidget(&t),
				Selectable: true,
				Tag:        "settings",
			})
		}
		{
			var t basicwidget.Text
			t.SetText("Basic")
			t.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
			t.SetSize(list.ItemWidth(context, s.listWidget), basicwidget.UnitSize(context))
			s.listItemWidgets = append(s.listItemWidgets, basicwidget.ListItem{
				Content:    guigui.NewWidget(&t),
				Selectable: true,
				Tag:        "basic",
			})
		}
		{
			var t basicwidget.Text
			t.SetText("Lists")
			t.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
			t.SetSize(list.ItemWidth(context, s.listWidget), basicwidget.UnitSize(context))
			s.listItemWidgets = append(s.listItemWidgets, basicwidget.ListItem{
				Content:    guigui.NewWidget(&t),
				Selectable: true,
				Tag:        "lists",
			})
		}
	}
	list.SetItems(s.listItemWidgets)
}

func (s *Sidebar) Update(context *guigui.Context, widget *guigui.Widget) error {
	list := s.listWidget.Behavior().(*basicwidget.List)
	for i, w := range s.listItemWidgets {
		t := w.Content.Behavior().(*basicwidget.Text)
		if list.SelectedItemIndex() == i {
			t.SetColor(basicwidget.DefaultActiveListItemTextColor(context))
		} else {
			t.SetColor(basicwidget.DefaultTextColor(context))
		}
	}
	return nil
}

func (s *Sidebar) Size(context *guigui.Context, widget *guigui.Widget) (int, int) {
	_, h := widget.Parent().Size(context)
	return sidebarWidth(context), h
}

func (s *Sidebar) SelectedItemTag() string {
	if s.listWidget == nil {
		return ""
	}
	list := s.listWidget.Behavior().(*basicwidget.List)
	item, ok := list.SelectedItem()
	if !ok {
		return ""
	}
	return item.Tag.(string)
}
