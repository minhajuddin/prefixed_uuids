package prefixed_uuids

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Start of setup
const (
	User            Entity = 1
	Post            Entity = 2
	Comment         Entity = 3
	Other           Entity = 4
	UserV2          Entity = 5
	UserV3          Entity = 6
	SessionID       Entity = 7
	UserPost        Entity = 10
	UserPostComment Entity = 11
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

	err = prefixer.AddMultiPrefix(MultiPrefixInfo{UserPost, "up", []Entity{User, Post}})
	if err != nil {
		panic(err)
	}
	err = prefixer.AddMultiPrefix(MultiPrefixInfo{UserPostComment, "upc", []Entity{User, Post, Comment}})
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

func TestCustomSeparator(t *testing.T) {
	// Create a new registry with a custom separator
	customRegistry, err := NewRegistry([]PrefixInfo{
		{User, "user"},
		{Post, "post"},
	})
	assert.NoError(t, err)

	// Set custom separator using fluent interface
	customRegistry, err = customRegistry.WithSeparator("~")
	assert.NoError(t, err)

	// Verify serialize works with custom separator
	u, err := uuid.Parse("0195e37b-f93f-7518-a9ac-a2be68463c7e")
	assert.NoError(t, err)

	prefixedUUID := customRegistry.Serialize(User, u)
	assert.Equal(t, "user~AZXje_k_dRiprKK-aEY8fg", prefixedUUID)

	// Verify deserialize works with custom separator
	parsedUUID, err := customRegistry.Deserialize(User, prefixedUUID)
	assert.NoError(t, err)
	assert.Equal(t, u.String(), parsedUUID.String())

	// Test deserialize with entity
	entity, parsedUUID, err := customRegistry.DeserializeWithEntity(prefixedUUID)
	assert.NoError(t, err)
	assert.Equal(t, User, entity)
	assert.Equal(t, u.String(), parsedUUID.String())

	// Test with invalid separator
	_, err = customRegistry.WithSeparator(":")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidSeparator))

	// Test with another valid separator
	customRegistry, err = customRegistry.WithSeparator(".")
	assert.NoError(t, err)

	// Verify it works with the changed separator
	prefixedUUID = customRegistry.Serialize(User, u)
	assert.Equal(t, "user.AZXje_k_dRiprKK-aEY8fg", prefixedUUID)
}

func TestMultiRoundTrip(t *testing.T) {
	userUUID := uuid.MustParse("0195e37b-f93f-7518-a9ac-a2be68463c7e")
	postUUID := uuid.MustParse("0195e37b-f93f-7518-a9ac-a2be68463c7f")
	commentUUID := uuid.MustParse("0195e37b-f93f-7518-a9ac-a2be68463c80")

	// Test 2-UUID multi type
	encoded, err := prefixer.SerializeMulti(UserPost,
		EntityUUID{User, userUUID},
		EntityUUID{Post, postUUID},
	)
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(encoded, "up."))

	var parsedUser, parsedPost uuid.UUID
	err = prefixer.DeserializeMulti(UserPost, encoded,
		EntityUUIDPtr{User, &parsedUser},
		EntityUUIDPtr{Post, &parsedPost},
	)
	assert.NoError(t, err)
	assert.Equal(t, userUUID, parsedUser)
	assert.Equal(t, postUUID, parsedPost)

	// Test 3-UUID multi type
	encoded, err = prefixer.SerializeMulti(UserPostComment,
		EntityUUID{User, userUUID},
		EntityUUID{Post, postUUID},
		EntityUUID{Comment, commentUUID},
	)
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(encoded, "upc."))

	var parsedUser2, parsedPost2, parsedComment uuid.UUID
	err = prefixer.DeserializeMulti(UserPostComment, encoded,
		EntityUUIDPtr{User, &parsedUser2},
		EntityUUIDPtr{Post, &parsedPost2},
		EntityUUIDPtr{Comment, &parsedComment},
	)
	assert.NoError(t, err)
	assert.Equal(t, userUUID, parsedUser2)
	assert.Equal(t, postUUID, parsedPost2)
	assert.Equal(t, commentUUID, parsedComment)
}

func TestMultiUUIDCountMismatch(t *testing.T) {
	userUUID := uuid.MustParse("0195e37b-f93f-7518-a9ac-a2be68463c7e")

	// Too few pairs for serialize
	_, err := prefixer.SerializeMulti(UserPost,
		EntityUUID{User, userUUID},
	)
	assert.ErrorIs(t, err, ErrUUIDCountMismatch)

	// Too many pairs for serialize
	_, err = prefixer.SerializeMulti(UserPost,
		EntityUUID{User, userUUID},
		EntityUUID{Post, userUUID},
		EntityUUID{Comment, userUUID},
	)
	assert.ErrorIs(t, err, ErrUUIDCountMismatch)

	// Wrong count for deserialize
	encoded, err := prefixer.SerializeMulti(UserPost,
		EntityUUID{User, userUUID},
		EntityUUID{Post, userUUID},
	)
	assert.NoError(t, err)

	var parsedUser uuid.UUID
	err = prefixer.DeserializeMulti(UserPost, encoded,
		EntityUUIDPtr{User, &parsedUser},
	)
	assert.ErrorIs(t, err, ErrUUIDCountMismatch)
}

func TestMultiEntityOrderMismatch(t *testing.T) {
	userUUID := uuid.MustParse("0195e37b-f93f-7518-a9ac-a2be68463c7e")
	postUUID := uuid.MustParse("0195e37b-f93f-7518-a9ac-a2be68463c7f")

	// Serialize with wrong order
	_, err := prefixer.SerializeMulti(UserPost,
		EntityUUID{Post, postUUID},
		EntityUUID{User, userUUID},
	)
	assert.ErrorIs(t, err, ErrEntityOrderMismatch)

	// Deserialize with wrong order
	encoded, err := prefixer.SerializeMulti(UserPost,
		EntityUUID{User, userUUID},
		EntityUUID{Post, postUUID},
	)
	assert.NoError(t, err)

	var parsedUser, parsedPost uuid.UUID
	err = prefixer.DeserializeMulti(UserPost, encoded,
		EntityUUIDPtr{Post, &parsedPost},
		EntityUUIDPtr{User, &parsedUser},
	)
	assert.ErrorIs(t, err, ErrEntityOrderMismatch)
}

func TestMultiEntityMismatch(t *testing.T) {
	userUUID := uuid.MustParse("0195e37b-f93f-7518-a9ac-a2be68463c7e")
	postUUID := uuid.MustParse("0195e37b-f93f-7518-a9ac-a2be68463c7f")

	// Serialize as UserPost, try to deserialize as UserPostComment
	encoded, err := prefixer.SerializeMulti(UserPost,
		EntityUUID{User, userUUID},
		EntityUUID{Post, postUUID},
	)
	assert.NoError(t, err)

	var u1, u2, u3 uuid.UUID
	err = prefixer.DeserializeMulti(UserPostComment, encoded,
		EntityUUIDPtr{User, &u1},
		EntityUUIDPtr{Post, &u2},
		EntityUUIDPtr{Comment, &u3},
	)
	assert.ErrorIs(t, err, ErrEntityMismatch)
}

func TestMultiNotMultiEntity(t *testing.T) {
	userUUID := uuid.MustParse("0195e37b-f93f-7518-a9ac-a2be68463c7e")

	// Use a regular (non-multi) entity with SerializeMulti
	_, err := prefixer.SerializeMulti(User,
		EntityUUID{User, userUUID},
	)
	assert.ErrorIs(t, err, ErrNotMultiEntity)

	// Use a regular entity with DeserializeMulti
	encoded := prefixer.Serialize(User, userUUID)
	var parsed uuid.UUID
	err = prefixer.DeserializeMulti(User, encoded,
		EntityUUIDPtr{User, &parsed},
	)
	assert.ErrorIs(t, err, ErrNotMultiEntity)
}

func TestMultiDeserializeErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError error
	}{
		{
			name:          "bad format no separator",
			input:         "upAZXje_k_dRiprKK-aEY8fg",
			expectedError: ErrInvalidPrefixedUUIDFormat,
		},
		{
			name:          "unknown prefix",
			input:         "zzz.AZXje_k_dRiprKK-aEY8fg",
			expectedError: ErrUnknownPrefix,
		},
		{
			name:          "bad base64",
			input:         "up.invalid-base64!",
			expectedError: ErrInvalidUUIDBadBase64,
		},
		{
			name:          "wrong byte length",
			input:         "up.AZXje_k_dRiprKK-aEY8fg",
			expectedError: ErrInvalidUUIDFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u1, u2 uuid.UUID
			err := prefixer.DeserializeMulti(UserPost, tt.input,
				EntityUUIDPtr{User, &u1},
				EntityUUIDPtr{Post, &u2},
			)
			assert.Error(t, err)
			assert.ErrorIs(t, err, tt.expectedError)
		})
	}
}

func TestMultiWithCustomSeparator(t *testing.T) {
	customRegistry, err := NewRegistry([]PrefixInfo{
		{User, "user"},
		{Post, "post"},
	})
	assert.NoError(t, err)
	customRegistry, err = customRegistry.WithSeparator("~")
	assert.NoError(t, err)

	err = customRegistry.AddMultiPrefix(MultiPrefixInfo{UserPost, "up", []Entity{User, Post}})
	assert.NoError(t, err)

	userUUID := uuid.MustParse("0195e37b-f93f-7518-a9ac-a2be68463c7e")
	postUUID := uuid.MustParse("0195e37b-f93f-7518-a9ac-a2be68463c7f")

	encoded, err := customRegistry.SerializeMulti(UserPost,
		EntityUUID{User, userUUID},
		EntityUUID{Post, postUUID},
	)
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(encoded, "up~"))

	var parsedUser, parsedPost uuid.UUID
	err = customRegistry.DeserializeMulti(UserPost, encoded,
		EntityUUIDPtr{User, &parsedUser},
		EntityUUIDPtr{Post, &parsedPost},
	)
	assert.NoError(t, err)
	assert.Equal(t, userUUID, parsedUser)
	assert.Equal(t, postUUID, parsedPost)
}

func TestAddMultiPrefixValidation(t *testing.T) {
	t.Run("null entity", func(t *testing.T) {
		registry, err := NewRegistry([]PrefixInfo{{User, "user"}, {Post, "post"}})
		assert.NoError(t, err)
		err = registry.AddMultiPrefix(MultiPrefixInfo{NullEntity, "up", []Entity{User, Post}})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "NullEntity")
	})

	t.Run("bad prefix", func(t *testing.T) {
		registry, err := NewRegistry([]PrefixInfo{{User, "user"}, {Post, "post"}})
		assert.NoError(t, err)
		err = registry.AddMultiPrefix(MultiPrefixInfo{UserPost, "UP!", []Entity{User, Post}})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "prefix must be in lowercase")
	})

	t.Run("unregistered component", func(t *testing.T) {
		registry, err := NewRegistry([]PrefixInfo{{User, "user"}})
		assert.NoError(t, err)
		err = registry.AddMultiPrefix(MultiPrefixInfo{UserPost, "up", []Entity{User, Post}})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not registered")
	})

	t.Run("duplicate entity", func(t *testing.T) {
		registry, err := NewRegistry([]PrefixInfo{{User, "user"}, {Post, "post"}})
		assert.NoError(t, err)
		err = registry.AddMultiPrefix(MultiPrefixInfo{User, "up", []Entity{User, Post}})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("duplicate prefix", func(t *testing.T) {
		registry, err := NewRegistry([]PrefixInfo{{User, "user"}, {Post, "post"}})
		assert.NoError(t, err)
		err = registry.AddMultiPrefix(MultiPrefixInfo{UserPost, "user", []Entity{User, Post}})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("fewer than 2 entities", func(t *testing.T) {
		registry, err := NewRegistry([]PrefixInfo{{User, "user"}, {Post, "post"}})
		assert.NoError(t, err)
		err = registry.AddMultiPrefix(MultiPrefixInfo{UserPost, "up", []Entity{User}})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least 2")
	})
}
