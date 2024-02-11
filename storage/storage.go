package storage

import (
	"errors"
	"io"
)

var ErrInternal = errors.New("internal failure")

type Storage interface {
	io.Closer
}
