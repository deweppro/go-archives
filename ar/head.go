package ar

import (
	"reflect"
	"strconv"
	"strings"
)

type kind struct {
	From, Len, Base int
	Prefix          string
	Type            string
}

var (
	fileName = kind{From: 0, Len: 16, Type: "string"}
	modif    = kind{From: 16, Len: 12, Base: 10, Type: "int"}
	ownerID  = kind{From: 28, Len: 6, Base: 10, Type: "int"}
	groupID  = kind{From: 34, Len: 6, Base: 10, Type: "int"}
	fileMode = kind{From: 40, Len: 8, Base: 8, Prefix: "100", Type: "int"}
	fileSize = kind{From: 48, Len: 10, Base: 10, Type: "int"}
	endChar  = kind{From: 58, Len: 2, Type: "nil"}
)

type buffer []byte

func newBuffer(size int) buffer {
	v := make(buffer, size)
	for i := 0; i < size; i++ {
		v[i] = whitespace
	}
	return v
}

func (v buffer) Write(k kind, d interface{}) error {
	var data []byte

	switch vv := d.(type) {
	case []byte:
		data = vv
	case string:
		data = []byte(k.Prefix + vv)
	case int64:
		data = []byte(k.Prefix + strconv.FormatInt(vv, k.Base))
	default:
		return ErrUnsupportedValue
	}

	if len(data) > k.Len {
		return ErrTooLongValue
	}
	copy(v[k.From:k.From+k.Len], data)
	return nil
}

func (v buffer) Read(k kind, d interface{}) error {
	rv := reflect.ValueOf(d)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return ErrInvalidParseValue
	}

	data := strings.TrimRight(string(v[k.From:k.From+k.Len]), " /")

	if len(k.Prefix) > 0 {
		data = strings.TrimPrefix(data, k.Prefix)
	}

	switch k.Type {
	case "string":
		rv.Elem().SetString(data)
		return nil
	case "int":
		i, err := strconv.ParseInt(data, k.Base, 64)
		if err != nil {
			return err
		}
		rv.Elem().SetInt(i)
		return nil
	}

	return ErrInvalidParseValue
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////

//HEAD_SIZE default size for meta data of contained files
const HEAD_SIZE int = 60

//Header meta data of contained files
type Header struct {
	FileName  string
	Timestamp int64
	Mode      int64
	Size      int64
}

//Bytes make string from Header model
func (v *Header) Bytes() ([]byte, error) {
	data := newBuffer(HEAD_SIZE)

	list := map[kind]interface{}{
		fileName: v.FileName,
		modif:    v.Timestamp,
		ownerID:  zero,
		groupID:  zero,
		fileMode: v.Mode,
		fileSize: v.Size,
		endChar:  end,
	}

	for k, val := range list {
		if err := data.Write(k, val); err != nil {
			return nil, err
		}
	}

	return []byte(data), nil
}

//Parse decode string to Header model
func (v *Header) Parse(b []byte) error {
	vv := buffer(b)

	list := []func() error{
		func() error { return vv.Read(fileName, &v.FileName) },
		func() error { return vv.Read(modif, &v.Timestamp) },
		func() error { return vv.Read(fileMode, &v.Mode) },
		func() error { return vv.Read(fileSize, &v.Size) },
	}

	for _, fn := range list {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}
