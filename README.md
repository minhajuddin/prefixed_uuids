# Prefixed UUIDs

A Go package that provides a type-safe way to work with prefixed UUIDs. This package allows you to create and parse UUIDs with entity-specific prefixes, making them more readable and ensuring type safety at runtime.

## Features

- Create prefixed UUIDs for different entity types
- Parse prefixed UUIDs back to their original UUID form
- Type-safe entity handling
- URL-safe base64 encoding for compact representation
- Runtime validation of entity types and prefixes

## Installation

```bash
go get github.com/minhajuddin/prefixed_uuids
```

## Usage

### Basic Setup

First, define your entity types and create a registry:

```go
const (
    User    Entity = 1
    Post    Entity = 2
    Comment Entity = 3
)

var registry = NewRegistry([]PrefixInfo{
    {User, "user"},
    {Post, "post"},
    {Comment, "comment"},
})
```

### Creating Prefixed UUIDs

```go
uuid := uuid.MustParse("0195e37b-f93f-7518-a9ac-a2be68463c7e")
prefixedUUID := registry.Serialize(User, uuid)
// Result: "user.AZXje_k_dRiprKK-aEY8fg"
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

## Error Handling

The package returns errors in the following cases:
- Invalid UUID format
- Unknown prefix
- Entity type mismatch
- Invalid base64 encoding

## Prefix Rules

Prefixes must:
- Be lowercase
- Contain only alphanumeric characters, underscores, and hyphens
- Match the regex pattern: `^[a-z0-9_-]+$`

## Example Use Cases

1. **Database IDs**: Use prefixed UUIDs as primary keys in your database to identify different types of entities.
2. **API Resources**: Use prefixed UUIDs in your API endpoints to identify resources.
3. **File Storage**: Use prefixed UUIDs as filenames to identify the type of content stored.


## Inspiration

- [Stripe's prefixed ids](https://docs.stripe.com/api) and the urn format https://datatracker.ietf.org/doc/html/rfc8141
- https://gist.github.com/fnky/76f533366f75cf75802c8052b577e2a5
