package storage

import (
	"io"
	"os"
)

type FileStorage interface {
	Write(string, io.Reader) (*os.File, error)
	Read(string) (*os.File, error)
	Evict()
}
