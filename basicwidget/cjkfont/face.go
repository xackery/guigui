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
)

//go:generate go run gen.go

//go:embed NotoSansCJK-VF.otf.ttc.gz
var notoSansCJKVFOTFTTCGz []byte

type faceID int

const (
	faceJP faceID = iota
	faceKR
	faceSC
	faceTC
	faceHK
)

var faceSources = map[faceID]*text.GoTextFaceSource{}

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
			faceSources[faceJP] = f
		case "Noto Sans CJK KR":
			faceSources[faceKR] = f
		case "Noto Sans CJK SC":
			faceSources[faceSC] = f
		case "Noto Sans CJK TC":
			faceSources[faceTC] = f
		case "Noto Sans CJK HK":
			faceSources[faceHK] = f
		default:
			panic(fmt.Sprintf("cjkfont: unknown family: %s", f.Metadata().Family))
		}
	}

	basicwidget.RegisterFaceSource(getFaceSources)
}

func getFaceSources(langs []language.Tag) ([]*text.GoTextFaceSource, error) {
	faceIDRanks := map[faceID]int{}
	var rank int
	for _, l := range langs {
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

	ids := []faceID{faceSC, faceTC, faceHK, faceJP, faceKR}
	slices.SortStableFunc(ids, func(l1, l2 faceID) int {
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

func langToFaceID(lang language.Tag) (faceID, bool) {
	switch base, _ := lang.Base(); base.String() {
	case "ja":
		return faceJP, true
	case "ko":
		return faceKR, true
	case "zh":
		script, _ := lang.Script()
		region, _ := lang.Region()
		switch script.String() {
		case "Hans":
			return faceSC, true
		case "Hant":
			if region.String() == "HK" {
				return faceHK, true
			}
			return faceTC, true
		}
		switch region.String() {
		case "HK":
			return faceHK, true
		case "MO", "TW":
			return faceTC, true
		}
		return faceSC, true
	}
	return 0, false
}
