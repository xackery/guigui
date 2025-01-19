// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"image"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

type Sidebar struct {
	guigui.DefaultWidget

	sidebar         basicwidget.Sidebar
	list            basicwidget.List
	listItemWidgets []basicwidget.ListItem
}

func sidebarWidth(context *guigui.Context) int {
	return 8 * basicwidget.UnitSize(context)
}

func (s *Sidebar) AppendChildWidgets(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	_, h := s.Size(context)
	s.sidebar.SetSize(context, sidebarWidth(context), h)
	s.sidebar.SetContent(context, func(context *guigui.Context, childAppender *basicwidget.ContainerChildWidgetAppender, offsetX, offsetY float64) {
		s.list.SetSize(context, sidebarWidth(context), h)
		guigui.SetPosition(&s.list, guigui.Position(s).Add(image.Pt(int(offsetX), int(offsetY))))
		childAppender.AppendChildWidget(&s.list)
	})
	guigui.SetPosition(&s.sidebar, guigui.Position(s))
	appender.AppendChildWidget(&s.sidebar)

	s.list.SetStyle(basicwidget.ListStyleSidebar)
	if len(s.listItemWidgets) == 0 {
		{
			var t basicwidget.Text
			t.SetText("Settings")
			t.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
			t.SetSize(s.list.ItemWidth(context), basicwidget.UnitSize(context))
			s.listItemWidgets = append(s.listItemWidgets, basicwidget.ListItem{
				Content:    &t,
				Selectable: true,
				Tag:        "settings",
			})
		}
		{
			var t basicwidget.Text
			t.SetText("Basic")
			t.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
			t.SetSize(s.list.ItemWidth(context), basicwidget.UnitSize(context))
			s.listItemWidgets = append(s.listItemWidgets, basicwidget.ListItem{
				Content:    &t,
				Selectable: true,
				Tag:        "basic",
			})
		}
		{
			var t basicwidget.Text
			t.SetText("Buttons")
			t.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
			t.SetSize(s.list.ItemWidth(context), basicwidget.UnitSize(context))
			s.listItemWidgets = append(s.listItemWidgets, basicwidget.ListItem{
				Content:    &t,
				Selectable: true,
				Tag:        "buttons",
			})
		}
		{
			var t basicwidget.Text
			t.SetText("Lists")
			t.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
			t.SetSize(s.list.ItemWidth(context), basicwidget.UnitSize(context))
			s.listItemWidgets = append(s.listItemWidgets, basicwidget.ListItem{
				Content:    &t,
				Selectable: true,
				Tag:        "lists",
			})
		}
		{
			var t basicwidget.Text
			t.SetText("Popups")
			t.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
			t.SetSize(s.list.ItemWidth(context), basicwidget.UnitSize(context))
			s.listItemWidgets = append(s.listItemWidgets, basicwidget.ListItem{
				Content:    &t,
				Selectable: true,
				Tag:        "popups",
			})
		}
	}
	s.list.SetItems(s.listItemWidgets)
}

func (s *Sidebar) Update(context *guigui.Context) error {
	for i, w := range s.listItemWidgets {
		t := w.Content.(*basicwidget.Text)
		if s.list.SelectedItemIndex() == i {
			t.SetColor(basicwidget.DefaultActiveListItemTextColor(context))
		} else {
			t.SetColor(basicwidget.DefaultTextColor(context))
		}
	}
	return nil
}

func (s *Sidebar) Size(context *guigui.Context) (int, int) {
	_, h := guigui.Parent(s).Size(context)
	return sidebarWidth(context), h
}

func (s *Sidebar) SelectedItemTag() string {
	item, ok := s.list.SelectedItem()
	if !ok {
		return ""
	}
	return item.Tag.(string)
}
