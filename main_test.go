package prefixed_uuids

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Start of setup
const (
	User      Entity = 1
	Post      Entity = 2
	Comment   Entity = 3
	Other     Entity = 4
	UserV2    Entity = 5
	UserV3    Entity = 6
	SessionID Entity = 7
)

var prefixer *Registry

func init() {
	var err error
	prefixer, err = NewRegistry([]PrefixInfo{
		{SessionID, "sid"},
		{User, "user"},
		{UserV2, "user_v2"},
		{UserV3, "user_v3"},
		{Post, "post"},
		{Comment, "comment"},
		{Other, "other"},
	})
	if err != nil {
		panic(err)
	}
}

func TestPrefixes(t *testing.T) {
	u, err := uuid.Parse("0195e37b-f93f-7518-a9ac-a2be68463c7e")
	assert.NoError(t, err)
	assert.Equal(t, "user.AZXje_k_dRiprKK-aEY8fg", prefixer.Serialize(User, u))
	assert.Equal(t, "user_v2.AZXje_k_dRiprKK-aEY8fg", prefixer.Serialize(UserV2, u))

	u, err = prefixer.Deserialize(User, "user.AZXje_k_dRiprKK-aEY8fg")
	assert.NoError(t, err)
	assert.Equal(t, u.String(), "0195e37b-f93f-7518-a9ac-a2be68463c7e")

	// Shorter than hex uuid and urlsafe base64
	assert.Equal(t, len(prefixer.Serialize(User, u)), 27) // user.AZXje_k_dRiprKK-aEY8fg
	assert.Equal(t, len(u.String()), 36)                  // 0195e37b-f93f-7518-a9ac-a2be68463c7e

	e, u, err := prefixer.DeserializeWithEntity("post.AZXje_k_dRiprKK-aEY8fg")
	assert.NoError(t, err)
	assert.Equal(t, e, Post)
	assert.Equal(t, u.String(), "0195e37b-f93f-7518-a9ac-a2be68463c7e")

	assert.Equal(t, prefixer.Serialize(SessionID, u), "sid.AZXje_k_dRiprKK-aEY8fg")
}

func TestErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError error
	}{
		{
			name:          "empty string",
			input:         "",
			expectedError: ErrInvalidPrefixedUUIDFormat,
		},
		{
			name:          "no separator",
			input:         "userAZXje_k_dRiprKK-aEY8fg",
			expectedError: ErrInvalidPrefixedUUIDFormat,
		},
		{
			name:          "multiple separators",
			input:         "user.AZXje_k_dRiprKK.aEY8fg",
			expectedError: ErrInvalidPrefixedUUIDFormat,
		},
		{
			name:          "unknown prefix",
			input:         "unknown.AZXje_k_dRiprKK-aEY8fg",
			expectedError: ErrUnknownPrefix,
		},
		{
			name:          "invalid base64",
			input:         "user.invalid-base64!",
			expectedError: ErrInvalidUUIDBadBase64,
		},
		{
			name:          "invalid uuid bytes",
			input:         "user.AAAAAA", // too short to be valid UUID bytes
			expectedError: ErrInvalidUUIDFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test DeserializeWithEntity
			entity, u, err := prefixer.DeserializeWithEntity(tt.input)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, tt.expectedError))
			assert.Equal(t, entity, NullEntity)
			assert.Equal(t, u, uuid.UUID{})

			// Test Deserialize with User entity
			u, err = prefixer.Deserialize(User, tt.input)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, tt.expectedError))
			assert.Equal(t, u, uuid.UUID{})
		})
	}
}

func TestEntityMismatch(t *testing.T) {
	u, err := uuid.Parse("0195e37b-f93f-7518-a9ac-a2be68463c7e")
	assert.NoError(t, err)

	// Test when entity type doesn't match
	prefixedUUID := prefixer.Serialize(User, u)
	parsedUUID, err := prefixer.Deserialize(Post, prefixedUUID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrEntityMismatch))
	assert.Equal(t, parsedUUID, uuid.UUID{})
}

func TestRegistryCreation(t *testing.T) {
	tests := []struct {
		name          string
		prefixes      []PrefixInfo
		expectedError string
	}{
		{
			name: "null entity",
			prefixes: []PrefixInfo{
				{NullEntity, "test"},
			},
			expectedError: "entity cannot be NullEntity",
		},
		{
			name: "uppercase prefix",
			prefixes: []PrefixInfo{
				{Entity(100), "Test"},
			},
			expectedError: "prefix must be in lowercase",
		},
		{
			name: "prefix with spaces",
			prefixes: []PrefixInfo{
				{Entity(100), "test prefix"},
			},
			expectedError: "prefix must be in lowercase",
		},
		{
			name: "prefix with special chars",
			prefixes: []PrefixInfo{
				{Entity(100), "test@prefix"},
			},
			expectedError: "prefix must be in lowercase",
		},
		{
			name: "valid prefix",
			prefixes: []PrefixInfo{
				{Entity(100), "test-prefix_123"},
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry, err := NewRegistry(tt.prefixes)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, registry)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, registry)
			assert.Equal(t, tt.prefixes[0].Prefix, registry.prefixes[tt.prefixes[0].Entity])
			assert.Equal(t, tt.prefixes[0].Entity, registry.reverse[tt.prefixes[0].Prefix])
		})
	}
}
