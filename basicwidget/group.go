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

	items             []*GroupItem
	widthMinusDefault int

	primaryBounds   []image.Rectangle
	secondaryBounds []image.Rectangle
}

func groupItemPadding(context *guigui.Context) (int, int) {
	return UnitSize(context) / 2, UnitSize(context) / 4
}

func (g *Group) SetItems(items []*GroupItem) {
	g.items = slices.Delete(g.items, 0, len(g.items))
	g.items = append(g.items, items...)
}

func (g *Group) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	g.calcItemBounds(context)

	for i, item := range g.items {
		g.primaryBounds = append(g.primaryBounds, image.Rectangle{})
		g.secondaryBounds = append(g.secondaryBounds, image.Rectangle{})

		if item.PrimaryWidget == nil && item.SecondaryWidget == nil {
			continue
		}

		if item.PrimaryWidget != nil {
			guigui.SetPosition(item.PrimaryWidget, g.primaryBounds[i].Min)
			appender.AppendChildWidget(item.PrimaryWidget)
		}
		if item.SecondaryWidget != nil {
			guigui.SetPosition(item.SecondaryWidget, g.secondaryBounds[i].Min)
			appender.AppendChildWidget(item.SecondaryWidget)
		}
	}
}

func (g *Group) calcItemBounds(context *guigui.Context) {
	g.primaryBounds = slices.Delete(g.primaryBounds, 0, len(g.primaryBounds))
	g.secondaryBounds = slices.Delete(g.secondaryBounds, 0, len(g.secondaryBounds))

	paddingX, paddingY := groupItemPadding(context)

	var y int
	for i, item := range g.items {
		g.primaryBounds = append(g.primaryBounds, image.Rectangle{})
		g.secondaryBounds = append(g.secondaryBounds, image.Rectangle{})

		if item.PrimaryWidget == nil && item.SecondaryWidget == nil {
			continue
		}
		if !guigui.IsVisible(item.SecondaryWidget) {
			continue
		}

		var primaryH int
		var secondaryH int
		if item.PrimaryWidget != nil {
			_, primaryH = item.PrimaryWidget.Size(context)
		}
		if item.SecondaryWidget != nil {
			_, secondaryH = item.SecondaryWidget.Size(context)
		}
		h := max(primaryH, secondaryH, g.minItemHeight(context))
		baseBounds := guigui.Bounds(g)
		baseBounds.Min.X += paddingX
		baseBounds.Max.X -= paddingX
		baseBounds.Min.Y += y
		baseBounds.Max.Y = baseBounds.Min.Y + h

		if item.PrimaryWidget != nil {
			bounds := baseBounds
			ww, wh := item.PrimaryWidget.Size(context)
			bounds.Max.X = bounds.Min.X + ww
			pY := (h + 2*paddingY - wh) / 2
			if wh < UnitSize(context)+2*paddingY {
				pY = min(pY, max(0, (UnitSize(context)+2*paddingY-wh)/2))
			}
			bounds.Min.Y += pY
			bounds.Max.Y += pY
			g.primaryBounds[i] = bounds
		}
		if item.SecondaryWidget != nil {
			bounds := baseBounds
			ww, wh := item.SecondaryWidget.Size(context)
			bounds.Min.X = bounds.Max.X - ww
			pY := (h + 2*paddingY - wh) / 2
			if wh < UnitSize(context)+2*paddingY {
				pY = min(pY, (UnitSize(context)+2*paddingY-wh)/2)
			}
			bounds.Min.Y += pY
			bounds.Max.Y += pY
			g.secondaryBounds[i] = bounds
		}

		y += h + 2*paddingY
	}
}

func (g *Group) Draw(context *guigui.Context, dst *ebiten.Image) {
	bounds := guigui.Bounds(g)
	bounds.Max.Y = bounds.Min.Y + g.height(context)
	DrawRoundedRect(context, dst, bounds, Color(context.ColorMode(), ColorTypeBase, 0.925), RoundedCornerRadius(context))

	if len(g.items) > 0 {
		paddingX, paddingY := groupItemPadding(context)
		y := paddingY
		for _, item := range g.items[:len(g.items)-1] {
			var primaryH int
			var secondaryH int
			if item.PrimaryWidget != nil {
				_, primaryH = item.PrimaryWidget.Size(context)
			}
			if item.SecondaryWidget != nil {
				_, secondaryH = item.SecondaryWidget.Size(context)
			}
			h := max(primaryH, secondaryH, g.minItemHeight(context))
			y += h + 2*paddingY

			x0 := float32(bounds.Min.X + paddingX)
			x1 := float32(bounds.Max.X - paddingX)
			y := float32(y) + float32(paddingY)
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
		var primaryH int
		var secondaryH int
		if item.PrimaryWidget != nil {
			_, primaryH = item.PrimaryWidget.Size(context)
		}
		if item.SecondaryWidget != nil {
			_, secondaryH = item.SecondaryWidget.Size(context)
		}
		h := max(primaryH, secondaryH, g.minItemHeight(context))
		y += h + 2*paddingY
	}
	return y
}

func (g *Group) minItemHeight(context *guigui.Context) int {
	return UnitSize(context)
}
