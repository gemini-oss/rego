// pkg/common/cache/cache.go
package cache

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gemini-oss/rego/pkg/common/crypt"
)

var (
	ErrInvalidKeySize = errors.New("invalid encryption key size")
)

type Cache struct {
	accessList      []string             // Least Recently Used (LRU) list
	accessMap       map[string]int       // Least Recently Used (LRU) map
	data            map[string]CacheItem // Cache data
	encryptionKey   []byte               // Encryption key
	hashes          map[string]string    // SHA-256 hashes of the data
	inMemory        bool                 // Defines if the cache is memory-based
	maxItems        int                  // Maximum number of items in the cache
	mutex           sync.RWMutex         // Mutex for thread safety
	persistencePath string               // Path to the file for disk-based cache
	Enabled         bool                 // Defines if the cache is enabled
}

type CacheItem struct {
	Data    string
	Expires time.Time
}

// CacheOptions defines options for creating a new cache
type CacheOptions struct {
	EncryptionKey   []byte
	PersistencePath string
	InMemory        bool
	MaxItems        int
}

func NewCache(args ...interface{}) (*Cache, error) {
	gob.Register([]byte{})

	// Default options
	opts := CacheOptions{
		InMemory: false,
		MaxItems: 1000,
	}

	for _, arg := range args {
		switch v := arg.(type) {
		case []byte:
			opts.EncryptionKey = v
		case string:
			opts.PersistencePath = filepath.Join(os.TempDir(), v)
		case bool:
			opts.InMemory = v
		case int:
			opts.MaxItems = v
		default:
			// Handle unknown option
			if v != nil {
				return nil, errors.New("unknown cache option provided")
			}
		}
	}

	// Validate the encryption key
	err := crypt.ValidPassphrase(opts.EncryptionKey)
	if err != nil {
		return nil, err
	}

	// Initialize Cache with options
	c := &Cache{
		data:            make(map[string]CacheItem),
		hashes:          make(map[string]string),
		accessList:      make([]string, 0, opts.MaxItems),
		accessMap:       make(map[string]int),
		encryptionKey:   opts.EncryptionKey,
		persistencePath: opts.PersistencePath,
		inMemory:        opts.InMemory,
		maxItems:        opts.MaxItems,
	}

	if !opts.InMemory && opts.PersistencePath != "" {
		if err := c.loadFromDisk(); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// Encrypts data using the AES-GCM (256) algorithm
func (c *Cache) encrypt(data []byte) (string, error) {
	return crypt.EncryptAES(data, c.encryptionKey)
}

// Decrypts data using the AES-GCM (256) algorithm
func (c *Cache) decrypt(data string) ([]byte, error) {
	return crypt.DecryptAES(data, c.encryptionKey)
}

func (c *Cache) Set(key string, value interface{}, duration time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	serializedValue, err := c.serializeWithGob(value)
	if err != nil {
		return err
	}

	encryptedValue, err := c.encrypt(serializedValue)
	if err != nil {
		return err
	}

	hash := sha256Hash(serializedValue)
	if existingKey, exists := c.hashes[hash]; exists {
		existingItem := c.data[existingKey]
		existingItem.Expires = time.Now().Add(duration)
		c.data[existingKey] = existingItem
		c.updateAccess(existingKey)
		return nil
	}

	c.data[key] = CacheItem{
		Data:    encryptedValue,
		Expires: time.Now().Add(duration),
	}
	c.hashes[hash] = key
	c.updateAccess(key)

	if !c.inMemory {
		return c.persistToDisk()
	}

	return nil
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	d, exists := c.data[key]
	if !exists || time.Now().After(d.Expires) {
		return nil, false
	}

	// Update the expiration time upon access
	newExpiration := time.Now().Add(1 * time.Minute) // or some other default duration
	d.Expires = newExpiration
	c.data[key] = d

	c.updateAccess(key)

	decryptedValue, err := c.decrypt(d.Data)
	if err != nil {
		return nil, false
	}

	var result []byte
	if err := c.deserializeWithGob(decryptedValue, &result); err != nil {
		return nil, false
	}

	return result, true
}

func (c *Cache) serializeWithGob(data interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	gz := gzip.NewWriter(&buffer)
	enc := gob.NewEncoder(gz)

	if err := enc.Encode(data); err != nil {
		gz.Close()
		return nil, err
	}

	// It's important to close the gzip.Writer to flush the data to the buffer
	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (c *Cache) deserializeWithGob(data []byte, result interface{}) error {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer gz.Close()
	dec := gob.NewDecoder(gz)
	return dec.Decode(result)
}

func sha256Hash(data []byte) string {
	hash := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(hash[:])
}

func (c *Cache) persistToDisk() error {
	if c.inMemory {
		return nil // No action needed for in-memory cache
	}

	fileData, err := c.serializeWithGob(c.data)
	if err != nil {
		return err
	}

	return os.WriteFile(c.persistencePath, fileData, 0600)
}

func (c *Cache) loadFromDisk() error {
	if c.inMemory {
		return nil // No action needed for in-memory cache
	}

	fileData, err := os.ReadFile(c.persistencePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No file to load, not an error
		}
		return err
	}

	return c.deserializeWithGob(fileData, &c.data)
}

// updateAccess updates the access order for a given key
func (c *Cache) updateAccess(key string) {
	if idx, found := c.accessMap[key]; found {
		// Remove the item from its current position
		c.accessList = append(c.accessList[:idx], c.accessList[idx+1:]...)

		// Update accessMap for all items that shifted
		for i := idx; i < len(c.accessList); i++ {
			c.accessMap[c.accessList[i]] = i
		}
	}

	// Add the item to the end of accessList
	c.accessList = append(c.accessList, key)
	c.accessMap[key] = len(c.accessList) - 1

	// Evict the least recently used item if necessary
	if len(c.accessList) > c.maxItems {
		oldest := c.accessList[0]
		c.accessList = c.accessList[1:]
		delete(c.data, oldest)
		delete(c.accessMap, oldest)
	}
}
