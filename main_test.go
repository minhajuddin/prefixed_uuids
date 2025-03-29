package prefixed_uuids

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	User      Entity = 1
	Post      Entity = 2
	Comment   Entity = 3
	Other     Entity = 4
	UserV2    Entity = 5
	UserV3    Entity = 6
	SessionID Entity = 7
)

var prefixer = NewRegistry([]PrefixInfo{
	{SessionID, "sid"},
	{User, "user"},
	{UserV2, "user_v2"},
	{UserV3, "user_v3"},
	{Post, "post"},
	{Comment, "comment"},
	{Other, "other"},
})

func TestPrefixes(t *testing.T) {
	uuid, err := uuid.Parse("0195e37b-f93f-7518-a9ac-a2be68463c7e")
	assert.NoError(t, err)
	assert.Equal(t, "user.AZXje_k_dRiprKK-aEY8fg", prefixer.Serialize(User, uuid))
	assert.Equal(t, "user_v2.AZXje_k_dRiprKK-aEY8fg", prefixer.Serialize(UserV2, uuid))

	uuid, err = prefixer.Deserialize(User, "user.AZXje_k_dRiprKK-aEY8fg")
	assert.NoError(t, err)
	assert.Equal(t, uuid.String(), "0195e37b-f93f-7518-a9ac-a2be68463c7e")

	// Shorter than hex uuid and urlsafe base64
	assert.Equal(t, len(prefixer.Serialize(User, uuid)), 27) // user.AZXje_k_dRiprKK-aEY8fg
	assert.Equal(t, len(uuid.String()), 36)                  // 0195e37b-f93f-7518-a9ac-a2be68463c7e

	e, uuid, err := prefixer.DeserializeWithEntity("post.AZXje_k_dRiprKK-aEY8fg")
	assert.NoError(t, err)
	assert.Equal(t, e, Post)
	assert.Equal(t, uuid.String(), "0195e37b-f93f-7518-a9ac-a2be68463c7e")

	assert.Equal(t, prefixer.Serialize(SessionID, uuid), "sid.AZXje_k_dRiprKK-aEY8fg")
}
