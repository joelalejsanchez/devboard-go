// Package auth contains auth and crypto functions
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// ErrInvalidPasswordHash indica que el hash almacenado tiene el formato incorrecto
var ErrInvalidPasswordHash = errors.New("formato de hash invalido")

// ErrPasswordMissmatch indica que el password no coincide con el hash
var ErrPasswordMissmatch = errors.New("password incorrecto")

type argon2Params struct {
	memory      uint32 // en KBs - 64MBs recomendados
	iterations  uint32 //
	parallelism uint8
	saltLength  uint32 // bytes
	keyLength   uint32 // bytes
}

var defaultArgon2Params = argon2Params{
	memory:      64 * 1024,
	iterations:  3,
	parallelism: 2,
	saltLength:  16, // bytes
	keyLength:   32, // bytes
}

// HashPassword returns a secure hash
func HashPassword(password string) (string, error) {
	salt := make([]byte, defaultArgon2Params.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generando salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		defaultArgon2Params.iterations,
		defaultArgon2Params.memory,
		defaultArgon2Params.parallelism,
		defaultArgon2Params.keyLength,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		defaultArgon2Params.memory,
		defaultArgon2Params.iterations,
		defaultArgon2Params.parallelism,
		b64Salt,
		b64Hash,
	)

	return encoded, nil
}

// VerifyPassword compares password and returns errors if not matches
func VerifyPassword(password, encodedHash string) error {
	params, salt, hash, err := parseHash(encodedHash)
	if err != nil {
		return fmt.Errorf("parseando hash: %w", err)
	}

	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.iterations,
		params.memory,
		params.parallelism,
		params.keyLength,
	)

	if subtle.ConstantTimeCompare(hash, otherHash) != 1 {
		return ErrPasswordMissmatch
	}
	return nil
}

// parseHash extracts parms and values from hash
func parseHash(encodedHash string) (*argon2Params, []byte, []byte, error) {
	// $argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, ErrInvalidPasswordHash
	}
	if parts[1] != "argon2id" {
		return nil, nil, nil, ErrInvalidPasswordHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, nil, nil, ErrInvalidPasswordHash

	}
	if version != argon2.Version {
		return nil, nil, nil, ErrInvalidPasswordHash
	}
	params := &argon2Params{}
	if _, err := fmt.Sscanf(
		parts[3],
		"m=%d,t%d,p=%d",
		&params.memory,
		&params.iterations,
		&params.parallelism,
	); err != nil {
		return nil, nil, nil, ErrInvalidPasswordHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, ErrInvalidPasswordHash
	}
	params.saltLength = uint32(len(salt))

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, ErrInvalidPasswordHash
	}
	params.keyLength = uint32(len(hash))

	return params, salt, hash, nil
}
