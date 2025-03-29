package prefixed_uuids

import (
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidEntity             = errors.New("invalid entity")
	ErrInvalidPrefixedUUIDFormat = errors.New("invalid prefixed uuid format")
	ErrInvalidUUIDBadBase64      = errors.New("invalid uuid bad base64 part")
	ErrInvalidUUIDFormat         = errors.New("invalid uuid format")
	ErrUnknownPrefix             = errors.New("unknown prefix")
)

type Entity int

var NullEntity Entity = 0

type PrefixInfo struct {
	Entity Entity
	Prefix string
}

const separator = "."

var prefixAllowedCharsRegex = regexp.MustCompile(`^[a-z0-9_-]+$`)

func NewPrefixInfo(entity Entity, prefix string) PrefixInfo {
	if entity == NullEntity {
		panic("entity cannot be NullEntity, use a non-zero value")
	}
	if !prefixAllowedCharsRegex.MatchString(prefix) {
		panic("prefix must be in lowercase and contain only alphanumeric characters, underscores, and hyphens")
	}
	return PrefixInfo{
		Entity: entity,
		Prefix: prefix,
	}
}

type Registry struct {
	prefixes map[Entity]string
	reverse  map[string]Entity
}

var base64withNoPadding = base64.URLEncoding.WithPadding(base64.NoPadding)

func (r Registry) Serialize(entity Entity, uuid uuid.UUID) string {
	// MarshalBinary never returns an error
	uuidBytes, _ := uuid.MarshalBinary()
	return fmt.Sprintf("%s.%s", r.prefixes[entity], base64withNoPadding.EncodeToString(uuidBytes))
}

func NewRegistry(prefixes []PrefixInfo) Registry {
	registry := Registry{
		prefixes: make(map[Entity]string),
		reverse:  make(map[string]Entity),
	}
	for _, prefix := range prefixes {
		registry.prefixes[prefix.Entity] = prefix.Prefix
		registry.reverse[prefix.Prefix] = prefix.Entity
	}
	return registry
}

func (r Registry) DeserializeWithEntity(uuidStr string) (Entity, uuid.UUID, error) {
	parts := strings.Split(uuidStr, separator)
	if len(parts) != 2 {
		return NullEntity, uuid.Nil, fmt.Errorf("%w", ErrInvalidPrefixedUUIDFormat)
	}
	prefix := parts[0]
	parsedEntity, ok := r.reverse[prefix]
	if !ok {
		return NullEntity, uuid.Nil, fmt.Errorf("%w", ErrUnknownPrefix)
	}

	uuidBytes, err := base64withNoPadding.DecodeString(parts[1])
	if err != nil {
		return NullEntity, uuid.Nil, errors.Join(err, ErrInvalidUUIDBadBase64)
	}
	parsedUUID, err := uuid.FromBytes(uuidBytes)
	if err != nil {
		return NullEntity, uuid.Nil, errors.Join(err, ErrInvalidUUIDFormat)
	}
	return parsedEntity, parsedUUID, nil
}

func (r Registry) Deserialize(entity Entity, uuidStr string) (uuid.UUID, error) {
	parsedEntity, parsedUUID, err := r.DeserializeWithEntity(uuidStr)
	if err != nil {
		return uuid.Nil, err
	}

	if parsedEntity != entity {
		return uuid.Nil, fmt.Errorf("%w", ErrInvalidEntity)
	}

	return parsedUUID, nil
}
