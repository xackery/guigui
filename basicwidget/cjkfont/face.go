// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package cjkfont

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

func preferTC(region language.Region) bool {
	switch region {
	case language.MustParseRegion("MO"), language.MustParseRegion("TW"):
		return true
	}
	return false
}

func preferHK(region language.Region) bool {
	switch region {
	case language.MustParseRegion("HK"):
		return true
	}
	return false
}

func init() {
	r, err := gzip.NewReader(bytes.NewReader(notoSansCJKVFOTFTTCGz))
	if err != nil {
		panic(err)
	}
	fs, err := text.NewGoTextFaceSourcesFromCollection(r)
	if err != nil {
		panic(err)
	}
	var (
		faceSC *text.GoTextFaceSource
		faceTC *text.GoTextFaceSource
		faceHK *text.GoTextFaceSource
		faceJP *text.GoTextFaceSource
		faceKR *text.GoTextFaceSource
	)
	for _, f := range fs {
		switch f.Metadata().Family {
		case "Noto Sans CJK SC":
			faceSC = f
		case "Noto Sans CJK TC":
			faceTC = f
		case "Noto Sans CJK HK":
			faceHK = f
		case "Noto Sans CJK JP":
			faceJP = f
		case "Noto Sans CJK KR":
			faceKR = f
		default:
			panic(fmt.Sprintf("cjkfont: unknown family: %s", f.Metadata().Family))
		}
	}

	basicwidget.RegisterFaceSource(faceSC, faceSCPriority)
	basicwidget.RegisterFaceSource(faceTC, faceTCPriority)
	basicwidget.RegisterFaceSource(faceHK, faceHKPriority)
	basicwidget.RegisterFaceSource(faceJP, faceJPPriority)
	basicwidget.RegisterFaceSource(faceKR, faceKRPriority)
}

func faceSCPriority(locale language.Tag) float64 {
	if script, conf := locale.Script(); conf > language.No {
		switch script {
		case language.MustParseScript("Hans"):
			switch conf {
			case language.Exact, language.High:
				return 1
			case language.Low:
				// As a special case, if only `zh` is specified, prefer SC.
				if base, conf := locale.Base(); base == language.MustParseBase("zh") && conf > language.No {
					return 1
				}
				return 0.5
			}
		case language.MustParseScript("Hant"):
			return 0.5
		}
	}
	return 0
}

func faceTCPriority(locale language.Tag) float64 {
	if script, conf := locale.Script(); conf > language.No {
		switch script {
		case language.MustParseScript("Hans"):
			return 0.5
		case language.MustParseScript("Hant"):
			if region, conf := locale.Region(); conf > language.No {
				if preferTC(region) {
					return 1
				}
				if preferHK(region) {
					return 0.5
				}
			}
			switch conf {
			case language.Exact, language.High:
				return 1
			case language.Low:
				return 0.5
			}
		}
	}
	return 0
}

func faceHKPriority(locale language.Tag) float64 {
	if script, conf := locale.Script(); conf > language.No {
		switch script {
		case language.MustParseScript("Hans"):
			return 0.5
		case language.MustParseScript("Hant"):
			if region, conf := locale.Region(); conf > language.No {
				if preferHK(region) {
					return 1
				}
			}
			return 0.5
		}
	}
	return 0
}

func faceJPPriority(locale language.Tag) float64 {
	if script, conf := locale.Script(); script == language.MustParseScript("Jpan") ||
		script == language.MustParseScript("Hira") ||
		script == language.MustParseScript("Kana") ||
		script == language.MustParseScript("Hrkt") {
		switch conf {
		case language.Exact, language.High:
			return 1
		case language.Low:
			return 0.5
		}
	}
	return 0
}

func faceKRPriority(locale language.Tag) float64 {
	if script, conf := locale.Script(); script == language.MustParseScript("Hang") || script == language.MustParseScript("Kore") {
		switch conf {
		case language.Exact, language.High:
			return 1
		case language.Low:
			return 0.5
		}
	}
	return 0
}
