# Prefixed UUIDs

A Go package that provides a type-safe way to work with prefixed UUIDs. This package allows you to create and parse UUIDs with entity-specific prefixes, making them more readable and ensuring type safety at runtime.

## Features

- Create prefixed UUIDs for different entity types
- Parse prefixed UUIDs back to their original UUID form
- Type-safe entity handling
- URL-safe base64 encoding for compact representation
- Runtime validation of entity types and prefixes
- Support for versioned entities (e.g., UserV2, UserV3)

## Installation

```bash
go get github.com/minhajuddin/prefixed_uuids
```

## Usage

### Basic Setup

First, define your entity types and create a registry:

```go
const (
    User      Entity = 1
    Post      Entity = 2
    Comment   Entity = 3
    UserV2    Entity = 5
    UserV3    Entity = 6
    SessionID Entity = 7
)

registry, err := NewRegistry([]PrefixInfo{
    {User, "user"},
    {Post, "post"},
    {Comment, "comment"},
    {UserV2, "user_v2"},
    {UserV3, "user_v3"},
    {SessionID, "sid"},
})
if err != nil {
    // Handle error
}
```

### Creating Prefixed UUIDs

```go
uuid := uuid.MustParse("0195e37b-f93f-7518-a9ac-a2be68463c7e")
prefixedUUID := registry.Serialize(User, uuid)
// Result: "user.AZXje_k_dRiprKK-aEY8fg"

// Versioned entities
prefixedUUIDV2 := registry.Serialize(UserV2, uuid)
// Result: "user_v2.AZXje_k_dRiprKK-aEY8fg"
```

### Parsing Prefixed UUIDs

You can parse prefixed UUIDs in two ways:

1. When you know the expected entity type:
```go
uuid, err := registry.Deserialize(User, "user.AZXje_k_dRiprKK-aEY8fg")
if err != nil {
    // Handle error
}
// uuid.String() == "0195e37b-f93f-7518-a9ac-a2be68463c7e"
```

2. When you want to determine the entity type from the prefix:
```go
entity, uuid, err := registry.DeserializeWithEntity("post.AZXje_k_dRiprKK-aEY8fg")
if err != nil {
    // Handle error
}
// entity == Post
// uuid.String() == "0195e37b-f93f-7518-a9ac-a2be68463c7e"
```

## Benefits

1. **Type Safety**: The package ensures that UUIDs are used with their correct entity types at runtime.
2. **Readability**: Prefixed UUIDs are more human-readable while maintaining their uniqueness.
3. **Compactness**: The base64 encoding makes the prefixed UUIDs shorter than the original hex representation.
4. **URL Safety**: The encoding is URL-safe, making it suitable for use in URLs and filenames.
5. **Versioning Support**: Easy support for versioned entities with distinct prefixes.

## Error Handling

The package returns errors in the following cases:

- `ErrInvalidPrefixedUUIDFormat`: When the input string doesn't match the expected format (e.g., empty string, no separator)
- `ErrUnknownPrefix`: When the prefix is not registered in the registry
- `ErrInvalidUUIDBadBase64`: When the base64 part is invalid
- `ErrInvalidUUIDFormat`: When the decoded bytes don't form a valid UUID
- `ErrEntityMismatch`: When the entity type doesn't match the prefix

Example error handling:
```go
// Invalid format
_, err := registry.Deserialize(User, "userAZXje_k_dRiprKK-aEY8fg")
// err == ErrInvalidPrefixedUUIDFormat

// Unknown prefix
_, err = registry.Deserialize(User, "unknown.AZXje_k_dRiprKK-aEY8fg")
// err == ErrUnknownPrefix

// Entity mismatch
prefixedUUID := registry.Serialize(User, uuid)
_, err = registry.Deserialize(Post, prefixedUUID)
// err == ErrEntityMismatch
```

## Prefix Rules

Prefixes must:
- Be lowercase
- Contain only alphanumeric characters, underscores, and hyphens
- Match the regex pattern: `^[a-z0-9_-]+$`

Invalid prefixes will cause `NewRegistry` to return an error:
```go
// These will all return errors:
registry, err := NewRegistry([]PrefixInfo{
    {Entity(1), "Test"},      // uppercase
    {Entity(1), "test prefix"}, // contains space
    {Entity(1), "test@prefix"}, // contains special char
})
```

## Example Use Cases

1. **Database IDs**: Use prefixed UUIDs as primary keys in your database to identify different types of entities.
2. **API Resources**: Use prefixed UUIDs in your API endpoints to identify resources.
3. **File Storage**: Use prefixed UUIDs as filenames to identify the type of content stored.
4. **Versioned Resources**: Use prefixed UUIDs to identify different versions of the same entity type.

## Size Comparison

The prefixed UUID format is more compact than the standard UUID format:
```go
uuid := uuid.MustParse("0195e37b-f93f-7518-a9ac-a2be68463c7e")
prefixedUUID := registry.Serialize(User, uuid)
// len(prefixedUUID) == 27  // "user.AZXje_k_dRiprKK-aEY8fg"
// len(uuid.String()) == 36 // "0195e37b-f93f-7518-a9ac-a2be68463c7e"
```

## Inspiration

- [Stripe's prefixed ids](https://docs.stripe.com/api) and the urn format https://datatracker.ietf.org/doc/html/rfc8141
- https://gist.github.com/fnky/76f533366f75cf75802c8052b577e2a5
