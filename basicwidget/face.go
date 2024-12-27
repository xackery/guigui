// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"bytes"
	"cmp"
	"compress/gzip"
	_ "embed"
	"slices"
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

type FaceSourcesQuerier func(locale language.Tag) ([]FaceSourceQueryResult, error)

var faceSourcesQueriers []FaceSourcesQuerier

func RegisterFaceSource(f FaceSourcesQuerier) {
	faceSourcesQueriers = append(faceSourcesQueriers, f)
}

var defaultFaceSource *text.GoTextFaceSource

func queryDefaultFaceSource(locale language.Tag) ([]FaceSourceQueryResult, error) {
	if defaultFaceSource == nil {
		r, err := gzip.NewReader(bytes.NewReader(notoSansTTFGz))
		if err != nil {
			return nil, err
		}
		defer r.Close()
		f, err := text.NewGoTextFaceSource(r)
		if err != nil {
			return nil, err
		}
		defaultFaceSource = f
	}

	var priority float64
	script, conf := locale.Script()
	if script == language.MustParseScript("Latn") || script == language.MustParseScript("Grek") || script == language.MustParseScript("Cyrl") {
		switch conf {
		case language.Exact:
			priority = 1
		case language.High:
			priority = 1
		case language.Low:
			priority = 0.5
		case language.No:
			priority = 0
		}
	}
	return []FaceSourceQueryResult{
		{
			FaceSource: defaultFaceSource,
			Priority:   priority,
		},
	}, nil
}

func init() {
	RegisterFaceSource(queryDefaultFaceSource)
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

	results := map[*text.GoTextFaceSource][]float64{}
	for i, l := range locales {
		for _, f := range faceSourcesQueriers {
			rs, err := f(l)
			if err != nil {
				panic(err)
			}
			for _, r := range rs {
				if len(results[r.FaceSource]) < i+1 {
					results[r.FaceSource] = append(results[r.FaceSource], make([]float64, i+1-len(results[r.FaceSource]))...)
				}
				results[r.FaceSource][i] = r.Priority
			}
		}
	}

	var faceSources []*text.GoTextFaceSource
	for f := range results {
		faceSources = append(faceSources, f)
	}
	slices.SortFunc(faceSources, func(fs0, fs1 *text.GoTextFaceSource) int {
		ps0 := results[fs0]
		ps1 := results[fs1]
		for i := range ps0 {
			var p0, p1 float64
			if i < len(ps0) {
				p0 = min(max(ps0[i], 0), 1)
			}
			if i < len(ps1) {
				p1 = min(max(ps1[i], 0), 1)
			}
			if p0 != p1 {
				return -cmp.Compare(p0, p1)
			}
		}
		// Deprioritize the default face source.
		if fs0 == defaultFaceSource && fs1 != defaultFaceSource {
			return 1
		}
		if fs0 != defaultFaceSource && fs1 == defaultFaceSource {
			return -1
		}
		// This is the final tie breaker.
		return cmp.Compare(fs0.Metadata().Family, fs1.Metadata().Family)
	})

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
