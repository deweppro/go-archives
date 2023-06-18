/*
 *  Copyright (c) 2021-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package ar_test

import (
	"testing"

	"github.com/osspkg/go-archives/ar"
	"github.com/stretchr/testify/require"
)

func TestUnit_NewBuffer(t *testing.T) {
	demo := []byte("hello.go        123456      0     0     100777  999       `\n")
	h := &ar.Header{
		FileName:  "hello.go",
		Timestamp: 123456,
		Size:      999,
		Mode:      0777,
	}
	b, err := h.Bytes()
	require.NoError(t, err)
	require.Equal(t, demo, b)

	h2 := &ar.Header{}
	require.NoError(t, h2.Parse(demo))

	require.Equal(t, h, h2)
}
