# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `crypt` package provides secure encryption/decryption using AES-256-GCM with Argon2 ID key derivation. It enforces strong passphrase requirements and offers cryptographically secure random number generation for security-critical operations throughout Rego.

**Key Features:**
- AES-256-GCM authenticated encryption
- Argon2 ID memory-hard key derivation
- Strict passphrase validation (32-128 bytes)
- Cryptographically secure random number generation
- Thread-safe stateless operations
- Base64 encoding for encrypted output
- Integration with cache package for data protection

## Architecture

### Core Components

1. **Encryption/Decryption**:
   - `EncryptAES()`: Encrypts data with AES-GCM, returns base64
   - `DecryptAES()`: Decrypts base64 data
   - Uses random salts and nonces for each operation

2. **Key Derivation**:
   - Argon2 ID with fixed parameters:
     - Time: 1 iteration
     - Memory: 64 MB
     - Parallelism: 4 threads
     - Output: 32-byte key

3. **Passphrase Validation**:
   - `ValidPassphrase()`: Enforces security requirements
   - Length: 32-128 bytes
   - Must contain uppercase and special characters
   - Special chars: `!@#$%^&*()-_=+[]{}|;:',.<>/?`~`
   - No repeated patterns (3-6 chars)
   - Minimum Shannon entropy: 4 bits
   - Returns `PassphraseError` with specific issues

4. **Utilities**:
   - `SecureRandomInt()`: Cryptographically secure random integers
   - `generateSalt()`: 16-byte random salts

### Encryption Format

```
[Salt (16 bytes)][Nonce (12 bytes)][Ciphertext (variable)][Auth Tag (16 bytes)]
                 └─────────── Encrypted with AES-GCM ────────────┘
└──────────────────── Base64 encoded ─────────────────────────────────┘
```

## Development Tasks

### Common Operations

1. **Basic Encryption/Decryption**:
   ```go
   // Encrypt data
   encrypted, err := crypt.EncryptAES(data, passphrase)
   if err != nil {
       // Could be: passphrase validation or encryption failure
       return err
   }

   // Decrypt data
   decrypted, err := crypt.DecryptAES(encrypted, passphrase)
   if err != nil {
       // Could be: wrong passphrase, corrupted data, or invalid format
       // Package doesn't distinguish between these cases
       return err
   }
   ```

2. **Passphrase Validation with Error Details**:
   ```go
   if err := crypt.ValidPassphrase(passphrase); err != nil {
       if passphraseErr, ok := err.(*crypt.PassphraseError); ok {
           // Access specific validation issues
           for _, issue := range passphraseErr.Issues {
               log.Printf("Passphrase issue: %s", issue)
           }
       }
       return err
   }
   ```

3. **Secure Random Numbers**:
   ```go
   // Generate random index (0-99)
   randomIndex, err := crypt.SecureRandomInt(100)

   // Note: Has modulo bias for non-power-of-2 values
   // For cryptographic selection, consider power-of-2 ranges
   ```

4. **Integration with Environment Variables**:
   ```go
   // Common pattern in Rego services
   encryptionKey := []byte(config.GetEnv("REGO_ENCRYPTION_KEY"))
   if err := crypt.ValidPassphrase(encryptionKey); err != nil {
       log.Fatal("Invalid REGO_ENCRYPTION_KEY: ", err)
   }
   ```

## Important Notes

- **Passphrase Requirements**: Very strict (32+ bytes) - designed for system keys, not user passwords
- **No Password Storage**: This package doesn't handle secure passphrase storage
- **Authenticated Encryption**: GCM mode provides both encryption and integrity
- **Fixed Parameters**: Argon2 parameters are hardcoded for consistency
- **Thread Safety**: All functions are stateless and thread-safe
- **Performance**: Key derivation happens on every encrypt/decrypt (no caching)
- **No Streaming**: Entire data must fit in memory
- **No Key Rotation**: No built-in mechanism for changing encryption keys

### Error Handling

```go
// Encryption errors are generic
encrypted, err := crypt.EncryptAES(data, passphrase)
if err != nil {
    // Could be:
    // - Invalid passphrase (ValidPassphrase failed)
    // - Encryption failure (rare)
}

// Decryption errors don't specify cause
decrypted, err := crypt.DecryptAES(encrypted, passphrase)
if err != nil {
    // Could be:
    // - Wrong passphrase
    // - Corrupted data
    // - Invalid base64
    // - Modified ciphertext (auth tag mismatch)
}
```

## Security Considerations

1. **Strong KDF**: Argon2 ID is memory-hard, resisting GPU/ASIC attacks
2. **No Key Reuse**: Fresh salt and nonce for every encryption
3. **Integrity Protection**: GCM mode detects tampering
4. **Timing Safety**: Uses constant-time comparisons in GCM
5. **Minimum Security**: 32 bytes = 256-bit keys
6. **Environment Variables**: Use for system keys, not hardcoded
7. **Memory Zeroing**: Passphrases not zeroed after use (Go limitation)

### Best Practices

```go
// Generate strong system key
// Option 1: Use a password manager
// Option 2: Generate cryptographically
openssl rand -base64 32

// Store in environment
export REGO_ENCRYPTION_KEY="generated-32-byte-key-here"

// Validate on startup
func init() {
    key := []byte(os.Getenv("REGO_ENCRYPTION_KEY"))
    if err := crypt.ValidPassphrase(key); err != nil {
        log.Fatal("Invalid encryption key:", err)
    }
}
```

## Common Pitfalls

1. **Passphrase Length**: 32 bytes minimum is very long for human-memorable passphrases
2. **Base64 Overhead**: Encrypted data is ~33% larger due to encoding
3. **Generic Errors**: Decryption failures don't specify if it's wrong password or corruption
5. **No Versioning**: Can't upgrade algorithms without breaking compatibility
6. **Modulo Bias**: `SecureRandomInt()` has bias for non-power-of-2 max values
7. **Key Generation**: Don't use user passwords - generate proper random keys

## Testing Guidelines

```go
// Test passphrase that meets all requirements
const testPassphrase = "8jCcfHzjg*8mXD8qWjj9mk*QNZnVsMRt"

// Round-trip test
func TestEncryptDecrypt(t *testing.T) {
    data := []byte("sensitive data")
    passphrase := []byte(testPassphrase)

    encrypted, err := crypt.EncryptAES(data, passphrase)
    require.NoError(t, err)
    require.NotEqual(t, data, encrypted) // Ensure it's encrypted

    decrypted, err := crypt.DecryptAES(encrypted, passphrase)
    require.NoError(t, err)
    require.Equal(t, data, decrypted)
}

// Test wrong passphrase
func TestWrongPassphrase(t *testing.T) {
    encrypted, _ := crypt.EncryptAES([]byte("data"), []byte(testPassphrase))
    _, err := crypt.DecryptAES(encrypted, []byte("wrong-passphrase-32-chars-long!!"))
    require.Error(t, err) // Should fail
}
```

## Usage in Rego

### Primary Consumer: Cache Package

```go
// From cache.NewCache()
if err := crypt.ValidPassphrase(encryptionKey); err != nil {
    return nil, err
}

// Cache encrypts all data before storage
encrypted, err := crypt.EncryptAES(data, c.encryptionKey)
```

### Environment Variable Pattern

All services that use caching require:
```bash
export REGO_ENCRYPTION_KEY="your-32-to-128-byte-key-here"
```

## Future Considerations

1. **Algorithm Versioning**: Add version header to support upgrades
2. **Key Rotation**: Implement key rotation mechanism
3. **Streaming Support**: For large files that don't fit in memory
4. **Hardware Acceleration**: Use AES-NI when available
5. **Alternative KDFs**: Support for scrypt, bcrypt for different use cases
6. **Key Management**: Integration with KMS systems
