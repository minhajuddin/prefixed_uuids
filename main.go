package prefixed_uuids

import (
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

const defaultSeparator = "."

var (
	ErrEntityMismatch            = errors.New("entity mismatch")
	ErrInvalidPrefixedUUIDFormat = errors.New("invalid prefixed uuid format")
	ErrInvalidUUIDBadBase64      = errors.New("invalid uuid bad base64 part")
	ErrInvalidUUIDFormat         = errors.New("invalid uuid format")
	ErrUnknownPrefix             = errors.New("unknown prefix")
	ErrInvalidSeparator          = errors.New("invalid separator")
)
var (
	NullEntity                 Entity = 0
	base64withNoPadding               = base64.URLEncoding.WithPadding(base64.NoPadding)
	prefixAllowedCharsRegex           = regexp.MustCompile(`^[a-z0-9_-]+$`)
	separatorAllowedCharsRegex        = regexp.MustCompile(`^[\.~]$`)
)

type Entity int
type PrefixInfo struct {
	Entity Entity
	Prefix string
}

type Registry struct {
	prefixes  map[Entity]string
	reverse   map[string]Entity
	separator string
}

func NewRegistry(prefixes []PrefixInfo) (*Registry, error) {
	registry := &Registry{
		prefixes:  make(map[Entity]string, len(prefixes)),
		reverse:   make(map[string]Entity, len(prefixes)),
		separator: defaultSeparator,
	}
	for _, prefix := range prefixes {
		if prefix.Entity == NullEntity {
			return nil, fmt.Errorf("entity cannot be NullEntity, use a non-zero value")
		}
		if !prefixAllowedCharsRegex.MatchString(prefix.Prefix) {
			return nil, fmt.Errorf("prefix must be in lowercase and contain only alphanumeric characters, underscores, and hyphens")
		}

		registry.prefixes[prefix.Entity] = prefix.Prefix
		registry.reverse[prefix.Prefix] = prefix.Entity
	}
	return registry, nil
}

// WithSeparator sets a custom separator for the Registry.
// Only '.' and '~' are allowed as separators since they are
// not part of the base64url encoding alphabet and not encoded in URLs.
func (r *Registry) WithSeparator(separator string) (*Registry, error) {
	if !separatorAllowedCharsRegex.MatchString(separator) {
		return nil, fmt.Errorf("%w: only '.' and '~' are allowed", ErrInvalidSeparator)
	}
	r.separator = separator
	return r, nil
}

func (r *Registry) Serialize(entity Entity, uuid uuid.UUID) string {
	// MarshalBinary never returns an error
	uuidBytes, _ := uuid.MarshalBinary()
	return fmt.Sprintf("%s%s%s", r.prefixes[entity], r.separator, base64withNoPadding.EncodeToString(uuidBytes))
}

func (r *Registry) DeserializeWithEntity(uuidStr string) (Entity, uuid.UUID, error) {
	parts := strings.Split(uuidStr, r.separator)
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

func (r *Registry) Deserialize(entity Entity, uuidStr string) (uuid.UUID, error) {
	parsedEntity, parsedUUID, err := r.DeserializeWithEntity(uuidStr)
	if err != nil {
		return uuid.Nil, err
	}

	if parsedEntity != entity {
		return uuid.Nil, fmt.Errorf("%w", ErrEntityMismatch)
	}

	return parsedUUID, nil
}
