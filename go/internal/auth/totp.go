package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1" //nolint:gosec // TOTP standard requires SHA-1
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	totpPeriod    = 30
	totpDigits    = 6
	totpWindow    = 1
	totpSecretLen = 20
	totpIssuer    = "BesOps"
)

// GenerateTOTPSecret creates a new random base32-encoded secret suitable for TOTP.
func GenerateTOTPSecret() (string, error) {
	buf := make([]byte, totpSecretLen)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generating TOTP secret: %w", err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf), nil
}

// TOTPKeyURI builds an otpauth:// URI for QR code display.
func TOTPKeyURI(secret, username string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=%d&period=%d",
		totpIssuer, username, secret, totpIssuer, totpDigits, totpPeriod)
}

// VerifyTOTP checks a 6-digit TOTP token against a base32-encoded secret.
// Allows a window of ±1 period to account for clock drift.
func VerifyTOTP(secret, token string) bool {
	now := time.Now().Unix()
	for i := -totpWindow; i <= totpWindow; i++ {
		counter := (now / totpPeriod) + int64(i)
		expected := generateTOTP(secret, counter)
		if expected == token {
			return true
		}
	}
	return false
}

func generateTOTP(secret string, counter int64) string {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
	if err != nil {
		return ""
	}

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(counter)) //nolint:gosec // counter is always positive (unix time / 30)

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	code := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff

	otp := code % uint32(math.Pow10(totpDigits))
	return fmt.Sprintf("%06d", otp)
}
