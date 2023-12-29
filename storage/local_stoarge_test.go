package storage

import (
	"bytes"
	"fmt"
	"testing"
)

func TestLocalStorage(t *testing.T) {
	storage := NewLocalFileStorage(10)
	go storage.Evict()
	for i := 0; i < 3; i++ {
		if _, err := storage.Write(fmt.Sprintf("hello%v.txt", i+1), bytes.NewBuffer([]byte(fmt.Sprintf("hello %v", i+1)))); err != nil {
			fmt.Println(err)
		}
	}

}
