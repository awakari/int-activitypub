package converter

import (
	"context"
	"encoding/json"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	vocab "github.com/go-ap/activitypub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestService_Convert(t *testing.T) {
	svc := NewService()
	cases := map[string]struct {
		actor vocab.Actor
		in    string
		out   *pb.CloudEvent
		err   error
	}{
		"mastodon post with image attached": {
			actor: vocab.Actor{
				ID:   "https://mastodon.social/users/johndoe",
				Name: vocab.DefaultNaturalLanguageValue("John Doe"),
			},
			in: `
{
  "id": "https://mastodon.social/users/akurilov/statuses/111941782784824099/activity",
  "type": "Create",
  "to": [
	"https://www.w3.org/ns/activitystreams#Public"
  ],
  "cc": [
	"https://mastodon.social/users/akurilov/followers"
  ],
  "published": "2024-02-16T15:07:30Z",
  "actor": "https://mastodon.social/users/akurilov",
  "object": {
	"id": "https://mastodon.social/users/akurilov/statuses/111941782784824099",
	"type": "Note",
	"content": "\u003cp\u003eimage test\u003c/p\u003e",
	"attachment": {
	  "type": "Document",
	  "mediaType": "image/png",
	  "url": "https://files.mastodon.social/media_attachments/files/111/941/781/316/804/883/original/86df8fb1c70309c7.png"
	},
	"attributedTo": "https://mastodon.social/users/akurilov",
	"replies": {
	  "id": "https://mastodon.social/users/akurilov/statuses/111941782784824099/replies",
	  "type": "Collection",
	  "first": {
		"type": "CollectionPage",
		"partOf": "https://mastodon.social/users/akurilov/statuses/111941782784824099/replies",
		"next": "https://mastodon.social/users/akurilov/statuses/111941782784824099/replies?only_other_accounts=true\u0026page=true",
		"totalItems": 0
	  },
	  "totalItems": 0
	},
	"url": "https://mastodon.social/@akurilov/111941782784824099",
	"to": [
	  "https://www.w3.org/ns/activitystreams#Public"
	],
	"cc": [
	  "https://mastodon.social/users/akurilov/followers"
	],
	"published": "2024-02-16T15:07:30Z"
  }
}`,
		},
		"object instead of activity": {
			actor: vocab.Actor{
				ID:   "https://mastodon.social/users/johndoe",
				Name: vocab.DefaultNaturalLanguageValue("John Doe"),
			},
			in: `
{
  "@context": ["https://www.w3.org/ns/activitystreams",
               {"@language": "en-GB"}],
  "id": "https://rhiaro.co.uk/2016/05/minimal-activitypub",
  "type": "Article",
  "name": "Minimal ActivityPub update client",
  "content": "Today I finished morph, a client for posting ActivityStreams2...",
  "attributedTo": "https://rhiaro.co.uk/#amy",
  "to": "https://rhiaro.co.uk/followers/",
  "cc": "https://e14n.com/evan"
}`,
		},
		"location": {
			in: `
{
  "id": "https://mastodon.social/users/akurilov/statuses/111941782784824099/activity",
  "type": "Create",
  "published": "2024-02-16T15:07:30Z",
  "actor": "https://mastodon.social/users/akurilov",
  "object": {
	  "@context": "https://www.w3.org/ns/activitystreams",
	  "id": "https://location.edent.tel/9bc18f6eb339ec475c9bcfe69acf21fb",
	  "type": "Note",
	  "published": "2024-01-28T12:13:38+00:00",
	  "attributedTo": "https://location.edent.tel/edent_location",
	  "content": "I just checked-in to <a href=\"https://www.openstreetmap.org/way/958999496\">John Lennon's Imagine Mosaic</a>.",
	  "to": [
		"https://www.w3.org/ns/activitystreams#Public"
	  ],
	  "location": {
		"name": "John Lennon's Imagine",
		"type": "Place",
		"longitude": 40.77563,
		"latitude": -73.97474
	  }
  }
}`,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			var activity vocab.Activity
			err := json.Unmarshal([]byte(c.in), &activity)
			require.Nil(t, err)
			evt, err := svc.Convert(context.TODO(), c.actor, activity)
			assert.Equal(t, c.out, evt)
			assert.ErrorIs(t, err, c.err)
		})
	}

}
