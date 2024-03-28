// pkg/common/crypt/crypt.go
package crypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"unicode"

	"golang.org/x/crypto/argon2"
)

const (
	minPassphraseLength = 32
	maxPassphraseLength = 128
	saltSize            = 16
)

type PassphraseError struct {
	Issues []string
}

func (e *PassphraseError) Error() string {
	return "Passphrase validation failed: " + strings.Join(e.Issues, "; ")
}

func generateSalt() ([]byte, error) {
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// deriveKey uses Argon2 to derive a key from a passphrase and a salt
func deriveKey(passphrase, salt []byte) []byte {
	return argon2.IDKey(passphrase, salt, 1, 64*1024, 4, 32)
}

// EncryptAES encrypts data using AES-GCM with a passphrase
func EncryptAES(data []byte, passphrase []byte) (string, error) {
	salt, err := generateSalt()
	if err != nil {
		return "", err
	}
	key := deriveKey(passphrase, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	encrypted := gcm.Seal(nonce, nonce, data, nil)
	encryptedDataWithSalt := append(salt, encrypted...)
	return base64.StdEncoding.EncodeToString(encryptedDataWithSalt), nil
}

// DecryptAES decrypts data encrypted with AES-GCM using a passphrase
func DecryptAES(data string, passphrase []byte) ([]byte, error) {
	encryptedDataWithSalt, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	if len(encryptedDataWithSalt) < saltSize {
		return nil, errors.New("encrypted data is too short to contain salt")
	}

	salt, encryptedData := encryptedDataWithSalt[:saltSize], encryptedDataWithSalt[saltSize:]
	key := deriveKey(passphrase, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, errors.New("encrypted data is too short")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func ValidPassphrase(passphrase []byte) error {
	if len(passphrase) < minPassphraseLength || len(passphrase) > maxPassphraseLength {
		return &PassphraseError{Issues: []string{fmt.Sprintf("must be between %d and %d bytes long", minPassphraseLength, maxPassphraseLength)}}
	}

	hasUpper, hasSpecial := false, false
	repeatedPatternFound := false
	charFrequency := make(map[byte]int)
	issues := make([]string, 0)

	for i, b := range passphrase {
		charFrequency[b]++

		if unicode.IsUpper(rune(b)) {
			hasUpper = true
		}
		if strings.ContainsRune("!@#$%^&*()-_=+[]{}|;:',.<>/?`~", rune(b)) {
			hasSpecial = true
		}

		if !repeatedPatternFound && i < len(passphrase)-3 {
			for j := 3; j <= 6 && i+j < len(passphrase); j++ {
				if bytes.Contains(passphrase[i+1:], passphrase[i:i+j]) {
					repeatedPatternFound = true
					break
				}
			}
		}
	}

	if !hasSpecial {
		issues = append(issues, "must contain special characters")
	}
	if !hasUpper {
		issues = append(issues, "must contain uppercase letters")
	}
	if repeatedPatternFound {
		issues = append(issues, "contains repeated patterns")
	}
	if entropy(charFrequency, len(passphrase)) < 4 {
		issues = append(issues, "does not meet the entropy requirements")
	}

	if len(issues) > 0 {
		return &PassphraseError{Issues: issues}
	}

	return nil
}

func entropy(charFrequency map[byte]int, length int) float64 {
	entropy := 0.0
	for _, freq := range charFrequency {
		probability := float64(freq) / float64(length)
		entropy -= probability * math.Log2(probability)
	}
	return entropy
}

func SecureRandomInt(max int) (int, error) {
	if max <= 0 {
		return 0, fmt.Errorf("invalid max value")
	}
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return 0, err
	}
	value := binary.LittleEndian.Uint64(b[:])
	return int(value % uint64(max)), nil
}
