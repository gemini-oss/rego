// pkg/internal/tests/common/crypt/crypt_test.go
package crypt_test

import (
	"testing"

	"github.com/gemini-oss/rego/pkg/common/crypt"
)

func TestValidPassphrase(t *testing.T) {
	// Test a passphrase with repeating patterns
	weakPassphrase := []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	err := crypt.ValidPassphrase(weakPassphrase)
	if err == nil {
		t.Errorf("ValidPassphrase() should return an error for weak key with repeating pattern")
	}

	// Test a passphrase without special characters
	weakPassphrase = []byte("Abcdefghijklmnopqrstuvwxyz01234")
	err = crypt.ValidPassphrase(weakPassphrase)
	if err == nil {
		t.Errorf("ValidPassphrase() should return an error for key without special characters")
	}

	// Test a strong passphrase
	strongPassphrase := []byte("8jCcfHzjg*8mXD8qWjj9mk*QNZnVsMRt")
	err = crypt.ValidPassphrase(strongPassphrase)
	if err != nil {
		t.Errorf("ValidPassphrase() should return nil for strong key, got error: %v", err)
	}

	// Test with a key of incorrect length
	shortPassphrase := []byte("ShortKey")
	err = crypt.ValidPassphrase(shortPassphrase)
	if err == nil {
		t.Errorf("ValidPassphrase() should return an error for short key")
	}
}
