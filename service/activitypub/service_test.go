package activitypub

import (
	"context"
	"fmt"
	vocab "github.com/go-ap/activitypub"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
)

func TestService_ResolveActor(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip()
	}
	svc := NewService(http.DefaultClient, "activitypub.awakari.com", []byte{}, nil)
	self, err := svc.ResolveActorLink(context.TODO(), "mastodon.social", "akurilov")
	assert.Equal(t, "https://mastodon.social/users/akurilov", self.String())
	assert.Nil(t, err)
}

func TestService_FetchActor(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip()
	}
	svc := NewService(http.DefaultClient, "activitypub.awakari.com", []byte{}, nil)
	actor, _, err := svc.FetchActor(context.TODO(), "https://mastodon.social/users/akurilov", "https://activitypub.awakari.com/actor#main-key")
	assert.Equal(t, "https://mastodon.social/users/akurilov/inbox", actor.Inbox.GetLink().String())
	assert.Nil(t, err)
}

func TestService_RequestFollow(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping test in CI environment")
	}
	privKey := []byte(`TODO: put private key pem here to test`)
	svc := NewService(http.DefaultClient, "activitypub.awakari.com", privKey, nil)
	err := svc.SendActivity(
		context.TODO(),
		vocab.Activity{
			Type:    vocab.FollowType,
			Context: vocab.IRI("https://www.w3.org/ns/activitystreams"),
			Actor:   vocab.IRI(fmt.Sprintf("https://%s/actor", "activitypub.awakari.com")),
			Object:  vocab.IRI("https://mastodon.social/users/akurilov"),
		},
		"https://mastodon.social/users/akurilov/inbox",
		"foo.bar#main.key",
	)
	assert.Nil(t, err)
}
