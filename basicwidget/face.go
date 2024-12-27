// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/text/language"
)

//go:generate go run gen.go

//go:embed NotoSans.ttf.gz
var notoSansTTFGz []byte

type FaceSourceQueryResult struct {
	FaceSource *text.GoTextFaceSource
	Priority   float64
}

type faceSourceWithPriority struct {
	faceSource *text.GoTextFaceSource
	priority   func(locale language.Tag) float64
}

var faceSourceWithPriorities []faceSourceWithPriority

func RegisterFaceSource(faceSource *text.GoTextFaceSource, priority func(locale language.Tag) float64) {
	faceSourceWithPriorities = append(faceSourceWithPriorities, faceSourceWithPriority{
		faceSource: faceSource,
		priority:   priority,
	})
}

func init() {
	r, err := gzip.NewReader(bytes.NewReader(notoSansTTFGz))
	if err != nil {
		panic(err)
	}
	defer r.Close()
	f, err := text.NewGoTextFaceSource(r)
	if err != nil {
		panic(err)
	}
	RegisterFaceSource(f, func(locale language.Tag) float64 {
		script, conf := locale.Script()
		if script == language.MustParseScript("Latn") || script == language.MustParseScript("Grek") || script == language.MustParseScript("Cyrl") {
			switch conf {
			case language.Exact, language.High:
				return 1
			case language.Low:
				return 0.5
			}
		}
		return 0
	})
}

var (
	faceCache map[faceCacheKey]text.Face
)

type faceCacheKey struct {
	size   float64
	weight text.Weight
	langs  string
}

func fontFace(size float64, weight text.Weight, locales []language.Tag) text.Face {
	var langStrs []string
	for _, l := range locales {
		langStrs = append(langStrs, l.String())
	}

	key := faceCacheKey{
		size:   size,
		weight: weight,
		langs:  strings.Join(langStrs, ","),
	}
	if f, ok := faceCache[key]; ok {
		return f
	}

	fps := append([]faceSourceWithPriority{}, faceSourceWithPriorities...)

	var faceSources []*text.GoTextFaceSource
	for _, l := range locales {
		var highestPriority float64
		var index int
		for i, fp := range fps {
			p := min(max(fp.priority(l), 0), 1)
			// If the priority is the same, the later one is used.
			if highestPriority <= p {
				highestPriority = p
				index = i
			}
		}
		faceSources = append(faceSources, fps[index].faceSource)
		fps = append(fps[:index], fps[index+1:]...)
	}

	var fs []text.Face
	var lang language.Tag
	if len(locales) > 0 {
		lang = locales[0]
	}
	for _, faceSource := range faceSources {
		f := &text.GoTextFace{
			Source:   faceSource,
			Size:     size,
			Language: lang,
		}
		f.SetVariation(text.MustParseTag("wght"), float32(weight))
		fs = append(fs, f)
	}
	mf, err := text.NewMultiFace(fs...)
	if err != nil {
		panic(err)
	}

	if faceCache == nil {
		faceCache = map[faceCacheKey]text.Face{}
	}
	faceCache[key] = mf

	return mf
}
