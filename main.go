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
	ErrNotMultiEntity            = errors.New("entity is not a multi type")
	ErrUUIDCountMismatch         = errors.New("number of uuids does not match multi type definition")
	ErrEntityOrderMismatch       = errors.New("entity at position does not match multi type definition")
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

type MultiPrefixInfo struct {
	Entity   Entity
	Prefix   string
	Entities []Entity
}

type EntityUUID struct {
	Entity Entity
	UUID   uuid.UUID
}

type EntityUUIDPtr struct {
	Entity Entity
	UUID   *uuid.UUID
}

type Registry struct {
	prefixes  map[Entity]string
	reverse   map[string]Entity
	separator string
	multi     map[Entity][]Entity
}

func NewRegistry(prefixes []PrefixInfo) (*Registry, error) {
	registry := &Registry{
		prefixes:  make(map[Entity]string, len(prefixes)),
		reverse:   make(map[string]Entity, len(prefixes)),
		separator: defaultSeparator,
		multi:     make(map[Entity][]Entity),
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

func (r *Registry) decodePayload(uuidStr string) (Entity, []byte, error) {
	parts := strings.Split(uuidStr, r.separator)
	if len(parts) != 2 {
		return NullEntity, nil, fmt.Errorf("%w", ErrInvalidPrefixedUUIDFormat)
	}
	prefix := parts[0]
	parsedEntity, ok := r.reverse[prefix]
	if !ok {
		return NullEntity, nil, fmt.Errorf("%w", ErrUnknownPrefix)
	}

	payload, err := base64withNoPadding.DecodeString(parts[1])
	if err != nil {
		return NullEntity, nil, errors.Join(err, ErrInvalidUUIDBadBase64)
	}
	return parsedEntity, payload, nil
}

func (r *Registry) DeserializeWithEntity(uuidStr string) (Entity, uuid.UUID, error) {
	parsedEntity, payload, err := r.decodePayload(uuidStr)
	if err != nil {
		return NullEntity, uuid.Nil, err
	}

	parsedUUID, err := uuid.FromBytes(payload)
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

func (r *Registry) AddMultiPrefix(info MultiPrefixInfo) error {
	if info.Entity == NullEntity {
		return fmt.Errorf("entity cannot be NullEntity, use a non-zero value")
	}
	if !prefixAllowedCharsRegex.MatchString(info.Prefix) {
		return fmt.Errorf("prefix must be in lowercase and contain only alphanumeric characters, underscores, and hyphens")
	}
	if _, exists := r.prefixes[info.Entity]; exists {
		return fmt.Errorf("entity %d is already registered", info.Entity)
	}
	if _, exists := r.reverse[info.Prefix]; exists {
		return fmt.Errorf("prefix %q is already registered", info.Prefix)
	}
	if len(info.Entities) < 2 {
		return fmt.Errorf("multi type must have at least 2 component entities")
	}
	for _, e := range info.Entities {
		if _, ok := r.prefixes[e]; !ok {
			return fmt.Errorf("component entity %d is not registered in the registry", e)
		}
	}

	r.prefixes[info.Entity] = info.Prefix
	r.reverse[info.Prefix] = info.Entity
	r.multi[info.Entity] = info.Entities
	return nil
}

func (r *Registry) SerializeMulti(entity Entity, pairs ...EntityUUID) (string, error) {
	components, ok := r.multi[entity]
	if !ok {
		return "", fmt.Errorf("%w", ErrNotMultiEntity)
	}
	if len(pairs) != len(components) {
		return "", fmt.Errorf("%w: expected %d, got %d", ErrUUIDCountMismatch, len(components), len(pairs))
	}

	buf := make([]byte, 0, len(components)*16)
	for i, pair := range pairs {
		if pair.Entity != components[i] {
			return "", fmt.Errorf("%w: position %d expected entity %d, got %d", ErrEntityOrderMismatch, i, components[i], pair.Entity)
		}
		uuidBytes, _ := pair.UUID.MarshalBinary()
		buf = append(buf, uuidBytes...)
	}

	return fmt.Sprintf("%s%s%s", r.prefixes[entity], r.separator, base64withNoPadding.EncodeToString(buf)), nil
}

func (r *Registry) DeserializeMulti(entity Entity, uuidStr string, targets ...EntityUUIDPtr) error {
	parsedEntity, payload, err := r.decodePayload(uuidStr)
	if err != nil {
		return err
	}
	if parsedEntity != entity {
		return fmt.Errorf("%w", ErrEntityMismatch)
	}

	components, ok := r.multi[entity]
	if !ok {
		return fmt.Errorf("%w", ErrNotMultiEntity)
	}
	if len(targets) != len(components) {
		return fmt.Errorf("%w: expected %d, got %d", ErrUUIDCountMismatch, len(components), len(targets))
	}
	for i, target := range targets {
		if target.Entity != components[i] {
			return fmt.Errorf("%w: position %d expected entity %d, got %d", ErrEntityOrderMismatch, i, components[i], target.Entity)
		}
	}

	expectedLen := len(components) * 16
	if len(payload) != expectedLen {
		return fmt.Errorf("%w", ErrInvalidUUIDFormat)
	}

	for i, target := range targets {
		chunk := payload[i*16 : (i+1)*16]
		parsed, err := uuid.FromBytes(chunk)
		if err != nil {
			return errors.Join(err, ErrInvalidUUIDFormat)
		}
		*target.UUID = parsed
	}

	return nil
}
