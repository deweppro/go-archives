package ar

import (
	"bytes"
	"fmt"
	"os"

	"github.com/pkg/errors"
)

var (
	errFileWithDifferentFormat = errors.New("existing file has a different format")
)

type Arch struct {
	file    *os.File
	headers []Header
}

func Open(filename string, perm os.FileMode) (*Arch, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_SYNC|os.O_CREATE|os.O_APPEND, perm)
	if err != nil {
		return nil, err
	}

	v := &Arch{file: file, headers: make([]Header, 0)}

	if err := v.writeSignature(); err != nil {
		return nil, err
	}

	if err := v.readAllFiles(); err != nil {
		return nil, err
	}

	return v, nil
}

func (v *Arch) Close() error {
	return v.file.Close()
}

func (v *Arch) List() {

}

func (v *Arch) ReadFile() {

}

func (v *Arch) WriteFile() {

}

func (v *Arch) writeSignature() error {
	data := make([]byte, len(signeture))
	i, err := v.file.Read(data)
	if err != nil {
		return err
	}

	if i == 0 {
		_, err = v.file.Write(signeture)
		if err != nil {
			return err
		}
	}

	if !bytes.Equal(signeture, data) {
		return errFileWithDifferentFormat
	}

	return nil
}

func (v *Arch) readAllFiles() error {
	position := len(signeture)
	data := make([]byte, HEAD_SIZE)
	for {
		_, err := v.file.Read(data)
		if err != nil {
			return err
		}

		head := &Header{}
		if err = head.Parse(data); err != nil {
			return err
		}

		fmt.Println(head.FileName)

		seek := head.Size
		if seek%2 > 0 {
			seek++
		}

		if _, err := v.file.Seek(seek, 1); err != nil {
			return err
		}
	}
}
