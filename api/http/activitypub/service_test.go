package activitypub

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
)

func TestService_ResolveActor(t *testing.T) {
	svc := NewService(http.DefaultClient, "activitypub.awakari.com", []byte{})
	self, err := svc.ResolveActor(context.TODO(), "mastodon.social", "akurilov")
	assert.Equal(t, "https://mastodon.social/users/akurilov", self.String())
	assert.Nil(t, err)
}

func TestService_ResolveInbox(t *testing.T) {
	svc := NewService(http.DefaultClient, "activitypub.awakari.com", []byte{})
	inbox, err := svc.ResolveInbox(context.TODO(), "https://mastodon.social/users/akurilov")
	assert.Equal(t, "https://mastodon.social/users/akurilov/inbox", inbox.String())
	assert.Nil(t, err)
}

func TestService_RequestFollow(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping test in CI environment")
	}
	privKey := []byte(`TODO: put private key pem here to test`)
	svc := NewService(http.DefaultClient, "activitypub.awakari.com", privKey)
	err := svc.RequestFollow(
		context.TODO(),
		"mastodon.social",
		"https://mastodon.social/users/akurilov",
		"https://mastodon.social/users/akurilov/inbox",
	)
	assert.Nil(t, err)
}
