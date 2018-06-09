/*
The Silo struct exposes data read/write/delete/exists functionality, and handles deciding checking

*/

package silo

import (
	"fmt"
	"github.com/gtank/cryptopasta"
)

const (
	ForbiddenPrefix = "denied"
)

type Silo struct {
	conf *Config
	store Storage
	key *[32]byte
}

// Build a new Silo instance from a config
//
func NewSilo(config *Config) (*Silo, error) {
	key, err := toKey(config.Misc.EncryptionKey)
	if err != nil {
		return nil, err
	}

	// In future we'll switch here based on the storage driver, but for now there is only one
	sConn, err := newFilesystemStorge(config.Store)
	if err != nil {
		return nil, err
	}

	return &Silo{
		conf: config,
		store: sConn,
		key: key,
	}, nil
}

// Turn the given string into a key we can use to encrypt with.
//
func toKey(in string) (*[32]byte, error) {
	c := &[32]byte{}
	bin := []byte(cryptopasta.Hash("", []byte(in)))

	if len(bin) < 32 {
		return c, fmt.Errorf("encryption key must be at least 32 bytes long")
	}

	for i := range c {
		c[i] = bin[i]
	}
	return c, nil
}

// Fetch the given user, assuming the user is found and the passwords match.
//
func (s *Silo) User(username, password string) (*Role, error) {
	u, ok := s.conf.User[username]
	if !ok {
		return nil, nil
	}

	if !u.CheckPassword(password) {
		return nil, fmt.Errorf("password mismatch")
	}

	return u, nil
}

// Store some data in the storage, using the given key as a unique reference.
//
func (s *Silo) Store(user *Role, key string, data []byte) error {
	if !user.CanPut {
		return fmt.Errorf("%s: user %s is not permitted to write", ForbiddenPrefix, user.Id)
	}

	if len(data) > s.conf.Misc.MaxDataBytes {
		return fmt.Errorf("%s: maxdatabytes is currently %d", ForbiddenPrefix, s.conf.Misc.MaxDataBytes)
	}
	if len([]byte(key)) > s.conf.Misc.MaxKeyBytes {
		return fmt.Errorf("%s: maxkeybytes is currently %d", ForbiddenPrefix, s.conf.Misc.MaxKeyBytes)
	}

	exists, err := s.Exists(key)
	if err != nil {
		return err
	}

	if exists && !user.CanRm {
		return fmt.Errorf("%s: file exists and user %s is not permitted to remove", ForbiddenPrefix, user.Id)
	}

	// We encrypt data give to us with our own key. Note it could well be encrypted already, this doesn't actually
	// matter to us.
	cyphertext, err := cryptopasta.Encrypt(data, s.key)
	if err != nil {
		return err
	}
	return s.store.Put(key, cyphertext)
}

// Remove some item by it's key
//
func (s *Silo) Remove(user *Role, key string) error {
	if !user.CanRm {
		return fmt.Errorf("%s: user %s is not permitted to delete", ForbiddenPrefix, user.Id)
	}
	if len([]byte(key)) > s.conf.Misc.MaxKeyBytes {
		return fmt.Errorf("%s: maxkeybytes is currently %d", ForbiddenPrefix, s.conf.Misc.MaxKeyBytes)
	}

	return s.store.Delete(key)
}

// Get the stored item given it's unique key
//
func (s *Silo) Get(user *Role, key string) ([]byte, error) {
	if !user.CanGet {
		return nil, fmt.Errorf("%s: user %s is not permitted to read", ForbiddenPrefix, user.Id)
	}
	if len([]byte(key)) > s.conf.Misc.MaxKeyBytes {
		return nil, fmt.Errorf("%s: maxkeybytes is currently %d", ForbiddenPrefix, s.conf.Misc.MaxKeyBytes)
	}

	cyphertext, err := s.store.Get(key)
	if err != nil {
		return nil, err
	}
	return cryptopasta.Decrypt(cyphertext, s.key)
}

// Return if something with the given key has been stored here already
//
func (s *Silo) Exists(key string) (bool, error) {
	return s.store.Exists(key)
}
