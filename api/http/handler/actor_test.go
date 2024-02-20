package handler

import (
	"encoding/json"
	"fmt"
	vocab "github.com/go-ap/activitypub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const host = "host.social"

func TestActor_marshalJsonAndFixContext(t *testing.T) {
	a := vocab.Actor{
		ID: vocab.ID(fmt.Sprintf("https://%s/actor", host)),
		Context: vocab.ItemCollection{
			vocab.IRI("https://www.w3.org/ns/activitystreams"),
			vocab.IRI("https://w3id.org/security/v1"),
		},
		Type: vocab.PersonType,
		Name: vocab.DefaultNaturalLanguageValue("awakari"),
		Icon: vocab.Image{
			MediaType: "image/png",
			Type:      vocab.ImageType,
			URL:       vocab.IRI("https://awakari.com/logo-color-64.png"),
		},
		Image: vocab.Image{
			MediaType: "image/svg+xml",
			Type:      vocab.ImageType,
			URL:       vocab.IRI("https://awakari.com/logo-color.svg"),
		},
		Summary:           vocab.DefaultNaturalLanguageValue("Awakari ActivityPub Bot"),
		URL:               vocab.IRI("https://awakari.com"),
		Inbox:             vocab.IRI(fmt.Sprintf("https://%s/inbox", host)),
		Outbox:            vocab.IRI(fmt.Sprintf("https://%s/outbox", host)),
		Following:         vocab.IRI(fmt.Sprintf("https://%s/following", host)),
		Followers:         vocab.IRI(fmt.Sprintf("https://%s/followers", host)),
		PreferredUsername: vocab.DefaultNaturalLanguageValue("awakari"),
		Endpoints: &vocab.Endpoints{
			SharedInbox: vocab.IRI(fmt.Sprintf("https://%s/inbox", "host.social")),
		},
		PublicKey: vocab.PublicKey{
			ID:           vocab.ID(fmt.Sprintf("https://%s/actor#main-key", host)),
			Owner:        vocab.IRI(fmt.Sprintf("https://%s/actor", host)),
			PublicKeyPem: "",
		},
		Attachment: vocab.ItemCollection{
			vocab.Page{
				ID:  vocab.ID("https://awakari.com"),
				URL: vocab.IRI("https://awakari.com"),
			},
		},
	}
	data, err := json.Marshal(a)
	require.Nil(t, err)
	raw := map[string]any{}
	err = json.Unmarshal(data, &raw)
	ctx, ctxFound := raw["context"]
	require.Nil(t, err)
	assert.NotNil(t, ctx)
	assert.True(t, ctxFound)
	delete(raw, "context")
	raw["@context"] = ctx
	data, err = json.Marshal(raw)
	assert.Nil(t, err)
	assert.Equal(
		t,
		`{"@context":["https://www.w3.org/ns/activitystreams","https://w3id.org/security/v1"],"attachment":{"id":"https://awakari.com","url":"https://awakari.com"},"endpoints":{"sharedInbox":"https://host.social/inbox"},"followers":"https://host.social/followers","following":"https://host.social/following","icon":{"mediaType":"image/png","type":"Image","url":"https://awakari.com/logo-color-64.png"},"id":"https://host.social/actor","image":{"mediaType":"image/svg+xml","type":"Image","url":"https://awakari.com/logo-color.svg"},"inbox":"https://host.social/inbox","name":"awakari","outbox":"https://host.social/outbox","preferredUsername":"awakari","publicKey":{"id":"https://host.social/actor#main-key","owner":"https://host.social/actor"},"summary":"Awakari ActivityPub Bot","type":"Person","url":"https://awakari.com"}`,
		string(data),
	)
}
