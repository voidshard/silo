package silo

import (
	"path/filepath"
	"os"
)

// Full silo config
//
type Config struct {
	Misc miscSettings

	Store *storageSettings

	User map[string]*Role
}

// Misc silo settings
//
type miscSettings struct {
	MaxDataBytes int
	MaxKeyBytes int
	EncryptionKey string
}

// Construct a new config with some defaults.
//
func NewConfig() *Config {
	return &Config {
		Misc: miscSettings{
			EncryptionKey: "YouReallyShouldChangeThisToSomethingElse",
			MaxDataBytes: 1000000,
			MaxKeyBytes: 100,
		},
		Store: &storageSettings{
			Driver: "",
			Location: filepath.Join(os.TempDir(), "silo", "store"),
		},
		User: map[string]*Role{
			"read": &Role{
				Id: "read",
				Password: []byte("readpassword"),
				CanGet: true,
			},
			"write": &Role{
				Id: "write",
				Password: []byte("writepassword"),
				CanPut: true,
			},
			"readwrite": &Role{
				Id: "readwrite",
				Password: []byte("readwritepassword"),
				CanGet: true,
				CanPut: true,
			},
			"admin": &Role{
				Id: "admin",
				Password: []byte("adminpassword"),
				CanRm: true,
				CanGet: true,
				CanPut: true,
			},
		},
	}
}
