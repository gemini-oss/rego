// pkg/internal/tests/common/cache/cache_test.go
package cache_test

import (
	"bytes"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gemini-oss/rego/pkg/common/cache"
)

func TestCacheSetAndGet(t *testing.T) {
	key := "testKey"
	value := []byte("testValue")
	encryptionKey := []byte("32~Byte-long_passphrase-key-1234") // 32 bytes

	c, err := cache.NewCache(encryptionKey, true)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Test setting and getting a value
	err = c.Set(key, value, 1*time.Minute)
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

func TestCacheCreateFail(t *testing.T) {
	encryptionKey := []byte("a-very-very-very-very-secret-key") // 32 bytes

	// We expect an error because the encryption key is
	_, err := cache.NewCache(encryptionKey, true)
	if err == nil {
		t.Fatal("Cache creation succeeded with invalid key")
	}
}

func TestCacheExpiration(t *testing.T) {
	key := "expireKey"
	value := []byte("expireValue")
	encryptionKey := []byte("32~Byte-long_passphrase-key-1234") // 32 bytes

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
	encryptionKey := []byte("32~Byte-long_passphrase-key-1234")
	c, _ := cache.NewCache(encryptionKey, true)

	_, exists := c.Get(key)
	if exists {
		t.Error("Get() found a value for a non-existent key")
	}
}

func TestCachePersistence(t *testing.T) {
	key := "persistKey"
	value := []byte("persistValue")
	encryptionKey := []byte("32~Byte-long_passphrase-key-1234")
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

func TestCacheConcurrentReads(t *testing.T) {
	encryptionKey := []byte("32~Byte-long_passphrase-key-1234")
	c, err := cache.NewCache(encryptionKey, true)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Prepopulate the cache with multiple items
	numItems := 10
	for i := 0; i < numItems; i++ {
		key := "key" + strconv.Itoa(i)
		value := []byte("value" + strconv.Itoa(i))
		err := c.Set(key, value, 5*time.Minute)
		if err != nil {
			t.Fatalf("Failed to set key %s: %v", key, err)
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < numItems; i++ {
		wg.Add(1)
		go func(keySuffix int) {
			defer wg.Done()
			key := "key" + strconv.Itoa(keySuffix)
			expectedValue := []byte("value" + strconv.Itoa(keySuffix))

			value, exists := c.Get(key)
			if !exists {
				t.Errorf("Key %s does not exist", key)
			}
			if !bytes.Equal(value, expectedValue) {
				t.Errorf("Value mismatch for key %s: got %v, want %v", key, value, expectedValue)
			}
		}(i)
	}
	wg.Wait()
}

func TestCacheConcurrentWrites(t *testing.T) {
	encryptionKey := []byte("32~Byte-long_passphrase-key-1234")
	c, err := cache.NewCache(encryptionKey, true)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	var wg sync.WaitGroup
	numWorkers := 10
	writeIterations := 5

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < writeIterations; j++ {
				key := "key" + strconv.Itoa(workerID)
				value := []byte("value" + strconv.Itoa(workerID) + "_" + strconv.Itoa(j))
				err := c.Set(key, value, time.Minute)
				if err != nil {
					t.Errorf("Failed to set key %s: %v", key, err)
				}

				// Optional: Read after write to verify
				retrievedValue, exists := c.Get(key)
				if !exists {
					t.Errorf("Key %s does not exist after set", key)
				}
				if !bytes.Equal(retrievedValue, value) {
					t.Errorf("Value mismatch for key %s: got %v, want %v", key, retrievedValue, value)
				}
			}
		}(i)
	}
	wg.Wait()
}

func TestCacheInvalidKey(t *testing.T) {
	invalidKey := []byte("invalid-encryption-key")
	_, err := cache.NewCache(invalidKey, true)
	if err == nil {
		t.Errorf("Expected an error for invalid key size, but got none")
	}
}

func TestCachePersistenceFailure(t *testing.T) {
	encryptionKey := []byte("32~Byte-long_passphrase-key-1234")
	nonExistentPath := "/non/existent/path/cache.gob"

	// Expecting no error even though the path doesn't exist
	_, err := cache.NewCache(encryptionKey, nonExistentPath)
	if err != nil {
		t.Errorf("Did not expect an error for non-existent path, got: %v", err)
	}
}

func TestCacheImmediateExpiration(t *testing.T) {
	encryptionKey := []byte("32~Byte-long_passphrase-key-1234")
	c, _ := cache.NewCache(encryptionKey, true)

	key := "immediateExpireKey"
	value := []byte("value")

	c.Set(key, value, 0) // Immediate expiration
	_, exists := c.Get(key)
	if exists {
		t.Error("Expected no data for immediately expired key")
	}
}

func TestCacheNoExpiration(t *testing.T) {
	encryptionKey := []byte("32~Byte-long_passphrase-key-1234")
	c, _ := cache.NewCache(encryptionKey, true)

	key := "noExpireKey"
	value := []byte("value")

	c.Set(key, value, 100*365*24*time.Hour) // A very long duration to behave as no expiration
	_, exists := c.Get(key)
	if !exists {
		t.Error("Expected data for non-expired key")
	}
}

func TestCacheLRUEviction(t *testing.T) {
	encryptionKey := []byte("32~Byte-long_passphrase-key-1234")
	maxItems := 5
	c, _ := cache.NewCache(maxItems, encryptionKey, true)

	// Add items to the cache, exceeding the maxItems limit
	for i := 0; i < maxItems+1; i++ {
		key := "key" + strconv.Itoa(i)
		value := []byte("value" + strconv.Itoa(i))
		c.Set(key, value, time.Minute)
	}

	// The first item should be evicted
	_, exists := c.Get("key0")
	if exists {
		t.Error("Expected the first item to be evicted, but it was not")
	}
}

func TestCacheDataCompression(t *testing.T) {
	encryptionKey := []byte("32~Byte-long_passphrase-key-1234")
	c, _ := cache.NewCache(encryptionKey, true)

	largeValue := make([]byte, 1024*1024) // 1MB
	key := "largeDataKey"

	err := c.Set(key, largeValue, time.Minute)
	if err != nil {
		t.Fatalf("Error setting large data: %v", err)
	}

	retrievedValue, exists := c.Get(key)
	if !exists {
		t.Fatal("Failed to retrieve the set large data")
	}

	if !bytes.Equal(retrievedValue, largeValue) {
		t.Error("Retrieved data does not match the original data")
	}
}

func TestCachePersistenceWithLargeData(t *testing.T) {
	encryptionKey := []byte("32~Byte-long_passphrase-key-1234")
	tempFile := "temp_large_data_cache.gob"
	defer os.Remove(tempFile)

	c, _ := cache.NewCache(encryptionKey, tempFile)

	largeValue := make([]byte, 1024*1024*10) // 10MB
	key := "largeDataKey"

	err := c.Set(key, largeValue, time.Minute)
	if err != nil {
		t.Fatalf("Error setting large data: %v", err)
	}

	// Create a new cache instance and load from disk
	newCache, _ := cache.NewCache(encryptionKey, tempFile)
	retrievedValue, exists := newCache.Get(key)
	if !exists {
		t.Fatalf("Failed to retrieve the set large data from new cache instance")
	}

	if !bytes.Equal(retrievedValue, largeValue) {
		t.Error("Retrieved data does not match the original data in new cache instance")
	}
}

func TestCacheExpirationUpdateOnAccess(t *testing.T) {
	encryptionKey := []byte("32~Byte-long_passphrase-key-1234")
	maxItems := 3
	c, _ := cache.NewCache(maxItems, encryptionKey, true)

	// Set and access the first key to update its expiration
	firstKey := "key1"
	c.Set(firstKey, []byte("value1"), 100*time.Millisecond)
	time.Sleep(50 * time.Millisecond) // Wait some time and access the first key
	c.Get(firstKey)

	// Fill up the cache
	c.Set("key2", []byte("value2"), time.Minute)
	c.Set("key3", []byte("value3"), time.Minute)

	// Wait for the first key's original expiration to pass
	time.Sleep(60 * time.Millisecond)

	// The first key should not be evicted since it was accessed recently
	_, exists := c.Get(firstKey)
	if !exists {
		t.Error("Expected the first key to be updated and not evicted, but it was evicted")
	}
}
