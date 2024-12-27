// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package cjkfont_test

import (
	"testing"

	"golang.org/x/text/language"

	"github.com/hajimehoshi/guigui/basicwidget/cjkfont"
)

func TestPriorities(t *testing.T) {
	testCases := []struct {
		locale     language.Tag
		prioritySC float64
		priorityTC float64
		priorityHK float64
		priorityJP float64
		priorityKR float64
	}{
		{
			locale:     language.Make("zh-Hans"),
			prioritySC: 1,
			priorityTC: 0.5,
			priorityHK: 0.5,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("zh-Hant"),
			prioritySC: 0.5,
			priorityTC: 1,
			priorityHK: 0.5,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("zh-CN"),
			prioritySC: 1,
			priorityTC: 0.5,
			priorityHK: 0.5,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("zh-Hans-CN"),
			prioritySC: 1,
			priorityTC: 0.5,
			priorityHK: 0.5,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("zh-Hant-CN"),
			prioritySC: 0.5,
			priorityTC: 1,
			priorityHK: 0.5,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("zh-MO"),
			prioritySC: 0.5,
			priorityTC: 1,
			priorityHK: 0.5,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("zh-Hans-MO"),
			prioritySC: 1,
			priorityTC: 0.5,
			priorityHK: 0.5,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("zh-Hant-MO"),
			prioritySC: 0.5,
			priorityTC: 1,
			priorityHK: 0.5,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("zh-TW"),
			prioritySC: 0.5,
			priorityTC: 1,
			priorityHK: 0.5,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("zh-Hans-TW"),
			prioritySC: 1,
			priorityTC: 0.5,
			priorityHK: 0.5,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("zh-Hant-TW"),
			prioritySC: 0.5,
			priorityTC: 1,
			priorityHK: 0.5,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("zh-HK"),
			prioritySC: 0.5,
			priorityTC: 0.5,
			priorityHK: 1,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("zh-Hans-HK"),
			prioritySC: 1,
			priorityTC: 0.5,
			priorityHK: 0.5,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("zh-Hant-HK"),
			prioritySC: 0.5,
			priorityTC: 0.5,
			priorityHK: 1,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("zh-Latn"),
			prioritySC: 0,
			priorityTC: 0,
			priorityHK: 0,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("zh-Latn-CN"),
			prioritySC: 0,
			priorityTC: 0,
			priorityHK: 0,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("zh-Latn-TW"),
			prioritySC: 0,
			priorityTC: 0,
			priorityHK: 0,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("zh-Latn-HK"),
			prioritySC: 0,
			priorityTC: 0,
			priorityHK: 0,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("en-Hans"),
			prioritySC: 1,
			priorityTC: 0.5,
			priorityHK: 0.5,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("en-Hant"),
			prioritySC: 0.5,
			priorityTC: 1,
			priorityHK: 0.5,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("zh"),
			prioritySC: 1,
			priorityTC: 0.5,
			priorityHK: 0.5,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("ja"),
			prioritySC: 0,
			priorityTC: 0,
			priorityHK: 0,
			priorityJP: 1,
			priorityKR: 0,
		},
		{
			locale:     language.Make("ja-Latn"),
			prioritySC: 0,
			priorityTC: 0,
			priorityHK: 0,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("en-Jpan"),
			prioritySC: 0,
			priorityTC: 0,
			priorityHK: 0,
			priorityJP: 1,
			priorityKR: 0,
		},
		{
			locale:     language.Make("ko"),
			prioritySC: 0,
			priorityTC: 0,
			priorityHK: 0,
			priorityJP: 0,
			priorityKR: 1,
		},
		{
			locale:     language.Make("ko-Latn"),
			prioritySC: 0,
			priorityTC: 0,
			priorityHK: 0,
			priorityJP: 0,
			priorityKR: 0,
		},
		{
			locale:     language.Make("en-Hang"),
			prioritySC: 0,
			priorityTC: 0,
			priorityHK: 0,
			priorityJP: 0,
			priorityKR: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.locale.String(), func(t *testing.T) {
			if got, want := cjkfont.FaceTCPriority(tc.locale), tc.priorityTC; got != want {
				t.Errorf("FaceTCPriority(%v) = %f; want %f", tc.locale, got, want)
			}
			if got, want := cjkfont.FaceSCPriority(tc.locale), tc.prioritySC; got != want {
				t.Errorf("FaceSCPriority(%v) = %f; want %f", tc.locale, got, want)
			}
			if got, want := cjkfont.FaceHKPriority(tc.locale), tc.priorityHK; got != want {
				t.Errorf("FaceHKPriority(%v) = %f; want %f", tc.locale, got, want)
			}
			if got, want := cjkfont.FaceJPPriority(tc.locale), tc.priorityJP; got != want {
				t.Errorf("FaceJPPriority(%v) = %f; want %f", tc.locale, got, want)
			}
			if got, want := cjkfont.FaceKRPriority(tc.locale), tc.priorityKR; got != want {
				t.Errorf("FaceKRPriority(%v) = %f; want %f", tc.locale, got, want)
			}
		})
	}
}
