package ar_test

import (
	"os"
	"testing"

	"github.com/deweppro/go-archives/ar"
	"github.com/stretchr/testify/require"
)

func TestUnit_Open(t *testing.T) {
	file, err := ar.Open("/tmp/zoom_amd64.deb", os.ModePerm)
	require.NoError(t, err)
	require.NotNil(t, file)
}
