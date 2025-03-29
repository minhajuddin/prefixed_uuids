package prefixed_uuids

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

type Entity int
type PrefixInfo struct {
	Entity Entity
	Prefix string
}

const separator = "."

var prefixAllowedCharsRegex = regexp.MustCompile(`^[a-z0-9_-]+$`)

func NewPrefixInfo(entity Entity, prefix string) PrefixInfo {
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

func NewRegistry(prefixes ...PrefixInfo) Registry {
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

func (r Registry) Deserialize(entity Entity, uuidStr string) (uuid.UUID, error) {
	parts := strings.Split(uuidStr, separator)
	if len(parts) != 2 {
		// TODO: sentinel error
		return uuid.Nil, fmt.Errorf("invalid uuid format")
	}
	prefix := parts[0]
	parsedEntity, ok := r.reverse[prefix]
	if !ok {
		return uuid.Nil, fmt.Errorf("unknown prefix")
	}

	if parsedEntity != entity {
		return uuid.Nil, fmt.Errorf("invalid entity")
	}

	uuidBytes, err := base64withNoPadding.DecodeString(parts[1])
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid uuid format")
	}
	return uuid.FromBytes(uuidBytes)
}
