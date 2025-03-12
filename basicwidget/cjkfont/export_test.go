// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package cjkfont

import (
	"github.com/xackery/guigui/basicwidget"
)

func FaceSCPriority(hint basicwidget.FaceSourceHint) float64 {
	return faceSCPriority(hint)
}

func FaceTCPriority(hint basicwidget.FaceSourceHint) float64 {
	return faceTCPriority(hint)
}

func FaceHKPriority(hint basicwidget.FaceSourceHint) float64 {
	return faceHKPriority(hint)
}

func FaceJPPriority(hint basicwidget.FaceSourceHint) float64 {
	return faceJPPriority(hint)
}

func FaceKRPriority(hint basicwidget.FaceSourceHint) float64 {
	return faceKRPriority(hint)
}
