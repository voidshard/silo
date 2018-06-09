package silo

import (
	"os"
	"path/filepath"
	"encoding/base64"
	"io/ioutil"
	"fmt"
)


// interface for some storage backend
//
type Storage interface {
	Put(string, []byte) error
	Exists(string) (bool, error)
	Get(string) ([]byte, error)
	Delete(string) error
}

// the most trivial kind of storage implementation
//
type filesystem struct {
	root string
}

type storageSettings struct {
	Driver string
	Location string
}

func newFilesystemStorge(settings *storageSettings) (Storage, error) {
	if settings.Location == "" {
		return nil, fmt.Errorf("filesystem storage requires setting Location")
	}
	return &filesystem{root: settings.Location}, os.MkdirAll(settings.Location, os.ModePerm)
}

// Return if the given key has been stored here.
//
func (f *filesystem) Exists(key string) (bool, error) {
	_, err := os.Stat(f.storagePath(key))

	if os.IsNotExist(err)  {
		return false, nil
	}

	return true, err
}

// internal func to encode the key into a non-filesystem interacting path.
//
func (f *filesystem) storagePath(key string) string {
	// Url encoding doesn't have '/' symbols, which are a bit awkward for us in a filesystem.
	// We encode the string to circumvent and weird chars supplied to us, limiting the available
	// chars to a-z, A-Z, 0-9, '-' and '_'.
	return filepath.Join(f.root, base64.RawURLEncoding.EncodeToString([]byte(key)))
}

// Write the given data to disk, using the given key
//
func (f *filesystem) Put(key string, data []byte) error {
	return ioutil.WriteFile(f.storagePath(key), data, 0644)
}

// Fetch the data indicated by the given key from disk
//
func (f *filesystem) Get(key string) ([]byte, error) {
	return ioutil.ReadFile(f.storagePath(key))
}

// Remove the data indicated by the given key from disk
//
func (f *filesystem) Delete(key string) error {
	return os.Remove(f.storagePath(key))
}
