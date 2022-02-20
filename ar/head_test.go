package ar_test

import (
	"testing"

	"github.com/deweppro/go-archives/ar"
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
