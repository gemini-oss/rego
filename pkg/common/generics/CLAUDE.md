# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `generics` package provides utilities for JSON marshaling/unmarshaling of generic structs with inline fields. It's designed to handle APIs that return different field types based on HTTP methods, enabling type-safe operations without runtime type assertions.

**Key Features:**
- Handles polymorphic JSON responses (different types for same field)
- Flattens inline fields during marshaling/unmarshaling
- Type-safe generic struct operations
- Runtime type introspection for generics

**Note:** Despite the main CLAUDE.md reference, this package does NOT provide:
- Generic collection operations (Filter, Map, Reduce)
- Pagination helpers (ProcessAllPagesOrdered)
- Those features are implemented directly in service packages

## Architecture

### Core Functions

1. **`MarshalGeneric[T any, M any](t *T) ([]byte, error)`**:
   - Marshals structs with generic fields to flattened JSON
   - Automatically detects and flattens inline fields
   - Merges method-specific fields into root object

2. **`UnmarshalGeneric[T any, M any](data []byte) (*T, error)`**:
   - Unmarshals JSON into generic structs
   - Distributes fields to appropriate struct members
   - Collects leftover fields into the generic field

3. **`DerefGeneric[E any]() (reflect.Type, bool)`**:
   - Returns concrete type information for generics
   - Indicates if type is a pointer
   - Used for runtime type introspection

4. **`Pointer[T any](val T) *T`**:
   - Simple utility to create pointers to values

### Design Pattern

The package enables this pattern for handling API inconsistencies:
```go
type Resource[M any] struct {
    *BaseFields   `json:",inline"`  // Common fields
    Method     M  `json:",inline"`  // Method-specific fields
}

// Different types for different HTTP methods
type ResourceGET struct {
    Status StatusObject `json:"status"`  // Object on GET
}
type ResourcePOST struct {
    Status string `json:"status"`       // String on POST
}
```

## Development Tasks

### Usage Example

1. **Define Base and Method-Specific Types**:
   ```go
   type UserBase struct {
       ID   int    `json:"id"`
       Name string `json:"name"`
   }

   type UserGET struct {
       Groups []Group `json:"groups"`  // Array on GET
   }

   type UserPOST struct {
       Groups string `json:"groups"`   // Comma-separated on POST
   }
   ```

2. **Create Generic Type**:
   ```go
   type User[M any] struct {
       *UserBase `json:",inline"`
       Method M  `json:",inline"`
   }
   ```

3. **Implement JSON Methods**:
   ```go
   func (u *User[M]) UnmarshalJSON(data []byte) error {
       user, err := generics.UnmarshalGeneric[User[M], M](data)
       if err != nil {
           return err
       }
       *u = *user
       return nil
   }
   ```

## Important Notes

- Only one generic field per struct is supported
- Use `json:",inline"` tag for fields to be flattened
- Generic field must be exported (capitalized)
- Handles both pointer and non-pointer generic fields
- Leftover JSON keys are assigned to the generic field

## Common Pitfalls

1. **Multiple Generic Fields**: Only one generic field is detected - others ignored
2. **Missing Inline Tags**: Without `,inline`, fields won't be flattened
3. **Type Constraints**: Generic field must be a struct or pointer to struct
4. **Field Conflicts**: Ensure no field name conflicts between base and method types

## Usage in Rego

### Primary Consumer: SnipeIT

The SnipeIT API returns different JSON structures for the same resource:
```go
// GET returns objects
{
  "warranty_expires": {
    "date": "2025-01-01",
    "formatted": "01/01/2025"
  }
}

// POST/PATCH expects strings
{
  "warranty_expires": "2025-01-01"
}

// Some strings even return as null
```

SnipeIT handles this with:
```go
type Hardware[M generics.M | generics.S] struct {
    *HardwareBase `json:",inline"`
    Method M      `json:",inline"`
}

// Method type changes based on operation
type HardwareGET struct {
    WarrantyExpires DateObject `json:"warranty_expires"`
}

type HardwarePOST struct {
    WarrantyExpires string `json:"warranty_expires"`
}
```

### Type Constants

The package defines marker types:
```go
type S struct{} // Selection (GET)
type M struct{} // Mutation (POST/PATCH/DELETE)
```

## Real-World Problem This Solves

Many APIs are inconsistent in their response formats:
- Return rich objects with metadata on GET
- Expect simple values on POST/PATCH
- Use different types for the same conceptual field

This package allows a single struct definition to handle both cases type-safely.

## Pagination Pattern (Not in This Package)

While the main CLAUDE.md mentions pagination helpers, they're actually implemented per-service:
- `doPaginated` in each service package
- Service-specific pagination interfaces
- No generic pagination utilities in this package
