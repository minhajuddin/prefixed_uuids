# Prefixed UUIDs

A Go package that provides a type-safe way to work with prefixed UUIDs. This package allows you to create and parse UUIDs with entity-specific prefixes, making them more readable and ensuring type safety at runtime.

## Features

- Create prefixed UUIDs for different entity types
- Parse prefixed UUIDs back to their original UUID form
- Type-safe entity handling
- URL-safe base64 encoding for compact representation
- Runtime validation of entity types and prefixes
- Support for versioned entities (e.g., UserV2, UserV3)
- Customizable separator character (defaults to `.`, can also use `~`)
- Multi UUID support for encoding multiple UUIDs with a single prefix

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

### Optional: Custom Separator

By default, the registry uses `.` as the separator. You can customize this using the fluent interface:

```go
// Use '~' as a separator instead of the default '.'
registry, err = registry.WithSeparator("~")
if err != nil {
    // Handle error
}
// This will produce IDs like "user~AZXje_k_dRiprKK-aEY8fg"
```

Note: Only `.` and `~` are allowed as separators since they are not part of the base64url encoding
alphabet and are not encoded in URLs.

### Creating Prefixed UUIDs

```go
uuid := uuid.MustParse("0195e37b-f93f-7518-a9ac-a2be68463c7e")
prefixedUUID := registry.Serialize(User, uuid)
// Result with default separator: "user.AZXje_k_dRiprKK-aEY8fg"
// Result with '~' separator: "user~AZXje_k_dRiprKK-aEY8fg"

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

### Multi UUIDs

For cases where you need to encode multiple related UUIDs into a single prefixed string (e.g., a composite key for a user's post comment), you can use multi types:

```go
const (
    User            Entity = 1
    Post            Entity = 2
    Comment         Entity = 3
    UserPost        Entity = 10
    UserPostComment Entity = 11
)

registry, err := NewRegistry([]PrefixInfo{
    {User, "user"},
    {Post, "post"},
    {Comment, "comment"},
})

// Register multi types — component entities must already be registered
registry.AddMultiPrefix(MultiPrefixInfo{UserPost, "up", []Entity{User, Post}})
registry.AddMultiPrefix(MultiPrefixInfo{UserPostComment, "upc", []Entity{User, Post, Comment}})
```

Serializing multi UUIDs:

```go
encoded, err := registry.SerializeMulti(UserPostComment,
    EntityUUID{User, userUUID},
    EntityUUID{Post, postUUID},
    EntityUUID{Comment, commentUUID},
)
// Result: "upc.AZXje_k_dRiprKK-aEY8fgGV43v5P3UYqayivmhGPH8BleN7-T91GKmsoryoRjyA"
```

Deserializing multi UUIDs:

```go
var parsedUser, parsedPost, parsedComment uuid.UUID
err := registry.DeserializeMulti(UserPostComment, encoded,
    EntityUUIDPtr{User, &parsedUser},
    EntityUUIDPtr{Post, &parsedPost},
    EntityUUIDPtr{Comment, &parsedComment},
)
```

Both `SerializeMulti` and `DeserializeMulti` enforce that the entity types are provided in the correct order matching the multi type definition.

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
- `ErrNotMultiEntity`: When using `SerializeMulti`/`DeserializeMulti` with a non-multi entity
- `ErrUUIDCountMismatch`: When the number of UUID pairs doesn't match the multi type definition
- `ErrEntityOrderMismatch`: When entities are provided in the wrong order for a multi type

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

The primary use case for this is to convert UUIDs to friendly external IDs when sending it to external systems.
Whenever you show the ID to the outside world, you use a prefixed_uuid and once it gets past your edge code, it is parsed properly.
This way, if you ask a customer for their id and they give you `session.AZXje_k_dRiprKK-aEY8fg` instead of `user.AZXje_k_dRiprKK-aEY8fg`, you'll know that they gave you the wrong ID. Also, looking at logs, whenever you see an id, you'll know exactly what it is referring to.

Example uses
- **Session IDs**: `sid.AZXje_k_dRiprKK-aEY8fg`
- **User ID**: `user.AZXje_k_dRiprKK-aEY8fg`
- **JWT Token ID**: `jti.AZXje_k_dRiprKK-aEY8fg`
- **Secret Key (Production)**: `sk_live.AZXje_k_dRiprKK-aEY8fg`
- **Secret Key (Test)**: `sk_test.AZXje_k_dRiprKK-aEY8fg`

## FAQs
1. Can I use this with integer IDs?
    No. However, feel free to fork this and change the code to serialize/deserialize ints. It is doable with just a few changes.
2. Why use `.` for the separator instead of `_` or `-`?
    `_` and `-` are part of the alphabet for the base64url encoding scheme that we use to encode the UUID bytes. To make the code more robust, we use a separator that is not part of that alphabet. Also, we don't use `:` because it is encoded in urls which is a minor annoyance. The only other separator that can be used other than `.` which is not encoded is `~`. You can now use either `.` or `~` as separators with the `WithSeparator` method.

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
