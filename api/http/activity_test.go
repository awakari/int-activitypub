package http

import (
	"encoding/json"
	"fmt"
	vocab "github.com/go-ap/activitypub"
	"github.com/stretchr/testify/assert"
	"testing"
)

const host = "host.social"

func Test_FixContext(t *testing.T) {
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
	m := FixContext(a)
	data, err := json.Marshal(m)
	assert.Nil(t, err)
	assert.Equal(
		t,
		"{\"@context\":[\"https://www.w3.org/ns/activitystreams\",\"https://w3id.org/security/v1\",{\"Curve25519Key\":\"toot:Curve25519Key\",\"Device\":\"toot:Device\",\"Ed25519Key\":\"toot:Ed25519Key\",\"Ed25519Signature\":\"toot:Ed25519Signature\",\"EncryptedMessage\":\"toot:EncryptedMessage\",\"PropertyValue\":\"schema:PropertyValue\",\"alsoKnownAs\":{\"@id\":\"as:alsoKnownAs\",\"@type\":\"@id\"},\"cipherText\":\"toot:cipherText\",\"claim\":{\"@id\":\"toot:claim\",\"@type\":\"@id\"},\"deviceId\":\"toot:deviceId\",\"devices\":{\"@id\":\"toot:devices\",\"@type\":\"@id\"},\"discoverable\":\"toot:discoverable\",\"featured\":{\"@id\":\"toot:featured\",\"@type\":\"@id\"},\"featuredTags\":{\"@id\":\"toot:featuredTags\",\"@type\":\"@id\"},\"fingerprintKey\":{\"@id\":\"toot:fingerprintKey\",\"@type\":\"@id\"},\"focalPoint\":{\"@container\":\"@list\",\"@id\":\"toot:focalPoint\"},\"identityKey\":{\"@id\":\"toot:identityKey\",\"@type\":\"@id\"},\"indexable\":\"toot:indexable\",\"manuallyApprovesFollowers\":\"as:manuallyApprovesFollowers\",\"memorial\":\"toot:memorial\",\"messageFranking\":\"toot:messageFranking\",\"messageType\":\"toot:messageType\",\"movedTo\":{\"@id\":\"as:movedTo\",\"@type\":\"@id\"},\"publicKeyBase64\":\"toot:publicKeyBase64\",\"schema\":\"http://schema.org#\",\"suspended\":\"toot:suspended\",\"toot\":\"http://joinmastodon.org/ns#\",\"value\":\"schema:value\"}],\"attachment\":{\"id\":\"https://awakari.com\",\"url\":\"https://awakari.com\"},\"endpoints\":{\"sharedInbox\":\"https://host.social/inbox\"},\"followers\":\"https://host.social/followers\",\"following\":\"https://host.social/following\",\"icon\":{\"mediaType\":\"image/png\",\"type\":\"Image\",\"url\":\"https://awakari.com/logo-color-64.png\"},\"id\":\"https://host.social/actor\",\"image\":{\"mediaType\":\"image/svg+xml\",\"type\":\"Image\",\"url\":\"https://awakari.com/logo-color.svg\"},\"inbox\":\"https://host.social/inbox\",\"name\":\"awakari\",\"outbox\":\"https://host.social/outbox\",\"preferredUsername\":\"awakari\",\"publicKey\":{\"id\":\"https://host.social/actor#main-key\",\"owner\":\"https://host.social/actor\"},\"summary\":\"Awakari ActivityPub Bot\",\"type\":\"Person\",\"url\":\"https://awakari.com\"}",
		string(data),
	)
}
