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
	PrimaryWidget   guigui.WidgetBehavior
	SecondaryWidget guigui.WidgetBehavior
}

type Group struct {
	guigui.DefaultWidgetBehavior

	items []*GroupItem

	widthMinusDefault int
}

func groupItemPadding(context *guigui.Context) (int, int) {
	return UnitSize(context) / 2, UnitSize(context) / 2
}

func (g *Group) SetItems(items []*GroupItem) {
	g.items = slices.Delete(g.items, 0, len(g.items))
	g.items = append(g.items, items...)
}

func (g *Group) AppendChildWidgets(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	for i, item := range g.items {
		if item.PrimaryWidget == nil && item.SecondaryWidget == nil {
			continue
		}
		bounds := g.itemBounds(context, i)
		appender.AppendChildWidget(item.PrimaryWidget, bounds.Min)
		w, _ := item.SecondaryWidget.Size(context)
		bounds.Min.X = bounds.Max.X - w
		appender.AppendChildWidget(item.SecondaryWidget, bounds.Min)
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
		if !context.WidgetFromBehavior(item.SecondaryWidget).IsVisible() {
			continue
		}
		_, kh := item.PrimaryWidget.Size(context)
		_, vh := item.SecondaryWidget.Size(context)
		h := max(kh, vh, int(LineHeight(context)))
		if i == childIndex {
			bounds := g.bounds(context)
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
	bounds := g.bounds(context)
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
		if (item.PrimaryWidget == nil || !context.WidgetFromBehavior(item.PrimaryWidget).IsVisible()) &&
			(item.SecondaryWidget == nil || !context.WidgetFromBehavior(item.SecondaryWidget).IsVisible()) {
			continue
		}
		_, kh := item.PrimaryWidget.Size(context)
		_, vh := item.SecondaryWidget.Size(context)
		h := max(kh, vh, int(LineHeight(context)))
		y += h + 2*paddingY
	}
	return y
}

func (g *Group) bounds(context *guigui.Context) image.Rectangle {
	p := context.WidgetFromBehavior(g).Position()
	w, h := g.Size(context)
	return image.Rectangle{
		Min: p,
		Max: p.Add(image.Pt(w, h)),
	}
}
