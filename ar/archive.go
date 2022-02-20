package ar

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"sync"
	"time"
)

type Arch struct {
	fd      *os.File
	headers []Header
	files   map[string]position
	mux     sync.RWMutex
}

type position struct {
	From int64
	Len  int64
}

func Open(filename string, perm os.FileMode) (*Arch, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_SYNC|os.O_CREATE|os.O_APPEND, perm)
	if err != nil {
		return nil, fmt.Errorf("open archive: %w", err)
	}

	v := &Arch{fd: file, headers: make([]Header, 0), files: make(map[string]position)}

	if err := v.rwSignature(); err != nil {
		return nil, fmt.Errorf("write archive signature: %w", err)
	}

	if err := v.readAllHeaders(); err != nil {
		return nil, fmt.Errorf("read archive: %w", err)
	}

	return v, nil
}

func (v *Arch) Close() error {
	return v.fd.Close()
}

func (v *Arch) List() []Header {
	nh := make([]Header, 0, len(v.headers))
	nh = append(nh, v.headers...)
	return nh
}

func (v *Arch) Read(filename string, w io.Writer) error {
	v.mux.RLock()
	defer v.mux.RUnlock()

	pos, ok := v.files[filename]
	if !ok {
		return fmt.Errorf("%w: %s", ErrFileNotFound, filename)
	}
	if _, err := v.fd.Seek(pos.From, io.SeekStart); err != nil {
		return err
	}

	buf := make([]byte, 256)
	max := pos.Len
	var ii int64

	for {
		if max == 0 {
			return nil
		}

		i, err := v.fd.Read(buf)
		if err != nil {
			return fmt.Errorf("read content: %w", err)
		}

		ii = int64(i)
		if ii > max {
			ii = max
		}

		if _, err = w.Write(buf[:ii]); err != nil {
			return fmt.Errorf("write content: %w", err)
		}
		max -= ii
	}
}

func (v *Arch) Write(filename string, b []byte, perm fs.FileMode) error {
	v.mux.Lock()
	defer v.mux.Unlock()

	if _, ok := v.files[filename]; ok {
		return fmt.Errorf("%w: %s", ErrFileExist, filename)
	}
	cur, err := v.fd.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}

	h := &Header{
		FileName:  filename,
		Timestamp: time.Now().Unix(),
		Mode:      int64(perm),
		Size:      int64(len(b)),
	}
	hb, err := h.Bytes()
	if err != nil {
		return err
	}

	_, _ = buf.Write(hb)
	_, _ = buf.Write(b)
	v.correctSize(h.Size, func() { _, _ = buf.Write(newline) })

	if _, err := v.fd.Write(buf.Bytes()); err != nil {
		return err
	}

	cur += int64(HEAD_SIZE)
	v.files[filename] = position{From: cur, Len: h.Size}
	v.headers = append(v.headers, *h)

	return nil
}

func (v *Arch) Export(filename, dir string) error {
	if err := os.MkdirAll(dir, fs.ModePerm); err != nil {
		return err
	}
	file, err := os.OpenFile(strings.TrimRight(dir, "/")+"/"+filename, os.O_RDWR|os.O_SYNC|os.O_CREATE, fs.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	return v.Read(filename, file)
}

func (v *Arch) Import(filename string) error {
	v.mux.Lock()
	defer v.mux.Unlock()

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close() //nolint: errcheck
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	if _, ok := v.files[stat.Name()]; ok {
		return fmt.Errorf("%w: %s", ErrFileExist, stat.Name())
	}
	cur, err := v.fd.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	h := &Header{
		FileName:  stat.Name(),
		Timestamp: stat.ModTime().Unix(),
		Mode:      int64(stat.Mode()),
		Size:      stat.Size(),
	}
	hb, err := h.Bytes()
	if err != nil {
		return err
	}
	if _, err = v.fd.Write(hb); err != nil {
		return err
	}

	if _, err = io.Copy(v.fd, file); err != nil {
		return err
	}

	v.correctSize(h.Size, func() { _, err = v.fd.Write(hb) })
	if err != nil {
		return err
	}

	cur += int64(HEAD_SIZE)
	v.files[filename] = position{From: cur, Len: h.Size}
	v.headers = append(v.headers, *h)

	return nil
}

func (v *Arch) correctSize(size int64, callFunc func()) {
	if size%2 != 0 {
		callFunc()
	}
}

func (v *Arch) rwSignature() error {
	data := make([]byte, len(signeture))
	i, err := v.fd.Read(data)
	if err != nil && err != io.EOF {
		return err
	}

	if i > 0 && !bytes.Equal(signeture, data) {
		return ErrInvalidFileFormat
	}

	if i == 0 {
		if _, err = v.fd.Write(signeture); err != nil {
			return err
		}
	}

	return nil
}

func (v *Arch) readAllHeaders() error {
	data := make([]byte, HEAD_SIZE)
	for {
		_, err := v.fd.Read(data)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		head := &Header{}
		if err = head.Parse(data); err != nil {
			return err
		}

		if cur, err := v.fd.Seek(0, io.SeekCurrent); err == nil {
			v.files[head.FileName] = position{From: cur, Len: head.Size}
		}

		seek := head.Size
		v.correctSize(seek, func() { seek++ })

		if _, err := v.fd.Seek(seek, io.SeekCurrent); err != nil {
			return err
		}
	}
}
