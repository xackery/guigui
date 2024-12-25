// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"bytes"
	"compress/gzip"
	_ "embed"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/text/language"
)

//go:generate go run gen.go

//go:embed NotoSans.ttf.gz
var notoSansTTFGz []byte

type FaceSourcesGetter func(lang language.Tag) ([]*text.GoTextFaceSource, error)

var faceSourcesGetters []FaceSourcesGetter

func RegisterFaceSource(f FaceSourcesGetter) {
	faceSourcesGetters = append(faceSourcesGetters, f)
}

var defaultFaceSource *text.GoTextFaceSource

func getDefaultFaceSource(lang language.Tag) ([]*text.GoTextFaceSource, error) {
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

func FontFace(size float64, weight text.Weight, lang language.Tag) text.Face {
	key := faceCacheKey{
		size:   size,
		weight: weight,
		lang:   lang,
	}
	if f, ok := faceCache[key]; ok {
		return f
	}

	var faceSources []*text.GoTextFaceSource
	for i := len(faceSourcesGetters) - 1; i >= 0; i-- {
		fs, err := faceSourcesGetters[i](lang)
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
