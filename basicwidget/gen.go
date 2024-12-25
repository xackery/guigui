// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

//go:build ignore

package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	if err := xmain(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func xmain() error {
	// Get `NotoSans[wdth,wght].ttf` (Full) from https://notofonts.github.io/#latin-greek-cyrillic.
	const url = "https://cdn.jsdelivr.net/gh/notofonts/notofonts.github.io/fonts/NotoSans/full/variable-ttf/NotoSans[wdth,wght].ttf"
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	bs, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	out, err := os.Create("NotoSans.ttf.gz")
	if err != nil {
		return err
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	gw, err := gzip.NewWriterLevel(w, gzip.BestCompression)
	if err != nil {
		return err
	}

	if _, err := gw.Write(bs); err != nil {
		return err
	}
	if err := gw.Close(); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}

	return nil
}
