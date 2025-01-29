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
	PrimaryWidget   guigui.Widget
	SecondaryWidget guigui.Widget
}

type Group struct {
	guigui.DefaultWidget

	items []*GroupItem

	widthMinusDefault int
}

func groupItemPadding(context *guigui.Context) (int, int) {
	return UnitSize(context) / 2, UnitSize(context) / 4
}

func (g *Group) SetItems(items []*GroupItem) {
	g.items = slices.Delete(g.items, 0, len(g.items))
	g.items = append(g.items, items...)
}

func (g *Group) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	for i, item := range g.items {
		if item.PrimaryWidget == nil && item.SecondaryWidget == nil {
			continue
		}
		if item.PrimaryWidget != nil {
			bounds := g.itemBounds(context, i)
			guigui.SetPosition(item.PrimaryWidget, bounds.Min)
			appender.AppendChildWidget(item.PrimaryWidget)
		}
		if item.SecondaryWidget != nil {
			bounds := g.itemBounds(context, i)
			w, _ := item.SecondaryWidget.Size(context)
			bounds.Min.X = bounds.Max.X - w
			guigui.SetPosition(item.SecondaryWidget, bounds.Min)
			appender.AppendChildWidget(item.SecondaryWidget)
		}
	}
}

func (g *Group) itemBounds(context *guigui.Context, childIndex int) image.Rectangle {
	paddingX, paddingY := groupItemPadding(context)

	var y int
	for i, item := range g.items {
		if i > childIndex {
			return image.Rectangle{}
		}
		if item.PrimaryWidget == nil && item.SecondaryWidget == nil {
			continue
		}
		if !guigui.IsVisible(item.SecondaryWidget) {
			continue
		}
		var kh int
		var vh int
		if item.PrimaryWidget != nil {
			_, kh = item.PrimaryWidget.Size(context)
		}
		if item.SecondaryWidget != nil {
			_, vh = item.SecondaryWidget.Size(context)
		}
		h := max(kh, vh, int(LineHeight(context)))
		if i == childIndex {
			bounds := guigui.Bounds(g)
			bounds.Min.X += paddingX
			bounds.Max.X -= paddingX
			bounds.Min.Y += y + paddingY
			bounds.Max.Y = bounds.Min.Y + h
			return bounds
		}
		y += h + 2*paddingY
	}

	return image.Rectangle{}
}

func (g *Group) Draw(context *guigui.Context, dst *ebiten.Image) {
	bounds := guigui.Bounds(g)
	bounds.Max.Y = bounds.Min.Y + g.height(context)
	DrawRoundedRect(context, dst, bounds, Color(context.ColorMode(), ColorTypeBase, 0.925), RoundedCornerRadius(context))

	if len(g.items) > 0 {
		paddingX, paddingY := groupItemPadding(context)
		for i := range g.items[:len(g.items)-1] {
			b := g.itemBounds(context, i)
			x0 := float32(bounds.Min.X + paddingX)
			x1 := float32(bounds.Max.X - paddingX)
			y := float32(b.Max.Y) + float32(paddingY)
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

func (g *Group) Size(context *guigui.Context) (int, int) {
	return g.widthMinusDefault + defaultGroupWidth(context), g.height(context)
}

func defaultGroupWidth(context *guigui.Context) int {
	return 6 * UnitSize(context)
}

func (g *Group) height(context *guigui.Context) int {
	_, paddingY := groupItemPadding(context)

	var y int
	for _, item := range g.items {
		if (item.PrimaryWidget == nil || !guigui.IsVisible(item.PrimaryWidget)) &&
			(item.SecondaryWidget == nil || !guigui.IsVisible(item.SecondaryWidget)) {
			continue
		}
		var kh int
		var vh int
		if item.PrimaryWidget != nil {
			_, kh = item.PrimaryWidget.Size(context)
		}
		if item.SecondaryWidget != nil {
			_, vh = item.SecondaryWidget.Size(context)
		}
		h := max(kh, vh, int(LineHeight(context)))
		y += h + 2*paddingY
	}
	return y
}
