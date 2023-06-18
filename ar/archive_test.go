/*
 *  Copyright (c) 2021-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package ar_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"testing"

	"github.com/osspkg/go-archives/ar"
	"github.com/stretchr/testify/require"
)

const demoDeb = "ITxhcmNoPgpkZWJpYW4tYmluYXJ5ICAgMTY0NTMxODYwMSAgMCAgICAgMCAgICAgMTAwNjQ0ICA0ICAgICAgICAgYAoyLjAKY29" +
	"udHJvbC50YXIuZ3ogIDE2NDUzMTg2MDEgIDAgICAgIDAgICAgIDEwMDY0NCAgNzEzICAgICAgIGAKH4sIAAAAAAAA/+yYz27bNhjAfeZTcNl1lkj" +
	"9tbVh2IActsOAAMF2p6hPMTH9A0m3TU9tX6DHHvsK7p8AQYo4r0C/USHHbWOlrYG2kpGWvwsFEtRH8cOPH23HHfUOIYTEYbhuCSHddv1MA48EPi" +
	"UxpSNCYj/2Rjjsf2mj0VxpJkeEfO17uh93R3DcMgvVvFQ9xmj3IwqCT+afRuF2/tsmGOFBNvEHzz+bkDCfAqd0SqaT6SSjNMtjEnACE+6RDGPQ3E" +
	"1Tl9dVLk6cU1YWiPIQAhb4EzbxacrCOPNozOOcEvAhejdJnSoNZbZp3TR1FMh7ggPa90db3uO0idWyLnqMscv/kES3/I8j6/8QHDH+PzuBBKcpOq" +
	"7nkl8//gdSibpKMHGIQ9Gfks+EBq7nEhLMyiwK0D9MVJqJCmSC/1Ug8W9zBfIPeMDKpgCH1+Xv6O9KaVYUkI2PxUNIMEWH0ECVqQRvTodfMGdjDlK" +
	"LXHCmQaFj4Hod+j6k6EiKWgp9muC6aXtZgf6qS2jWS55p3ajEdW+EdNEhKC5Fc/0K88wszCuzNG/M0rw0S3OBzdJcmfPVY7Mwl+Z89RSbK7MwF+Zs" +
	"9cQsMMIOwua5WZrXq0dmaV6YS7M0Z51ZbcfWrH1n8ctZ+5/nooD+bgC7/Cde0PGfehG1/g+B+5Hyvu81WYbDcRsJolK6xxi7/PdCr+t/O2z9H4Cff" +
	"3JTUbkpUzOERI41KI3HOT5wP3uFP/gV6xlUCGO8qeRcF1jpumlvD9u9mVAsLaAd6I4wKOtqLKGoWdYZk6BAj3MmCshQLpA9lXrBcZta6X4PgJ31P/K" +
	"7/tM4tv4PwZb/N01mUrfGfuiCamPxvpds+Yas678se42xu/6T7u//wLf+D8Jdqf/73qfvlev63+8BsLP+01v//0d+aP0fgi3/970Yi8VisQzG2wAAA" +
	"P//+kw+rgAiAAAKZGF0YS50YXIuZ3ogICAgIDE2NDUzMTg2MDEgIDAgICAgIDAgICAgIDEwMDY0NCAgNDczICAgICAgIGAKH4sIAAAAAAAA/+yWQW" +
	"vbMBSAfdav0B+o/RTH8TD4sEEpZfTStOwQcpCtF1dMkYL0nC7/fiRmgYVAyRZ76+YvBwlJRNb7nv0UJ1HvAADkWXZoAeC0PfTFdALTVEAuRASQp/" +
	"kk4ln/jxZFbSDpI4Df/Z/Tw70T4gSp7jkHLvYvQEyz0f8QdP6rqs8U+AX/WQqj/yE4+q+dXekm3sm1ufYeb/kX+fToP8sgAjGZzfKIDxLE/9w/2m3" +
	"BFW4ZM64peKJwmwRSriVmcIum4CljL0SbgnEulfIFh/jwKwTAB2AKq7Y5MznZT7K1Vsrgq/S4X0Iv3hEZbZuCp+8nRv8y3fsfdoFwrXoqApd//8U" +
	"sHev/IPzsv2uvnQaX+0/FTIz+h+Cs/6qKA/qtrvEqe7zlP83ESf3PRCrG+j8Ei2erack+rgh9aZFenf8ak/QNEmOLeZcFS/Yc0JfeOWJ33rWbrvu" +
	"I+9hR6ezNSmrTevwxNMe6TCGwp90Gy6DXG4Ps9hvW88P6pA0+qbRNqor71vKbm+7yWZ65i7LP2pgHp7DceFdjCIeBuW6sNOX8/u7p9vGBscW9DSSN" +
	"WbIv0hKqT7tS4Uq2ho5n+dOBHhkZGfnL+B4AAP//x5J3eQAWAAAK"

func setUp(filename string, data string) error {
	bin, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return fmt.Errorf("base64 decode: %w", err)
	}

	return os.WriteFile(filename, bin, fs.ModePerm)
}

func TestUnit_ArchiveRead(t *testing.T) {
	require.NoError(t, setUp("/tmp/demo.deb", demoDeb))
	os.Remove("/tmp/123/control.tar.gz")

	fd0, err := ar.Open("/tmp/demo.deb", os.ModePerm)
	require.NoError(t, err)
	require.NotNil(t, fd0)
	defer fd0.Close()

	buf := &bytes.Buffer{}

	require.NoError(t, fd0.Read("debian-binary", buf))
	require.Equal(t, "2.0\n", buf.String())

	require.NoError(t, fd0.Export("control.tar.gz", "/tmp/123/"))

	fd1, err := os.Open("/tmp/123/control.tar.gz")
	require.NoError(t, err)
	require.NotNil(t, fd1)
	defer fd1.Close()

	fd2, err := gzip.NewReader(fd1)
	require.NoError(t, err)
	require.NotNil(t, fd2)
	defer fd2.Close()

	fd3 := tar.NewReader(fd2)
	require.NotNil(t, fd3)

	list := func() []string {
		l := make([]string, 0)
		for {
			hdr, err := fd3.Next()
			if errors.Is(err, io.EOF) {
				return l
			}
			if err != nil {
				t.Fatal(err)
			}
			l = append(l, hdr.Name)
		}
	}()

	require.Equal(t, []string{"./", "./md5sums", "./control",
		"./conffiles", "./preinst", "./postinst", "./prerm", "./postrm"}, list)
}

func TestUnit_ArchiveCreate(t *testing.T) {
	os.Remove("/tmp/demo1.ar")
	fd0, err := ar.Open("/tmp/demo1.ar", os.ModePerm)
	require.NoError(t, err)
	require.NotNil(t, fd0)

	require.NoError(t, fd0.Write("file1", []byte("file1 text"), os.ModePerm))
	require.NoError(t, fd0.Write("file2", []byte("file2 text"), os.ModePerm))
	require.Error(t, fd0.Write("file2", []byte("file2 text!"), os.ModePerm))
	require.NoError(t, os.WriteFile("/tmp/ddddd.txt", []byte("ddddd file"), fs.ModePerm))
	require.NoError(t, fd0.Import("/tmp/ddddd.txt", 0))
	fd0.Close()

	fd0, err = ar.Open("/tmp/demo1.ar", os.ModePerm)
	require.NoError(t, err)
	require.NotNil(t, fd0)
	defer fd0.Close()

	buf := &bytes.Buffer{}

	require.NoError(t, fd0.Read("file1", buf))
	require.Equal(t, "file1 text", buf.String())

	buf.Reset()
	require.NoError(t, fd0.Read("file2", buf))
	require.Equal(t, "file2 text", buf.String())

	buf.Reset()
	require.NoError(t, fd0.Read("ddddd.txt", buf))
	require.Equal(t, "ddddd file", buf.String())

}
