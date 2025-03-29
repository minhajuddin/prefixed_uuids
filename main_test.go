package prefixed_uuids

import (
	"testing"

	"github.com/google/uuid"
	"gotest.tools/assert"
)

const (
	User    Entity = 1
	Post    Entity = 2
	Comment Entity = 3
)

var prefixer Registry = NewRegistry(
	NewPrefixInfo(User, "user"),
	NewPrefixInfo(Post, "post"),
	NewPrefixInfo(Comment, "comment"),
)

func TestPrefixes(t *testing.T) {
	uuid, err := uuid.Parse("0195e37b-f93f-7518-a9ac-a2be68463c7e")
	assert.NilError(t, err)
	assert.Equal(t, "user.AZXje_k_dRiprKK-aEY8fg", prefixer.Serialize(User, uuid))

	uuid, err = prefixer.Deserialize(User, "user.AZXje_k_dRiprKK-aEY8fg")
	assert.NilError(t, err)
	assert.Equal(t, uuid.String(), "0195e37b-f93f-7518-a9ac-a2be68463c7e")

	e, uuid, err := prefixer.DeserializeWithEntity("post.AZXje_k_dRiprKK-aEY8fg")
	assert.NilError(t, err)
	assert.Equal(t, e, Post)
	assert.Equal(t, uuid.String(), "0195e37b-f93f-7518-a9ac-a2be68463c7e")
}
