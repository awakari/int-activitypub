package handler

import (
	"fmt"
	vocab "github.com/go-ap/activitypub"
	"github.com/stretchr/testify/assert"
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
	txt := marshalJsonAndFixContext(a)
	assert.Equal(
		t,
		`{"id":"https://host.social/actor","type":"Person","name":"awakari","summary":"Awakari ActivityPub Bot","attachment":{"id":"https://awakari.com","url":"https://awakari.com"},"@context":["https://www.w3.org/ns/activitystreams","https://w3id.org/security/v1"],"icon":{"type":"Image","mediaType":"image/png","url":"https://awakari.com/logo-color-64.png"},"image":{"type":"Image","mediaType":"image/svg+xml","url":"https://awakari.com/logo-color.svg"},"url":"https://awakari.com","inbox":"https://host.social/inbox","outbox":"https://host.social/outbox","following":"https://host.social/following","followers":"https://host.social/followers","preferredUsername":"awakari","endpoints":{"sharedInbox":"https://host.social/inbox"},"publicKey":{"id":"https://host.social/actor#main-key","owner":"https://host.social/actor"}}`,
		txt,
	)
}
