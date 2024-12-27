// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"bytes"
	"compress/gzip"
	_ "embed"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/guigui/internal/locale"
	"golang.org/x/text/language"
)

//go:generate go run gen.go

//go:embed NotoSans.ttf.gz
var notoSansTTFGz []byte

var locales []language.Tag

func init() {
	ls, err := locale.Locales()
	if err != nil {
		panic(err)
	}
	locales = ls
}

type FaceSourcesGetter func(lang []language.Tag) ([]*text.GoTextFaceSource, error)

var faceSourcesGetters []FaceSourcesGetter

func RegisterFaceSource(f FaceSourcesGetter) {
	faceSourcesGetters = append(faceSourcesGetters, f)
}

var defaultFaceSource *text.GoTextFaceSource

func getDefaultFaceSource(lang []language.Tag) ([]*text.GoTextFaceSource, error) {
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
	return []*text.GoTextFaceSource{defaultFaceSource}, nil
}

func init() {
	RegisterFaceSource(getDefaultFaceSource)
}

var (
	faceCache map[faceCacheKey]text.Face
)

type faceCacheKey struct {
	size   float64
	weight text.Weight
	lang   language.Tag
}

func fontFace(size float64, weight text.Weight, lang language.Tag) text.Face {
	key := faceCacheKey{
		size:   size,
		weight: weight,
		lang:   lang,
	}
	if f, ok := faceCache[key]; ok {
		return f
	}

	var langs []language.Tag
	if lang != language.Und {
		langs = append(langs, lang)
	}
	langs = append(langs, locales...)

	var faceSources []*text.GoTextFaceSource
	for _, f := range faceSourcesGetters {
		fs, err := f(langs)
		if err != nil {
			panic(err)
		}
		faceSources = append(faceSources, fs...)
	}

	var fs []text.Face
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
