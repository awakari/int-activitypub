package converter

import (
	"context"
	"encoding/json"
	"github.com/awakari/int-activitypub/util"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	vocab "github.com/go-ap/activitypub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log/slog"
	"testing"
	"time"
)

func TestService_ConvertActivityToEvent(t *testing.T) {
	svc := NewService("foo", "urlBase", vocab.ServiceType)
	svc = NewLogging(svc, slog.Default())
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
			out: &pb.CloudEvent{
				SpecVersion: "1.0",
				Type:        "foo",
				Source:      "https://mastodon.social/users/johndoe",
				Attributes: map[string]*pb.CloudEventAttributeValue{
					"action": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "Create",
						},
					},
					"attachmenturl": {
						Attr: &pb.CloudEventAttributeValue_CeUri{
							CeUri: "https://files.mastodon.social/media_attachments/files/111/941/781/316/804/883/original/86df8fb1c70309c7.png",
						},
					},
					"attachmenttype": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "image/png",
						},
					},
					"cc": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "https://mastodon.social/users/akurilov/followers",
						},
					},
					"to": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "https://www.w3.org/ns/activitystreams#Public",
						},
					},
					"object": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "Note",
						},
					},
					"objecturl": {
						Attr: &pb.CloudEventAttributeValue_CeUri{
							CeUri: "https://mastodon.social/users/akurilov/statuses/111941782784824099",
						},
					},
					"subject": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "John Doe",
						},
					},
					"time": {
						Attr: &pb.CloudEventAttributeValue_CeTimestamp{
							CeTimestamp: &timestamppb.Timestamp{
								Seconds: 1708096050,
							},
						},
					},
				},
				Data: &pb.CloudEvent_TextData{
					TextData: "<p>image test</p>",
				},
			},
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
  "to": [
	"https://rhiaro.co.uk/followers/",
	"https://www.w3.org/ns/activitystreams#Public"
  ],
  "cc": "https://e14n.com/evan"
}`,
			out: &pb.CloudEvent{
				SpecVersion: "1.0",
				Type:        "foo",
				Source:      "https://mastodon.social/users/johndoe",
				Attributes: map[string]*pb.CloudEventAttributeValue{
					"subject": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "John Doe",
						},
					},
					"cc": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "https://e14n.com/evan",
						},
					},
					"to": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "https://rhiaro.co.uk/followers/ https://www.w3.org/ns/activitystreams#Public",
						},
					},
					"object": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "Article",
						},
					},
					"objecturl": {
						Attr: &pb.CloudEventAttributeValue_CeUri{
							CeUri: "https://rhiaro.co.uk/2016/05/minimal-activitypub",
						},
					},
					"time": {
						Attr: &pb.CloudEventAttributeValue_CeTimestamp{
							CeTimestamp: &timestamppb.Timestamp{
								Seconds: -62135596800,
							},
						},
					},
				},
				Data: &pb.CloudEvent_TextData{
					TextData: "Today I finished morph, a client for posting ActivityStreams2...",
				},
			},
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
			out: &pb.CloudEvent{
				SpecVersion: "1.0",
				Type:        "foo",
				Attributes: map[string]*pb.CloudEventAttributeValue{
					"action": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "Create",
						},
					},
					"latitude": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "-73.974740",
						},
					},
					"longitude": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "40.775630",
						},
					},
					"object": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "Note",
						},
					},
					"objecturl": {
						Attr: &pb.CloudEventAttributeValue_CeUri{
							CeUri: "https://location.edent.tel/9bc18f6eb339ec475c9bcfe69acf21fb",
						},
					},
					"time": {
						Attr: &pb.CloudEventAttributeValue_CeTimestamp{
							CeTimestamp: &timestamppb.Timestamp{
								Seconds: 1708096050,
							},
						},
					},
					"to": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "https://www.w3.org/ns/activitystreams#Public",
						},
					},
				},
				Data: &pb.CloudEvent_TextData{
					TextData: "I just checked-in to <a href=\"https://www.openstreetmap.org/way/958999496\">John Lennon's Imagine Mosaic</a>.",
				},
			},
		},
		"activitypub spec example 14": {
			in: `
{
  "@context": ["https://www.w3.org/ns/activitystreams",
               {"@language": "en"}],
  "type": "Like",
  "actor": "https://dustycloud.org/chris/",
  "summary": "Chris liked 'Minimal ActivityPub update client'",
  "object": "https://rhiaro.co.uk/2016/05/minimal-activitypub",
  "to": ["https://rhiaro.co.uk/#amy",
         "https://dustycloud.org/followers",
         "https://rhiaro.co.uk/followers/",
		 "https://www.w3.org/ns/activitystreams#Public"],
  "cc": "https://e14n.com/evan"
}`,
			out: &pb.CloudEvent{
				SpecVersion: "1.0",
				Type:        "foo",
				Attributes: map[string]*pb.CloudEventAttributeValue{
					"action": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "Like",
						},
					},
					"cc": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "https://e14n.com/evan",
						},
					},
					"summary": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "Chris liked 'Minimal ActivityPub update client'",
						},
					},
					"object": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "Like",
						},
					},
					"objecturl": {
						Attr: &pb.CloudEventAttributeValue_CeUri{
							CeUri: "https://rhiaro.co.uk/2016/05/minimal-activitypub",
						},
					},
					"time": {
						Attr: &pb.CloudEventAttributeValue_CeTimestamp{
							CeTimestamp: &timestamppb.Timestamp{
								Seconds: -62135596800,
							},
						},
					},
					"to": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "https://rhiaro.co.uk/#amy https://dustycloud.org/followers https://rhiaro.co.uk/followers/ https://www.w3.org/ns/activitystreams#Public",
						},
					},
				},
				Data: &pb.CloudEvent_TextData{
					TextData: "Chris liked 'Minimal ActivityPub update client'",
				},
			},
		},
		"activitystreams spec example 5": {
			in: `
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Martin added an article to his blog",
  "type": "Add",
  "published": "2015-02-10T15:04:55Z",
  "actor": {
   "type": "Person",
   "id": "http://www.test.example/martin",
   "name": "Martin Smith",
   "url": "http://example.org/martin",
   "image": {
     "type": "Link",
     "href": "http://example.org/martin/image.jpg",
     "mediaType": "image/jpeg"
   }
  },
  "object" : {
   "id": "http://www.test.example/blog/abc123/xyz",
   "type": "Article",
   "url": "http://example.org/blog/2011/02/entry",
   "name": "Why I love Activity Streams"
  },
  "target" : {
   "id": "http://example.org/blog/",
   "type": "OrderedCollection",
   "name": "Martin's Blog"
  },
  "cc": "https://www.w3.org/ns/activitystreams#Public"
}`,
			out: &pb.CloudEvent{
				SpecVersion: "1.0",
				Type:        "foo",
				Attributes: map[string]*pb.CloudEventAttributeValue{
					"action": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "Add",
						},
					},
					"cc": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "https://www.w3.org/ns/activitystreams#Public",
						},
					},
					"summary": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "Martin added an article to his blog",
						},
					},
					"object": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "Article",
						},
					},
					"objecturl": {
						Attr: &pb.CloudEventAttributeValue_CeUri{
							CeUri: "http://www.test.example/blog/abc123/xyz",
						},
					},
					"time": {
						Attr: &pb.CloudEventAttributeValue_CeTimestamp{
							CeTimestamp: &timestamppb.Timestamp{
								Seconds: 1423580695,
							},
						},
					},
				},
				Data: &pb.CloudEvent_TextData{
					TextData: "Martin added an article to his blog",
				},
			},
		},
		"nobot": {
			actor: vocab.Actor{
				ID: "https://mastodon.social/users/akurilov",
			},
			in: `{
  "@context": [
    "https://www.w3.org/ns/activitystreams",
    {
      "ostatus": "http://ostatus.org#",
      "atomUri": "ostatus:atomUri",
      "inReplyToAtomUri": "ostatus:inReplyToAtomUri",
      "conversation": "ostatus:conversation",
      "sensitive": "as:sensitive",
      "toot": "http://joinmastodon.org/ns#",
      "votersCount": "toot:votersCount",
      "Hashtag": "as:Hashtag"
    }
  ],
  "id": "https://mastodon.social/users/akurilov/statuses/112614067761000729",
  "type": "Note",
  "summary": null,
  "inReplyTo": null,
  "published": "2024-06-14T08:38:25Z",
  "url": "https://mastodon.social/@akurilov/112614067761000729",
  "attributedTo": "https://mastodon.social/users/akurilov",
  "to": [
    "https://www.w3.org/ns/activitystreams#Public"
  ],
  "cc": [
    "https://mastodon.social/users/akurilov/followers"
  ],
  "sensitive": false,
  "atomUri": "https://mastodon.social/users/akurilov/statuses/112614067761000729",
  "inReplyToAtomUri": null,
  "conversation": "tag:mastodon.social,2024-06-14:objectId=729942125:objectType=Conversation",
  "content": "\u003cp\u003etest \u003ca href=\"https://mastodon.social/tags/nobot\" class=\"mention hashtag\" rel=\"tag\"\u003e#\u003cspan\u003enobot\u003c/span\u003e\u003c/a\u003e\u003c/p\u003e",
  "contentMap": {
    "en": "\u003cp\u003etest \u003ca href=\"https://mastodon.social/tags/nobot\" class=\"mention hashtag\" rel=\"tag\"\u003e#\u003cspan\u003enobot\u003c/span\u003e\u003c/a\u003e\u003c/p\u003e"
  },
  "attachment": [],
  "tag": [
    {
      "type": "Hashtag",
      "href": "https://mastodon.social/tags/nobot",
      "name": "#nobot"
    }
  ],
  "replies": {
    "id": "https://mastodon.social/users/akurilov/statuses/112614067761000729/replies",
    "type": "Collection",
    "first": {
      "type": "CollectionPage",
      "next": "https://mastodon.social/users/akurilov/statuses/112614067761000729/replies?only_other_accounts=true\u0026page=true",
      "partOf": "https://mastodon.social/users/akurilov/statuses/112614067761000729/replies",
      "items": []
    }
  }
}`,
			out: &pb.CloudEvent{
				Id:          "cce71487-9c06-4316-9a36-da0d8654ca0e",
				Source:      "https://mastodon.social/users/akurilov",
				SpecVersion: "1.0",
				Type:        "foo",
				Attributes: map[string]*pb.CloudEventAttributeValue{
					"cc": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "https://mastodon.social/users/akurilov/followers",
						},
					},
					"object": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "Note",
						},
					},
					"objecturl": {
						Attr: &pb.CloudEventAttributeValue_CeUri{
							CeUri: "https://mastodon.social/users/akurilov/statuses/112614067761000729",
						},
					},
					"time": {
						Attr: &pb.CloudEventAttributeValue_CeTimestamp{
							CeTimestamp: timestamppb.New(time.Date(2024, 6, 14, 8, 38, 25, 0, time.UTC)),
						},
					},
					"to": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "https://www.w3.org/ns/activitystreams#Public",
						},
					},
				},
				Data: &pb.CloudEvent_TextData{
					TextData: "\u003cp\u003etest \u003ca href=\"https://mastodon.social/tags/nobot\" class=\"mention hashtag\" rel=\"tag\"\u003e#\u003cspan\u003enobot\u003c/span\u003e\u003c/a\u003e\u003c/p\u003e",
				},
			},
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			var activity vocab.Activity
			err := json.Unmarshal([]byte(c.in), &activity)
			require.Nil(t, err)
			evt, err := svc.ConvertActivityToEvent(context.TODO(), c.actor, activity, util.ActivityTags{})
			if c.out == nil {
				assert.Nil(t, evt)
			} else {
				assert.Equal(t, c.out.Type, evt.Type)
				assert.Equal(t, c.out.Source, evt.Source)
				assert.Equal(t, c.out.SpecVersion, evt.SpecVersion)
				assert.Equal(t, c.out.Data, evt.Data)
				assert.Equal(t, c.out.Attributes, evt.Attributes)
			}
			assert.ErrorIs(t, err, c.err)
		})
	}

}
