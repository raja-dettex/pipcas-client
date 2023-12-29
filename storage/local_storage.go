package storage

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type LocalFileStorage struct {
	mu        sync.RWMutex
	FileNames []FileMetaData
	TTL       int
}

type FileMetaData struct {
	FName     string
	TimeStamp time.Time
}

func NewLocalFileStorage(ttl int) *LocalFileStorage {
	return &LocalFileStorage{
		FileNames: make([]FileMetaData, 0),
		TTL:       ttl,
	}
}

func (storage *LocalFileStorage) Evict() {
	if len(storage.FileNames) == 0 {
		go storage.Evict()
		return
	}
	ticker := time.NewTicker(time.Second * time.Duration(storage.TTL))
	defer ticker.Stop()
	<-ticker.C
	if err := os.Remove(storage.FileNames[0].FName); err != nil {
		fmt.Println(err)
	}
	storage.FileNames = storage.FileNames[1:]
	go storage.Evict()
}

func (storage *LocalFileStorage) Write(fName string, r io.Reader) (*os.File, error) {
	f, err := os.Create(fName)
	if err != nil {
		return nil, err
	}
	storage.mu.Lock()
	storage.FileNames = append(storage.FileNames, FileMetaData{FName: fName, TimeStamp: time.Now()})
	storage.mu.Unlock()
	_, err = io.Copy(f, r)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (storage *LocalFileStorage) Read(fName string) (*os.File, error) {
	file, err := os.Open(fName)
	if err != nil {
		return nil, err
	}
	return file, err
}
