// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package cjkfont

import "golang.org/x/text/language"

func FaceSCPriority(locale language.Tag) float64 {
	return faceSCPriority(locale)
}

func FaceTCPriority(locale language.Tag) float64 {
	return faceTCPriority(locale)
}

func FaceHKPriority(locale language.Tag) float64 {
	return faceHKPriority(locale)
}

func FaceJPPriority(locale language.Tag) float64 {
	return faceJPPriority(locale)
}

func FaceKRPriority(locale language.Tag) float64 {
	return faceKRPriority(locale)
}
