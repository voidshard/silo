package silo

import (
	"github.com/gtank/cryptopasta"
)

// A silo user is someone (or something) that is allowed to read, write and/or delete
//
type Role struct {
	Id string
	Password []byte

	CanGet bool
	CanPut bool
	CanRm bool
}

// build a user from a name / password.
//  The password is hashed & the user struct(s) are kept in memory.
//
func NewRole(name, password string) (*Role, error) {
	hsh, err := cryptopasta.HashPassword([]byte(password))
	if err != nil {
		return nil, err
	}
	return &Role{
		Id: name,
		Password: hsh,
	}, nil
}

// Check the user's recorded password hash against the given password.
//
func (u *Role) CheckPassword(given string) bool {
	return cryptopasta.CheckPasswordHash(u.Password, []byte(given)) == nil
}
