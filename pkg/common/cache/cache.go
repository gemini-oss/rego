// pkg/common/cache/cache.go
package cache

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"
	"time"
)

var (
	ErrInvalidKeySize = errors.New("invalid encryption key size")
)

type Cache struct {
	mutex           sync.RWMutex
	data            map[string]CacheItem
	encryptionKey   []byte
	persistencePath string
	Use             bool
}

type CacheItem struct {
	Data    string
	Expires time.Time
}

func NewCache(encryptionKey []byte, persistencePath string) (*Cache, error) {
	if len(encryptionKey) != 16 && len(encryptionKey) != 24 && len(encryptionKey) != 32 {
		return nil, ErrInvalidKeySize
	}

	c := &Cache{
		data:            make(map[string]CacheItem),
		encryptionKey:   encryptionKey,
		persistencePath: persistencePath,
	}

	if persistencePath != "" {
		if err := c.loadFromDisk(); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// Encrypts data using the AES algorithm
func (c *Cache) encrypt(data []byte) (string, error) {
	block, err := aes.NewCipher(c.encryptionKey)
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
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// Decrypts data using the AES algorithm
func (c *Cache) decrypt(data string) ([]byte, error) {
	encryptedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(c.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, err
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func (c *Cache) Set(key string, value []byte, duration time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	encryptedValue, err := c.encrypt(value)
	if err != nil {
		return err
	}

	c.data[key] = CacheItem{
		Data:    encryptedValue,
		Expires: time.Now().Add(duration),
	}

	// Optionally write to disk for persistence
	if c.persistencePath != "" {
		return c.persistToDisk()
	}

	c.Use = false
	return nil
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	ev, exists := c.data[key]
	if !exists || time.Now().After(ev.Expires) {
		return nil, false
	}

	decryptedValue, err := c.decrypt(ev.Data)
	if err != nil {
		return nil, false
	}

	return decryptedValue, true
}

// persistToDisk writes the current cache state to the disk
func (c *Cache) persistToDisk() error {
	// Do not use RLock or RUnlock here; assume the caller holds a lock

	fileData, err := json.Marshal(c.data)
	if err != nil {
		return err
	}

	return os.WriteFile(c.persistencePath, fileData, 0644)
}

// loadFromDisk loads cache data from the disk
func (c *Cache) loadFromDisk() error {
	fileData, err := os.ReadFile(c.persistencePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No file to load, not an error
		}
		return err
	}

	return json.Unmarshal(fileData, &c.data)
}
