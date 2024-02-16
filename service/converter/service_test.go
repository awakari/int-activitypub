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
		"1": {
			actor: vocab.Actor{
				Name: vocab.DefaultNaturalLanguageValue("John Doe"),
			},
			in: `{
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
