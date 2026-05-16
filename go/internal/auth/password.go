package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

const (
	argon2Time    = 1
	argon2Memory  = 64 * 1024
	argon2Threads = 4
	argon2KeyLen  = 32
	argon2SaltLen = 16
)

// HashPassword hashes a password using Argon2id.
func HashPassword(password string) (string, error) {
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argon2Memory,
		argon2Time,
		argon2Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

// VerifyPassword checks a password against a stored hash.
// Supports both Argon2id (new) and bcrypt (legacy migration).
func VerifyPassword(password, encoded string) bool {
	if strings.HasPrefix(encoded, "$2a$") || strings.HasPrefix(encoded, "$2b$") {
		return bcrypt.CompareHashAndPassword([]byte(encoded), []byte(password)) == nil
	}
	return verifyArgon2id(password, encoded)
}

// NeedsRehash returns true if the hash is in a legacy format (bcrypt) and
// should be upgraded to Argon2id on next login.
func NeedsRehash(encoded string) bool {
	return strings.HasPrefix(encoded, "$2a$") || strings.HasPrefix(encoded, "$2b$")
}

func verifyArgon2id(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}

	var memory, iterations uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &threads); err != nil {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	keyLen := uint32(len(expectedHash)) //nolint:gosec // length is bounded by hash output
	hash := argon2.IDKey([]byte(password), salt, iterations, memory, threads, keyLen)
	return subtle.ConstantTimeCompare(hash, expectedHash) == 1
}
