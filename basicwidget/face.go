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
	Score      float64
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

	var score float64
	script, conf := locale.Script()
	if script == language.MustParseScript("Latn") || script == language.MustParseScript("Grek") || script == language.MustParseScript("Cyrl") {
		switch conf {
		case language.Exact:
			score = 1
		case language.High:
			score = 1
		case language.Low:
			score = 0.5
		case language.No:
			score = 0
		}
	}
	return []FaceSourceQueryResult{
		{
			FaceSource: defaultFaceSource,
			Score:      score,
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
				results[r.FaceSource][i] = r.Score
			}
		}
	}

	var faceSources []*text.GoTextFaceSource
	for f := range results {
		faceSources = append(faceSources, f)
	}
	slices.SortFunc(faceSources, func(fs0, fs1 *text.GoTextFaceSource) int {
		scores0 := results[fs0]
		scores1 := results[fs1]
		for i := range scores0 {
			var score0, score1 float64
			if i < len(scores0) {
				score0 = scores0[i]
			}
			if i < len(scores1) {
				score1 = scores1[i]
			}
			if score0 != score1 {
				return -cmp.Compare(score0, score1)
			}
		}
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
