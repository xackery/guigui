// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"image"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/hajimehoshi/guigui"
)

type GroupItem struct {
	PrimaryWidget   *guigui.Widget
	SecondaryWidget *guigui.Widget
}

type Group struct {
	guigui.DefaultWidgetBehavior

	items []*GroupItem

	widthMinusDefault int
}

func (g *Group) SetItems(items []*GroupItem) {
	g.items = slices.Delete(g.items, 0, len(g.items))
	g.items = append(g.items, items...)
}

func (g *Group) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	for i, item := range g.items {
		if item.PrimaryWidget == nil && item.SecondaryWidget == nil {
			continue
		}
		bounds := g.itemBounds(context, widget, i)
		appender.AppendChildWidgetWithBounds(item.PrimaryWidget, bounds)
		w, _ := item.SecondaryWidget.Size(context)
		bounds.Min.X = bounds.Max.X - w
		appender.AppendChildWidgetWithBounds(item.SecondaryWidget, bounds)
	}
}

func (g *Group) itemBounds(context *guigui.Context, widget *guigui.Widget, childIndex int) image.Rectangle {
	var y int
	for i, item := range g.items {
		if i > childIndex {
			return image.Rectangle{}
		}
		if item.PrimaryWidget == nil && item.SecondaryWidget == nil {
			continue
		}
		if !item.SecondaryWidget.IsVisible() {
			continue
		}
		_, kh := item.PrimaryWidget.Size(context)
		_, vh := item.SecondaryWidget.Size(context)
		h := max(kh, vh, int(LineHeight(context)))
		if i == childIndex {
			bounds := g.bounds(context, widget)
			bounds.Min.X += UnitSize(context) / 2
			bounds.Max.X -= UnitSize(context) / 2
			bounds.Min.Y += y + UnitSize(context)/2
			bounds.Max.Y = bounds.Min.Y + h
			return bounds
		}
		y += h + UnitSize(context)
	}

	return image.Rectangle{}
}

func (g *Group) Draw(context *guigui.Context, widget *guigui.Widget, dst *ebiten.Image) {
	bounds := g.bounds(context, widget)
	bounds.Max.Y = bounds.Min.Y + g.height(context)
	DrawRoundedRect(context, dst, bounds, Color(context.ColorMode(), ColorTypeBase, 0.925), RoundedCornerRadius(context))

	if len(g.items) > 0 {
		for i := range g.items[:len(g.items)-1] {
			b := g.itemBounds(context, widget, i)
			x0 := float32(bounds.Min.X + UnitSize(context)/2)
			x1 := float32(bounds.Max.X - UnitSize(context)/2)
			y := float32(b.Max.Y) + float32(UnitSize(context))/2
			width := 1 * float32(context.Scale())
			clr := Color(context.ColorMode(), ColorTypeBase, 0.875)
			vector.StrokeLine(dst, x0, y, x1, y, width, clr, false)
		}
	}

	DrawRoundedRectBorder(context, dst, bounds, Color(context.ColorMode(), ColorTypeBase, 0.875), RoundedCornerRadius(context), 1*float32(context.Scale()), RoundedRectBorderTypeRegular)
}

func (g *Group) SetWidth(context *guigui.Context, width int) {
	g.widthMinusDefault = width - defaultGroupWidth(context)
}

func (g *Group) Size(context *guigui.Context, widget *guigui.Widget) (int, int) {
	return g.widthMinusDefault + defaultGroupWidth(context), g.height(context)
}

func defaultGroupWidth(context *guigui.Context) int {
	return 6 * UnitSize(context)
}

func (g *Group) height(context *guigui.Context) int {
	var y int
	for _, item := range g.items {
		if (item.PrimaryWidget == nil || !item.PrimaryWidget.IsVisible()) &&
			(item.SecondaryWidget == nil || !item.SecondaryWidget.IsVisible()) {
			continue
		}
		_, kh := item.PrimaryWidget.Size(context)
		_, vh := item.SecondaryWidget.Size(context)
		h := max(kh, vh, int(LineHeight(context)))
		y += h + UnitSize(context)
	}
	return y
}

func (g *Group) bounds(context *guigui.Context, widget *guigui.Widget) image.Rectangle {
	bounds := widget.Bounds()
	bounds.Max.X = bounds.Min.X + g.widthMinusDefault + defaultGroupWidth(context)
	bounds.Max.Y = bounds.Min.Y + g.height(context)
	return bounds
}
