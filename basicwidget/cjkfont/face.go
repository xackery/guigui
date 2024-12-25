// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package font

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"fmt"
	"slices"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/text/language"

	"github.com/hajimehoshi/guigui/basicwidget"
	"github.com/hajimehoshi/guigui/internal/locale"
)

//go:generate go run gen.go

//go:embed NotoSansCJK-VF.otf.ttc.gz
var notoSansCJKVFOTFTTCGz []byte

type baseFaceID int

const (
	baseFaceJP baseFaceID = iota
	baseFaceKR
	baseFaceSC
	baseFaceTC
	baseFaceHK
)

var faceSources = map[baseFaceID]*text.GoTextFaceSource{}

func init() {
	r, err := gzip.NewReader(bytes.NewReader(notoSansCJKVFOTFTTCGz))
	if err != nil {
		panic(err)
	}
	fs, err := text.NewGoTextFaceSourcesFromCollection(r)
	if err != nil {
		panic(err)
	}
	for _, f := range fs {
		switch f.Metadata().Family {
		case "Noto Sans CJK JP":
			faceSources[baseFaceJP] = f
		case "Noto Sans CJK KR":
			faceSources[baseFaceKR] = f
		case "Noto Sans CJK SC":
			faceSources[baseFaceSC] = f
		case "Noto Sans CJK TC":
			faceSources[baseFaceTC] = f
		case "Noto Sans CJK HK":
			faceSources[baseFaceHK] = f
		default:
			panic(fmt.Sprintf("cjkfont: unknown family: %s", f.Metadata().Family))
		}
	}

	basicwidget.RegisterFaceSource(getFaceSources)
}

func getFaceSources(lang language.Tag) ([]*text.GoTextFaceSource, error) {
	faceIDRanks := map[baseFaceID]int{}
	id, ok := langToFaceID(lang)
	if ok {
		faceIDRanks[id] = 0
	}

	// Define the rank of each face based on the OS locations.
	locales, err := locale.Locales()
	if err != nil {
		return nil, err
	}
	rank := 1
	for _, locale := range locales {
		l, err := language.Parse(locale)
		if err != nil {
			continue
		}
		f, ok := langToFaceID(l)
		if !ok {
			continue
		}
		if _, ok := faceIDRanks[f]; ok {
			continue
		}
		faceIDRanks[f] = rank
		rank++
	}

	ids := []baseFaceID{baseFaceSC, baseFaceTC, baseFaceHK, baseFaceJP, baseFaceKR}
	slices.SortStableFunc(ids, func(l1, l2 baseFaceID) int {
		if _, ok := faceIDRanks[l1]; !ok {
			return 1
		}
		if _, ok := faceIDRanks[l2]; !ok {
			return -1
		}
		return faceIDRanks[l1] - faceIDRanks[l2]
	})

	fs := make([]*text.GoTextFaceSource, 0, len(ids))
	for _, id := range ids {
		f, ok := faceSources[id]
		if !ok {
			return nil, fmt.Errorf("cjkfont: face source not found: %d", id)
		}
		fs = append(fs, f)
	}
	return fs, nil
}

func langToFaceID(lang language.Tag) (baseFaceID, bool) {
	switch base, _ := lang.Base(); base.String() {
	case "ja":
		return baseFaceJP, true
	case "ko":
		return baseFaceKR, true
	case "zh":
		script, _ := lang.Script()
		region, _ := lang.Region()
		switch script.String() {
		case "Hans":
			return baseFaceSC, true
		case "Hant":
			if region.String() == "HK" {
				return baseFaceHK, true
			}
			return baseFaceTC, true
		}
		switch region.String() {
		case "HK":
			return baseFaceHK, true
		case "MO", "TW":
			return baseFaceTC, true
		}
		return baseFaceSC, true
	}
	return 0, false
}
