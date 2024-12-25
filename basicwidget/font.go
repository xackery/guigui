// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"image"
	"image/color"
	"strings"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/text/language"
)

//go:generate go run gen.go

//go:embed NotoSans.ttf.gz
var notoSansTTFGz []byte

var (
	faceSource *text.GoTextFaceSource
	faceCache  map[faceCacheKey]text.Face
)

type faceCacheKey struct {
	size   float64
	weight text.Weight
	lang   language.Tag
}

func FontFace(size float64, weight text.Weight, lang language.Tag) text.Face {
	if faceSource == nil {
		r, err := gzip.NewReader(bytes.NewReader(notoSansTTFGz))
		if err != nil {
			panic(err)
		}
		f, err := text.NewGoTextFaceSource(r)
		if err != nil {
			panic(err)
		}
		faceSource = f
	}
	key := faceCacheKey{
		size:   size,
		weight: weight,
		lang:   lang,
	}
	if f, ok := faceCache[key]; ok {
		return f
	}
	f := &text.GoTextFace{
		Source:   faceSource,
		Size:     size,
		Language: lang,
	}
	f.SetVariation(text.MustParseTag("wght"), float32(weight))
	if faceCache == nil {
		faceCache = map[faceCacheKey]text.Face{}
	}
	faceCache[key] = f
	return f
}

type HorizontalAlign int

const (
	HorizontalAlignStart HorizontalAlign = iota
	HorizontalAlignCenter
	HorizontalAlignEnd
)

type VerticalAlign int

const (
	VerticalAlignTop VerticalAlign = iota
	VerticalAlignMiddle
	VerticalAlignBottom
)

func drawText(bounds image.Rectangle, dst *ebiten.Image, str string, face text.Face, lineHeight float64, hAlign HorizontalAlign, vAlign VerticalAlign, clr color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(bounds.Min.X), float64(bounds.Min.Y))
	op.ColorScale.ScaleWithColor(clr)
	if dst.Bounds() != bounds {
		dst = dst.SubImage(bounds).(*ebiten.Image)
	}

	op.LineSpacing = lineHeight

	switch hAlign {
	case HorizontalAlignStart:
		op.PrimaryAlign = text.AlignStart
	case HorizontalAlignCenter:
		op.GeoM.Translate(float64(bounds.Dx())/2, 0)
		op.PrimaryAlign = text.AlignCenter
	case HorizontalAlignEnd:
		op.GeoM.Translate(float64(bounds.Dx()), 0)
		op.PrimaryAlign = text.AlignEnd
	}

	m := face.Metrics()
	padding := (lineHeight - (m.HAscent + m.HDescent)) / 2

	switch vAlign {
	case VerticalAlignTop:
		op.GeoM.Translate(0, padding)
		op.SecondaryAlign = text.AlignStart
	case VerticalAlignMiddle:
		op.GeoM.Translate(0, float64(bounds.Dy())/2)
		op.SecondaryAlign = text.AlignCenter
	case VerticalAlignBottom:
		op.GeoM.Translate(0, float64(bounds.Dy())-padding)
		op.SecondaryAlign = text.AlignEnd
	}

	text.Draw(dst, str, face, op)
}

func textUpperLeft(bounds image.Rectangle, str string, face text.Face, lineHeight float64, hAlign HorizontalAlign, vAlign VerticalAlign) (float64, float64) {
	w, h := text.Measure(str, face, lineHeight)
	x := float64(bounds.Min.X)
	y := float64(bounds.Min.Y)

	switch hAlign {
	case HorizontalAlignStart:
	case HorizontalAlignCenter:
		x += (float64(bounds.Dx()) - w) / 2
	case HorizontalAlignEnd:
		x += float64(bounds.Dx()) - w
	}

	m := face.Metrics()
	padding := (lineHeight - (m.HAscent + m.HDescent)) / 2

	switch vAlign {
	case VerticalAlignTop:
		y += padding
	case VerticalAlignMiddle:
		y = (float64(bounds.Dy()) - h) / 2
	case VerticalAlignBottom:
		y = float64(bounds.Dy()) - h - padding
	}

	return x, y
}

func textIndexFromPosition(textBounds image.Rectangle, x, y int, str string, face text.Face, lineHeight float64, hAlign HorizontalAlign, vAlign VerticalAlign) int {
	lines := strings.Split(str, "\n")
	if len(lines) == 0 {
		return 0
	}

	// Determine the line first.
	m := face.Metrics()
	gap := lineHeight - m.HAscent - m.HDescent
	top := float64(textBounds.Min.Y)
	n := int((float64(y) - top + gap/2) / lineHeight)
	if n < 0 {
		n = 0
	}
	if n >= len(lines) {
		n = len(lines) - 1
	}

	var idx int
	for _, l := range lines[:n] {
		idx += len(l) + 1 // 1 is for a new-line character.
	}

	// Deterine the line index.
	line := lines[n]
	left, _ := textUpperLeft(textBounds, line, face, lineHeight, hAlign, vAlign)
	var prevA float64
	var found bool
	for _, c := range visibleCulsters(line, face) {
		a := text.Advance(line[:c.EndIndexInBytes], face)
		if (float64(x) - left) < (prevA + (a-prevA)/2) {
			idx += c.StartIndexInBytes
			found = true
			break
		}
		prevA = a
	}
	if !found {
		idx += len(line)
	}

	return idx
}

func textPosition(textBounds image.Rectangle, str string, index int, face text.Face, lineHeight float64, hAlign HorizontalAlign, vAlign VerticalAlign) (x, top, bottom float64, ok bool) {
	if index < 0 || index > len(str) {
		return 0, 0, 0, false
	}

	// If the string is empty, add a dummy character to get the correct y position.
	if str == "" {
		str = " "
	}
	y := float64(textBounds.Min.Y)

	lines := strings.Split(str, "\n")
	var line string
	for _, l := range lines {
		// +1 is for \n.
		if index < len(l)+1 {
			line = l
			break
		}
		index -= len(l) + 1 // 1 is for a new-line character.
		y += lineHeight
	}

	x, _ = textUpperLeft(textBounds, line, face, lineHeight, hAlign, vAlign)
	x += text.Advance(line[:index], face)

	m := face.Metrics()
	paddingY := (lineHeight - (m.HAscent + m.HDescent)) / 2
	return x, y + paddingY, y + lineHeight - paddingY, true
}

func visibleCulsters(str string, face text.Face) []text.Glyph {
	return text.AppendGlyphs(nil, str, face, nil)
}

func logicalClusters(str string, face text.Face) []text.Glyph {
	gs := text.AppendGlyphs(nil, str, face, nil)
	result := make([]text.Glyph, 0, len(gs))

	var lastEndInBytes int
	for _, g := range gs {
		for i := range str[lastEndInBytes:g.StartIndexInBytes] {
			_, size := utf8.DecodeRuneInString(str[lastEndInBytes+i:])
			result = append(result, text.Glyph{
				StartIndexInBytes: lastEndInBytes + i,
				EndIndexInBytes:   lastEndInBytes + i + size,
			})
		}
		result = append(result, g)
		lastEndInBytes = g.EndIndexInBytes
	}

	for i := range str[lastEndInBytes:] {
		_, size := utf8.DecodeRuneInString(str[lastEndInBytes+i:])
		result = append(result, text.Glyph{
			StartIndexInBytes: lastEndInBytes + i,
			EndIndexInBytes:   lastEndInBytes + i + size,
		})
	}

	return result
}

func backspaceOnClusters(str string, face text.Face, position int) (string, int) {
	for _, c := range logicalClusters(str, face) {
		if position > c.EndIndexInBytes {
			continue
		}
		return str[:c.StartIndexInBytes] + str[c.EndIndexInBytes:], c.StartIndexInBytes
	}
	return str, position
}

func deleteOnClusters(str string, face text.Face, position int) (string, int) {
	for _, c := range logicalClusters(str, face) {
		if position > c.StartIndexInBytes {
			continue
		}
		return str[:c.StartIndexInBytes] + str[c.EndIndexInBytes:], c.StartIndexInBytes
	}
	return str, position
}

func prevPositionOnClusters(str string, face text.Face, position int) int {
	for _, c := range logicalClusters(str, face) {
		if position > c.EndIndexInBytes {
			continue
		}
		return c.StartIndexInBytes
	}
	return position
}

func nextPositionOnClusters(str string, face text.Face, position int) int {
	for _, c := range logicalClusters(str, face) {
		if position > c.StartIndexInBytes {
			continue
		}
		return c.EndIndexInBytes
	}
	return position
}
