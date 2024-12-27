// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package font

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"fmt"

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

	basicwidget.RegisterFaceSource(queryFaceSources)
}

func queryFaceSources(lang language.Tag) ([]basicwidget.FaceSourceQueryResult, error) {
	primaryID, primaryIDScore := langToFaceID(lang)

	ids := []faceID{faceSC, faceTC, faceHK, faceJP, faceKR}
	rs := make([]basicwidget.FaceSourceQueryResult, 0, len(ids))
	for _, id := range ids {
		var score float64
		if id == primaryID {
			score = primaryIDScore
		}
		rs = append(rs, basicwidget.FaceSourceQueryResult{
			FaceSource: faceSources[id],
			Score:      score,
		})
	}
	return rs, nil
}

func langToFaceID(lang language.Tag) (faceID, float64) {
	switch base, _ := lang.Base(); base.String() {
	case "ja":
		return faceJP, 1
	case "ko":
		return faceKR, 1
	case "zh":
		script, _ := lang.Script()
		region, _ := lang.Region()
		switch script.String() {
		case "Hans":
			return faceSC, 1
		case "Hant":
			if region.String() == "HK" {
				return faceHK, 1
			}
			return faceTC, 1
		}
		switch region.String() {
		case "HK":
			return faceHK, 1
		case "MO", "TW":
			return faceTC, 0.5
		}
		return faceSC, 0.5
	}
	return 0, 0
}
