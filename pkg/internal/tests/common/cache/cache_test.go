// pkg/internal/tests/common/cache/cache_test.go
package cache_test

import (
	"bytes"
	"os"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gemini-oss/rego/pkg/common/cache"
)

func TestCacheSetAndGet(t *testing.T) {
	key := "testKey"
	value := []byte("testValue")
	encryptionKey := []byte("a-very-very-very-very-secret-key") // 32 bytes

	c, _ := cache.NewCache(encryptionKey, true)

	// Test setting and getting a value
	err := c.Set(key, value, 1*time.Minute)
	if err != nil {
		t.Errorf("Set() error = %v, wantErr %v", err, nil)
	}

	retrievedValue, exists := c.Get(key)
	if !exists {
		t.Errorf("Get() exist = %v, want %v", exists, true)

	}

	if !bytes.Equal(retrievedValue, value) {
		t.Errorf("Get() value = %v, want %v", retrievedValue, value)
	}
}

func TestCacheExpiration(t *testing.T) {
	key := "expireKey"
	value := []byte("expireValue")
	encryptionKey := []byte("32-byte-long-encryption-key-1234") // 32 bytes

	c, _ := cache.NewCache(encryptionKey, true)

	// Set with a short expiration
	err := c.Set(key, value, 100*time.Millisecond)
	if err != nil {
		t.Errorf("Set() error = %v, wantErr %v", err, nil)
	}

	// Wait for the item to expire
	time.Sleep(150 * time.Millisecond)

	_, exists := c.Get(key)
	if exists {
		t.Error("Get() should not find expired item, but it did")
	}
}

func TestCacheNonExistentKey(t *testing.T) {
	key := "nonExistentKey"
	encryptionKey := []byte("key-for-nonexistent-test") // 32 bytes
	c, _ := cache.NewCache(encryptionKey, true)

	_, exists := c.Get(key)
	if exists {
		t.Error("Get() found a value for a non-existent key")
	}
}

func TestCachePersistence(t *testing.T) {
	key := "persistKey"
	value := []byte("persistValue")
	encryptionKey := []byte("32-byte-long-encryption-key-1234") // 32 bytes
	tempFile := "temp_cache.gob"

	defer os.Remove(tempFile)

	c, _ := cache.NewCache(encryptionKey, tempFile)

	err := c.Set(key, value, 1*time.Minute/2/2)
	if err != nil {
		t.Errorf("Set() error = %v, wantErr %v", err, nil)
	}

	newCache, _ := cache.NewCache(encryptionKey, tempFile)
	retrievedValue, exists := newCache.Get(key)
	if !exists {
		t.Errorf("Get() exist = %v, want %v", exists, true)
	}

	if !bytes.Equal(retrievedValue, value) {
		t.Errorf("Get() value = %v, want %v", retrievedValue, value)
	}
}

func TestCacheConcurrency(t *testing.T) {
	encryptionKey := []byte("32-byte-long-encryption-key-1234")
	c, _ := cache.NewCache(encryptionKey)

	var wg sync.WaitGroup
	numWorkers := 10

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			key := "key" + strconv.Itoa(workerID)
			value := []byte("value" + strconv.Itoa(workerID))
			c.Set(key, value, time.Minute)
			retrievedValue, _ := c.Get(key)
			if !bytes.Equal(retrievedValue, value) {
				t.Errorf("Get() value = %v, want %v", retrievedValue, value)
			}
		}(i)
	}

	wg.Wait()
}

func TestCacheWithLargeData(t *testing.T) {
	encryptionKey := []byte("32-byte-long-encryption-key-1234")
	c, _ := cache.NewCache(encryptionKey, "large_data.gob")

	largeValue := make([]byte, 1024*1024*1024) // 1GB
	key := "largeKey"

	err := c.Set(key, largeValue, time.Minute)
	if err != nil {
		t.Errorf("Error setting large data: %v", err)
	}

	retrievedValue, exists := c.Get(key)
	if !exists || !reflect.DeepEqual(retrievedValue, largeValue) {
		t.Errorf("Get() value = %v, want %v", retrievedValue, largeValue)
	}
}

func TestCacheInvalidKey(t *testing.T) {
	invalidKey := []byte("invalid-encryption-key")
	_, err := cache.NewCache(invalidKey, "")
	if err == nil {
		t.Errorf("Expected an error for invalid key size, but got none")
	}
}

func TestCachePersistenceFailure(t *testing.T) {
	encryptionKey := []byte("32-byte-long-encryption-key-1234")
	nonExistentPath := "/non/existent/path/cache.gob"

	// Expecting no error even though the path doesn't exist
	_, err := cache.NewCache(encryptionKey, nonExistentPath)
	if err != nil {
		t.Errorf("Did not expect an error for non-existent path, got: %v", err)
	}
}

func TestCacheImmediateExpiration(t *testing.T) {
	encryptionKey := []byte("32-byte-long-encryption-key-1234")
	c, _ := cache.NewCache(encryptionKey, "")

	key := "immediateExpireKey"
	value := []byte("value")

	c.Set(key, value, 0) // Immediate expiration
	_, exists := c.Get(key)
	if exists {
		t.Error("Expected no data for immediately expired key")
	}
}

func TestCacheNoExpiration(t *testing.T) {
	encryptionKey := []byte("32-byte-long-encryption-key-1234")
	c, _ := cache.NewCache(encryptionKey, "")

	key := "noExpireKey"
	value := []byte("value")

	c.Set(key, value, 100*365*24*time.Hour) // A very long duration to behave as no expiration
	_, exists := c.Get(key)
	if !exists {
		t.Error("Expected data for non-expired key")
	}
}
