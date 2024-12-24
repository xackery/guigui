// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"slices"

	"github.com/hajimehoshi/guigui"
)

type SizeUnit int

const (
	SizeUnitUnit SizeUnit = iota
	SizeUnitFraction
)

type LinearGridItem struct {
	Widget        *guigui.Widget
	Size          float64
	SizeUnit      SizeUnit
	PaddingLeft   float64
	PaddingRight  float64
	PaddingTop    float64
	PaddingBottom float64
}

type LinearGridDirection int

const (
	LinearGridDirectionHorizontal LinearGridDirection = iota
	LinearGridDirectionVertical
)

type LinearGrid struct {
	guigui.DefaultWidgetBehavior

	Direction LinearGridDirection
	Items     []LinearGridItem

	sizesInPixels []int
}

func (l *LinearGrid) AppendChildWidgets(context *guigui.Context, widget *guigui.Widget, appender *guigui.ChildWidgetAppender) {
	bounds := widget.Bounds()
	origBounds := bounds

	l.sizesInPixels = slices.Delete(l.sizesInPixels, 0, len(l.sizesInPixels))
	l.sizesInPixels = append(l.sizesInPixels, make([]int, len(l.Items))...)

	var remainingSize int
	var totalFraction float64

	switch l.Direction {
	case LinearGridDirectionHorizontal:
		remainingSize = bounds.Dx()
	case LinearGridDirectionVertical:
		remainingSize = bounds.Dy()
	}

	for i, item := range l.Items {
		if item.Size == 0 {
			continue
		}
		if item.SizeUnit == SizeUnitUnit {
			s := int(float64(UnitSize(context)) * item.Size)
			l.sizesInPixels[i] = s
			remainingSize -= s
		}
		if item.SizeUnit == SizeUnitFraction {
			totalFraction += item.Size
		}
	}
	if remainingSize > 0 && totalFraction > 0 {
		for i, item := range l.Items {
			if item.SizeUnit == SizeUnitFraction {
				l.sizesInPixels[i] = int(float64(remainingSize) * item.Size / totalFraction)
			}
		}
	}

	switch l.Direction {
	case LinearGridDirectionHorizontal:
		for i, item := range l.Items {
			bounds.Max.X = bounds.Min.X + l.sizesInPixels[i]

			b := bounds
			b.Min.X += int(float64(UnitSize(context)) * item.PaddingLeft)
			b.Max.X -= int(float64(UnitSize(context)) * item.PaddingRight)
			b.Min.Y += int(float64(UnitSize(context)) * item.PaddingTop)
			b.Max.Y -= int(float64(UnitSize(context)) * item.PaddingBottom)
			if item.Widget != nil {
				appender.AppendChildWidget(item.Widget, b.Intersect(origBounds))
			}

			bounds.Min.X = bounds.Max.X
		}
	case LinearGridDirectionVertical:
		for i, item := range l.Items {
			bounds.Max.Y = bounds.Min.Y + l.sizesInPixels[i]

			b := bounds
			b.Min.X += int(float64(UnitSize(context)) * item.PaddingLeft)
			b.Max.X -= int(float64(UnitSize(context)) * item.PaddingRight)
			b.Min.Y += int(float64(UnitSize(context)) * item.PaddingTop)
			b.Max.Y -= int(float64(UnitSize(context)) * item.PaddingBottom)
			if item.Widget != nil {
				appender.AppendChildWidget(item.Widget, b.Intersect(origBounds))
			}

			bounds.Min.Y = bounds.Max.Y
		}
	}
}

func (l *LinearGrid) MinimumSize(context *guigui.Context) int {
	var sum float64
	for _, item := range l.Items {
		if item.SizeUnit == SizeUnitUnit {
			sum += item.Size
		}
		switch l.Direction {
		case LinearGridDirectionHorizontal:
			sum += item.PaddingLeft + item.PaddingRight
		case LinearGridDirectionVertical:
			sum += item.PaddingTop + item.PaddingBottom
		}
	}
	return int(sum * float64(UnitSize(context)))
}
