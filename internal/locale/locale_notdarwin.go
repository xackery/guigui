// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

//go:build !darwin

package locale

import "github.com/jeandeaual/go-locale"

func Locales() ([]string, error) {
	return locale.GetLocales()
}
