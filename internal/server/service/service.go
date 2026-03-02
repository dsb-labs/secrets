// Package service provides types that manage interactions between inbound network requests and the persistent
// storage. It effectively encapsulates "business logic".
package service

import (
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

type (
	// The DatabaseManager interface describes types that manage individual user databases.
	DatabaseManager interface {
		// Unlock should open and decrypt the user database associated with the given user identifier.
		Unlock(uuid.UUID, []byte) error
		// Lock should close the user database associated with the given user identifier.
		Lock(uuid.UUID) error
		// Delete should delete the user database associated with the given user identifier.
		Delete(uuid.UUID) error
		// RotateKey should replace the first encryption key with the second encryption key for the
		// given user identifier.
		RotateKey(uuid.UUID, []byte, []byte) error
	}
)

var (
	// ErrReauthenticate is the error given when attempting to perform an operation on a user database that has
	// expired. The user will need to reauthenticate to unlock it again and continue.
	ErrReauthenticate = errors.New("reauthenticate")
)

func deriveKey(password string, salt []byte) []byte {
	return argon2.IDKey(
		[]byte(password),
		salt,
		3,       // iterations
		64*1024, // memory (64 MB)
		4,       // parallelism
		32,      // key length (32 bytes = AES-256)
	)
}
