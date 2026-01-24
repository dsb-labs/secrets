// Package service provides types that manage interactions between inbound network requests and the persistent
// storage. It effectively encapsulates "business logic".
package service

import (
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

type (
	DatabaseManager interface {
		Unlock(uuid.UUID, []byte) error
		Lock(uuid.UUID) error
	}
)

var (
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
