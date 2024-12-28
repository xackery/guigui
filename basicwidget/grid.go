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
	Widget   *guigui.Widget
	Size     float64
	SizeUnit SizeUnit
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
			if item.Widget != nil {
				appender.AppendChildWidget(item.Widget, bounds.Intersect(origBounds))
			}
			bounds.Min.X = bounds.Max.X
		}
	case LinearGridDirectionVertical:
		for i, item := range l.Items {
			bounds.Max.Y = bounds.Min.Y + l.sizesInPixels[i]
			if item.Widget != nil {
				appender.AppendChildWidget(item.Widget, bounds.Intersect(origBounds))
			}
			bounds.Min.Y = bounds.Max.Y
		}
	}
}

func (l *LinearGrid) ContentSize(context *guigui.Context, widget *guigui.Widget) (int, int) {
	s, flexible := l.size(context)
	switch l.Direction {
	case LinearGridDirectionHorizontal:
		if flexible {
			s = max(s, widget.Bounds().Dx())
		}
		return s, widget.Bounds().Dy()
	case LinearGridDirectionVertical:
		if flexible {
			s = max(s, widget.Bounds().Dy())
		}
		return widget.Bounds().Dx(), s
	default:
		panic("not reached")
	}
}

func (l *LinearGrid) size(context *guigui.Context) (int, bool) {
	var flexible bool
	var sum float64
	for _, item := range l.Items {
		if item.SizeUnit == SizeUnitFraction {
			flexible = true
		}
		if item.SizeUnit == SizeUnitUnit {
			sum += item.Size
		}
	}
	return int(sum * float64(UnitSize(context))), flexible
}
