package ar

import "github.com/pkg/errors"

var (
	signeture  = []byte("!<arch>\n")
	newline    = []byte("\n")
	zero       = []byte("0")
	whitespace = []byte(" ")[0]
	end        = []byte("`\n")
)

var (
	ErrTooLongValue      = errors.New("too long value")
	ErrUnsupportedValue  = errors.New("unsupported value")
	ErrInvalidParseValue = errors.New("parsing error")
	ErrInvalidFileFormat = errors.New("invalid file format")
	ErrFileNotFound      = errors.New("file not found")
	ErrFileExist         = errors.New("file already exist")
)
