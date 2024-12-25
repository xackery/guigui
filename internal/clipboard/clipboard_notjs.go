// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

//go:build !js

package clipboard

import "github.com/atotto/clipboard"

func ReadAll() (string, error) {
	return clipboard.ReadAll()
}

func WriteAll(text string) error {
	return clipboard.WriteAll(text)
}
