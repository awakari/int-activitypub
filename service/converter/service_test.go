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
	svc := NewService("foo", "urlBase", "https://reader/evt", vocab.ServiceType)
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
      "sensitive": "NsAs:sensitive",
      "toot": "http://joinmastodon.org/ns#",
      "votersCount": "toot:votersCount",
      "Hashtag": "NsAs:Hashtag"
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

func TestService_ConvertEventToActivity(t *testing.T) {
	svc := NewService("foo", "https://base", "https://reader/evt", vocab.ServiceType)
	svc = NewLogging(svc, slog.Default())
	ts := time.Date(2024, 7, 27, 1, 32, 21, 0, time.UTC)
	cases := map[string]struct {
		src        *pb.CloudEvent
		interestId string
		follower   *vocab.Actor
		dst        vocab.Activity
		err        error
	}{
		"1": {
			src: &pb.CloudEvent{
				Id:          "2jrVcFeXfGNcExKHLCcrrXBYyLJ",
				SpecVersion: CeSpecVersion,
				Source:      "https://otakukart.com/feed/",
				Type:        "com_awakari_feeds_v1",
				Data: &pb.CloudEvent_TextData{
					TextData: "<div><img width=\"1280\" height=\"720\" src=\"https://otakukart.com/wp-content/uploads/2024/07/The-10-Must-Watch-Futuristic-Anime-That-Every-Fan-Should-See.jpg\" class=\"attachment-medium size-medium wp-post-image\" alt=\"The 10 Must-Watch Futuristic Anime That Every Fan Should See\" decoding=\"async\" srcset=\"https://otakukart.com/wp-content/uploads/2024/07/The-10-Must-Watch-Futuristic-Anime-That-Every-Fan-Should-See.jpg 1280w, https://otakukart.com/wp-content/uploads/2024/07/The-10-Must-Watch-Futuristic-Anime-That-Every-Fan-Should-See-770x433.jpg 770w, https://otakukart.com/wp-content/uploads/2024/07/The-10-Must-Watch-Futuristic-Anime-That-Every-Fan-Should-See-150x84.jpg 150w, https://otakukart.com/wp-content/uploads/2024/07/The-10-Must-Watch-Futuristic-Anime-That-Every-Fan-Should-See-750x422.jpg 750w, https://otakukart.com/wp-content/uploads/2024/07/The-10-Must-Watch-Futuristic-Anime-That-Every-Fan-Should-See-1140x641.jpg 1140w\" sizes=\"(max-width: 1280px) 100vw, 1280px\" /></div><p><img src=\"https://otakukart.com/wp-content/uploads/2024/07/The-10-Must-Watch-Futuristic-Anime-That-Every-Fan-Should-See.jpg\" style=\"display: block; margin: 1em auto\"></p>\n<p>Anime is known for its wide range of stories, each reflecting the boundless creativity of its creators. Among the various genres, speculative fiction stands out for its focus on futuristic themes and ideas.</p>\n<p>This genre allows for an exploration of what the future might hold, resulting in innovative and impressive stories.</p>\n<p>By imagining different possibilities, speculative fiction in anime provides a platform for imaginative storytelling that stretches the limits of what is possible.</p>\n<p>Futuristic anime often presents two distinct visions of what lies ahead. Optimistic stories depict humanity&#8217;s journey into space and the thriving civilizations that await in the intergalactic era.</p>\n<p>These stories offer a hopeful and prosperous future filled with exploration and advancement.</p>\n<p>In contrast, dystopian tales serve as cautionary stories, highlighting the potential dangers and pitfalls that could emerge if society falters.</p>\n<p>This blend of utopian and dystopian perspectives adds richness and diversity to the genre, engaging viewers with both hope and warning.</p>\n<p>Whether predicting far-off futures or near-term developments, anime set in the future are among the most compelling in the medium.</p>\n<p>These speculative tales frequently top the charts, drawing viewers in with their visionary and thought-provoking storytelling.</p>\n<p>By exploring a wide range of potential futures, these anime entertain while also prompting reflection on the direction humanity might take.</p>\n<p>Through their creative and speculative nature, these stories not only impress fans but also inspire consideration of what our future could become.</p>\n<h2>1) Cowboy Bebop&nbsp;</h2>\n<p>Cowboy Bebop is renowned for its unique and stylish approach to the space Western genre. The series is set in a future where humanity has settled throughout the Solar System and space travel is commonplace.</p>\n<p>It follows a group of bounty hunters as they explore their way through this vast and varied world, taking on different missions along the way.</p>\n<p>Instead of sticking to typical sci-fi or space opera conventions, Cowboy Bebop combines a gritty, neo-noir atmosphere with influences from Westerns and cyberpunk.</p>\n<figure id=\"attachment_1570894\" aria-describedby=\"caption-attachment-1570894\" style=\"width: 1280px\" class=\"wp-caption alignnone\"><img loading=\"lazy\" decoding=\"async\" class=\"size-full wp-image-1570894\" src=\"https://otakukart.com/wp-content/uploads/2024/07/Cowboy-Bebop.jpg\" alt=\"The 10 Must-Watch Futuristic Anime That Every Fan Should See\" width=\"1280\" height=\"720\" srcset=\"https://otakukart.com/wp-content/uploads/2024/07/Cowboy-Bebop.jpg 1280w, https://otakukart.com/wp-content/uploads/2024/07/Cowboy-Bebop-770x433.jpg 770w, https://otakukart.com/wp-content/uploads/2024/07/Cowboy-Bebop-150x84.jpg 150w, https://otakukart.com/wp-content/uploads/2024/07/Cowboy-Bebop-750x422.jpg 750w, https://otakukart.com/wp-content/uploads/2024/07/Cowboy-Bebop-1140x641.jpg 1140w\" sizes=\"(max-width: 1280px) 100vw, 1280px\" /><figcaption id=\"caption-attachment-1570894\" class=\"wp-caption-text\">Cowboy Bebop (Sunrise)</figcaption></figure>\n<p>This blend of genres creates a distinctive and unforgettable vibe that sets the series apart from other space-themed anime, offering viewers a fresh and immersive experience.</p>\n<p>The anime&#8217;s episodic format enables it to cover a wide range of stories. Whether it’s light-hearted capers set in futuristic cities or deep, existential reflections on isolation in the expanse of space, Cowboy Bebop masterfully shifts between genres and tones.</p>\n<p>This versatility contributes to its enduring appeal and establishes it as a timeless and influential classic in anime.</p>\n<h2>2) Dr. Stone&nbsp;</h2>\n<p>Dr. Stone offers a distinctive vision of the future by placing its story in a world where all technology has disappeared.</p>\n<p>Following a strange cosmic event that petrifies every human, society collapses and nature reclaims the planet. As a result, the Earth is left in a primitive state with all human progress erased over the millennia.</p>\n<p>After 3,700 years, the story’s protagonist, Senkuu, a teenage genius, awakens in this new Stone Age. With a fierce determination to revive technological advancements, he sets out to rebuild civilization from scratch.</p>\n<figure id=\"attachment_1570896\" aria-describedby=\"caption-attachment-1570896\" style=\"width: 1280px\" class=\"wp-caption alignnone\"><img loading=\"lazy\" decoding=\"async\" class=\"size-full wp-image-1570896\" src=\"https://otakukart.com/wp-content/uploads/2024/07/Dr.-Stone-.jpg\" alt=\"The 10 Must-Watch Futuristic Anime That Every Fan Should See\" width=\"1280\" height=\"720\" srcset=\"https://otakukart.com/wp-content/uploads/2024/07/Dr.-Stone-.jpg 1280w, https://otakukart.com/wp-content/uploads/2024/07/Dr.-Stone--770x433.jpg 770w, https://otakukart.com/wp-content/uploads/2024/07/Dr.-Stone--150x84.jpg 150w, https://otakukart.com/wp-content/uploads/2024/07/Dr.-Stone--750x422.jpg 750w, https://otakukart.com/wp-content/uploads/2024/07/Dr.-Stone--1140x641.jpg 1140w\" sizes=\"(max-width: 1280px) 100vw, 1280px\" /><figcaption id=\"caption-attachment-1570896\" class=\"wp-caption-text\">Dr. Stone&nbsp;&nbsp;(TMS Entertainment)</figcaption></figure>\n<p>Despite the futuristic backdrop, the series centers on Senkuu’s innovative attempts to recreate modern technology using only the most basic materials available.</p>\n<p>What makes Dr. Stone stand out is its fresh approach to the concept of the future. Rather than depicting a high-tech or dystopian world, it focuses on the process of rediscovering and reinventing technology in a setting that has reverted to its primitive roots.</p>\n<p>The anime skillfully blends the ancient with the modern, offering an interesting story of progress and discovery.</p>\n<h2>3) Psycho-Pass&nbsp;&nbsp;</h2>\n<p>Psycho-Pass is set in a seemingly flawless future where personal freedoms are sacrificed for the illusion of safety and prosperity.</p>\n<p>In this dystopian world, advanced technology constantly monitors citizens to gauge their likelihood of committing crimes.</p>\n<p>While this system creates an outward appearance of order and control, it masks a much darker truth beneath the surface.</p>\n<p>The series centers on Akane Tsunemori, an Inspector tasked with enforcing laws based on these technological assessments.</p>\n<figure id=\"attachment_1570893\" aria-describedby=\"caption-attachment-1570893\" style=\"width: 1280px\" class=\"wp-caption alignnone\"><img loading=\"lazy\" decoding=\"async\" class=\"size-full wp-image-1570893\" src=\"https://otakukart.com/wp-content/uploads/2024/07/Psycho-Pass-.jpg\" alt=\"The 10 Must-Watch Futuristic Anime That Every Fan Should See\" width=\"1280\" height=\"720\" srcset=\"https://otakukart.com/wp-content/uploads/2024/07/Psycho-Pass-.jpg 1280w, https://otakukart.com/wp-content/uploads/2024/07/Psycho-Pass--770x433.jpg 770w, https://otakukart.com/wp-content/uploads/2024/07/Psycho-Pass--150x84.jpg 150w, https://otakukart.com/wp-content/uploads/2024/07/Psycho-Pass--750x422.jpg 750w, https://otakukart.com/wp-content/uploads/2024/07/Psycho-Pass--1140x641.jpg 1140w\" sizes=\"(max-width: 1280px) 100vw, 1280px\" /><figcaption id=\"caption-attachment-1570893\" class=\"wp-caption-text\">Psycho Pass&nbsp;(Production I.G)</figcaption></figure>\n<p>Though Akane and others believe in the fairness of their system, they are oblivious to the fact that the ruling authority, known as the Sibyl System, is merely a network of biocomputers.</p>\n<p>This impersonal system makes detached decisions about who deserves a good life and who should be removed from society.</p>\n<p>The true terror of Psycho-Pass emerges from its depiction of a society that seems perfect but is actually driven by an indifferent and authoritarian regime.</p>\n<p>The anime explores the chilling effects of a world where an emotionless machine dictates every aspect of people&#8217;s lives, prompting viewers to question concepts of freedom, justice, and morality.</p>\n<h2>4) Cyberpunk: Edgerunners&nbsp;</h2>\n<p>Set in the same world as the action RPG video game Cyberpunk 2077, Trigger’s Cyberpunk: Edgerunners unfolds about a year before the game&#8217;s events in the grim and chaotic Night City.</p>\n<p>This futuristic dystopia is plagued by lawlessness, rampant cybernetic modifications, unchecked corporate power, and a blatant disregard for human life, creating a harsh and unforgiving environment.</p>\n<p>In the midst of this turmoil, the protagonist, David Martinez, is drawn into the perilous world of a mercenary.</p>\n<p>His journey is marked by intense violence and adrenaline-charged encounters, leading him down a dark and tragic path.</p>\n<figure id=\"attachment_1570892\" aria-describedby=\"caption-attachment-1570892\" style=\"width: 1280px\" class=\"wp-caption alignnone\"><img loading=\"lazy\" decoding=\"async\" class=\"size-full wp-image-1570892\" src=\"https://otakukart.com/wp-content/uploads/2024/07/Cyberpunk-Edgerunners-.jpg\" alt=\"The 10 Must-Watch Futuristic Anime That Every Fan Should See\" width=\"1280\" height=\"720\" srcset=\"https://otakukart.com/wp-content/uploads/2024/07/Cyberpunk-Edgerunners-.jpg 1280w, https://otakukart.com/wp-content/uploads/2024/07/Cyberpunk-Edgerunners--770x433.jpg 770w, https://otakukart.com/wp-content/uploads/2024/07/Cyberpunk-Edgerunners--150x84.jpg 150w, https://otakukart.com/wp-content/uploads/2024/07/Cyberpunk-Edgerunners--750x422.jpg 750w, https://otakukart.com/wp-content/uploads/2024/07/Cyberpunk-Edgerunners--1140x641.jpg 1140w\" sizes=\"(max-width: 1280px) 100vw, 1280px\" /><figcaption id=\"caption-attachment-1570892\" class=\"wp-caption-text\">Cyberpunk Edgerunners&nbsp;&nbsp;(Trigger)</figcaption></figure>\n<p>David&#8217;s descent into the anarchic underbelly of Night City is both thrilling and heartbreaking, showcasing the brutal reality of living in such a dystopian future.</p>\n<p>Through Trigger&#8217;s dynamic and frenetic artistic direction, Night City is brought to life in a way that is both mesmerizing and terrifying.</p>\n<p>The series captures the haunting beauty of its doomed futuristic worlds, immersing viewers in a world that is as visually stunning as it is bleak.</p>\n<p>Cyberpunk: Edgerunners masterfully blends action and tragedy, creating an interesting story that leaves a lasting impression.</p>\n<h2>5) Trigun&nbsp;</h2>\n<p>Set on the distant planet of No Man&#8217;s Land, also known as Gunsmoke in the 1998 anime, Trigun combines the aesthetics of a space Western with a unique vision of the future.</p>\n<p>The planet&#8217;s harsh desert environment is a result of a failed space colonization mission that occurred 150 years before the series&#8217; events.</p>\n<p>The inhabitants of No Man&#8217;s Land struggle to survive amid remnants of advanced technology and alien life forms, blending with the rugged, Old West-inspired world and lifestyle.</p>\n<p>In this chaotic and lawless world, the central character of Trigun, Vash the Stampede, is a notorious outlaw known for his destructive presence.</p>\n<figure id=\"attachment_1570891\" aria-describedby=\"caption-attachment-1570891\" style=\"width: 1280px\" class=\"wp-caption alignnone\"><img loading=\"lazy\" decoding=\"async\" class=\"size-full wp-image-1570891\" src=\"https://otakukart.com/wp-content/uploads/2024/07/Trigun.jpg\" alt=\"The 10 Must-Watch Futuristic Anime That Every Fan Should See\" width=\"1280\" height=\"720\" srcset=\"https://otakukart.com/wp-content/uploads/2024/07/Trigun.jpg 1280w, https://otakukart.com/wp-content/uploads/2024/07/Trigun-770x433.jpg 770w, https://otakukart.com/wp-content/uploads/2024/07/Trigun-150x84.jpg 150w, https://otakukart.com/wp-content/uploads/2024/07/Trigun-750x422.jpg 750w, https://otakukart.com/wp-content/uploads/2024/07/Trigun-1140x641.jpg 1140w\" sizes=\"(max-width: 1280px) 100vw, 1280px\" /><figcaption id=\"caption-attachment-1570891\" class=\"wp-caption-text\">Trigun&nbsp;(MADHOUSE)</figcaption></figure>\n<p>Despite his fearsome reputation, Vash roams the planet promoting messages of love and peace. His efforts, however, often lead to unintended trouble and significant mayhem wherever he goes.</p>\n<p>Trigun stands out for its innovative mix of futuristic elements with classic Western themes. By integrating advanced technology and alien influences into a setting reminiscent of the Old West, the series creates a distinct and engaging story.</p>\n<p>Vash&#8217;s journey through this desolate and lawless planet provides a compelling exploration of contrasting ideals and the complexities of its setting.</p>\n<h2>6) Ergo Proxy&nbsp;</h2>\n<p>Ergo Proxy is an interesting anime set in a future where ecological disaster has confined humanity to domed cities.</p>\n<p>In these cities, peace is upheld by genetically engineered humans and androids called AutoReivs, which perform various tasks for the citizens.</p>\n<p>This delicate balance is shattered when a virus causes the AutoReivs to gain self-awareness and commit murders.</p>\n<p>Against this intricate backdrop, Ergo Proxy weaves a story that explores deep existential questions. It explores themes such as the search for life&#8217;s purpose and the repercussions of ecological collapse.</p>\n<figure id=\"attachment_1570890\" aria-describedby=\"caption-attachment-1570890\" style=\"width: 1280px\" class=\"wp-caption alignnone\"><img loading=\"lazy\" decoding=\"async\" class=\"size-full wp-image-1570890\" src=\"https://otakukart.com/wp-content/uploads/2024/07/Ergo-Proxy.jpg\" alt=\"The 10 Must-Watch Futuristic Anime That Every Fan Should See\" width=\"1280\" height=\"720\" srcset=\"https://otakukart.com/wp-content/uploads/2024/07/Ergo-Proxy.jpg 1280w, https://otakukart.com/wp-content/uploads/2024/07/Ergo-Proxy-770x433.jpg 770w, https://otakukart.com/wp-content/uploads/2024/07/Ergo-Proxy-150x84.jpg 150w, https://otakukart.com/wp-content/uploads/2024/07/Ergo-Proxy-750x422.jpg 750w, https://otakukart.com/wp-content/uploads/2024/07/Ergo-Proxy-1140x641.jpg 1140w\" sizes=\"(max-width: 1280px) 100vw, 1280px\" /><figcaption id=\"caption-attachment-1570890\" class=\"wp-caption-text\">Ergo Proxy (Manglobe)</figcaption></figure>\n<p>As the story progresses, viewers are drawn into the unfolding mysteries of the domed city of Romdeau and the evolving consciousness of the AutoReivs.</p>\n<p>The series stands out for its ability to provoke thought and reflection. It challenges viewers to interpret its complex messages and themes, making the viewing experience both intellectually stimulating and deeply rewarding.</p>\n<p>By combining a rich sci-fi setting with profound existential inquiries, Ergo Proxy establishes itself as a uniquely engaging anime.</p>\n<h2>7) Heavenly Delusion&nbsp;&nbsp;</h2>\n<p>Heavenly Delusion is an intricate anime set in a world that has been in ruins for over 15 years. The story takes place in a desolate and perilous world, where strange and terrifying creatures roam freely.</p>\n<p>In this unforgiving world, two teenagers, Maru and Kiruko, set out on a journey to find a mythical place known as Heaven, believed to be a sanctuary offering hope and safety amidst the chaos.</p>\n<p>The series also features a contrasting storyline involving a cutting-edge facility that houses children with extraordinary powers.</p>\n<p>This facility is isolated from the devastated world, creating an obvious contrast between the two stories.</p>\n<figure id=\"attachment_1570889\" aria-describedby=\"caption-attachment-1570889\" style=\"width: 1280px\" class=\"wp-caption alignnone\"><img loading=\"lazy\" decoding=\"async\" class=\"size-full wp-image-1570889\" src=\"https://otakukart.com/wp-content/uploads/2024/07/Heavenly-Delusion.jpg\" alt=\"The 10 Must-Watch Futuristic Anime That Every Fan Should See\" width=\"1280\" height=\"720\" srcset=\"https://otakukart.com/wp-content/uploads/2024/07/Heavenly-Delusion.jpg 1280w, https://otakukart.com/wp-content/uploads/2024/07/Heavenly-Delusion-770x433.jpg 770w, https://otakukart.com/wp-content/uploads/2024/07/Heavenly-Delusion-150x84.jpg 150w, https://otakukart.com/wp-content/uploads/2024/07/Heavenly-Delusion-750x422.jpg 750w, https://otakukart.com/wp-content/uploads/2024/07/Heavenly-Delusion-1140x641.jpg 1140w\" sizes=\"(max-width: 1280px) 100vw, 1280px\" /><figcaption id=\"caption-attachment-1570889\" class=\"wp-caption-text\">Heavenly Delusion (Production I.G)</figcaption></figure>\n<p>The anime’s dual plots add layers of mystery as viewers are challenged to uncover how these separate storylines are intertwined.</p>\n<p>What makes Heavenly Delusion particularly interesting is how it interlaces these two seemingly unrelated stories.</p>\n<p>The mysterious connection between the post-apocalyptic wasteland and the advanced facility adds depth to the story.</p>\n<p>By blending intense survival action with thought-provoking mysteries, Heavenly Delusion stands out as a remarkable and engaging modern anime.</p>\n<h2>8) Ghost in the Shell: Stand Alone Complex&nbsp;</h2>\n<p>Ghost in the Shell: Stand Alone Complex is a sci-fi series that presents a future both near and interestingly advanced.</p>\n<p>Set in the fictional Japanese prefecture of Niihama, the anime explores a world where cybernetic enhancements are the norm.</p>\n<p>People sport advanced prosthetics or fully artificial bodies, effectively turning them into cyborgs. In this high-tech future, cybercrime is a major threat, and Major Motoko Kusanagi, along with her team in Section 9, works tirelessly to combat it.</p>\n<figure id=\"attachment_1570888\" aria-describedby=\"caption-attachment-1570888\" style=\"width: 1280px\" class=\"wp-caption alignnone\"><img loading=\"lazy\" decoding=\"async\" class=\"size-full wp-image-1570888\" src=\"https://otakukart.com/wp-content/uploads/2024/07/Ghost-in-the-Shell-Stand-Alone-Complex.jpg\" alt=\"The 10 Must-Watch Futuristic Anime That Every Fan Should See\" width=\"1280\" height=\"720\" srcset=\"https://otakukart.com/wp-content/uploads/2024/07/Ghost-in-the-Shell-Stand-Alone-Complex.jpg 1280w, https://otakukart.com/wp-content/uploads/2024/07/Ghost-in-the-Shell-Stand-Alone-Complex-770x433.jpg 770w, https://otakukart.com/wp-content/uploads/2024/07/Ghost-in-the-Shell-Stand-Alone-Complex-150x84.jpg 150w, https://otakukart.com/wp-content/uploads/2024/07/Ghost-in-the-Shell-Stand-Alone-Complex-750x422.jpg 750w, https://otakukart.com/wp-content/uploads/2024/07/Ghost-in-the-Shell-Stand-Alone-Complex-1140x641.jpg 1140w\" sizes=\"(max-width: 1280px) 100vw, 1280px\" /><figcaption id=\"caption-attachment-1570888\" class=\"wp-caption-text\">Ghost in the Shell Stand Alone Complex (Production I.G)</figcaption></figure>\n<p>The series doesn&#8217;t idealize this world; instead, it portrays a dystopian society where technological advancements come with significant downsides.</p>\n<p>The grim reality of this future setting serves as a backdrop for the characters&#8217; battles against crime.</p>\n<p>Through its complex story, Ghost in the Shell prompts viewers to reflect on themes of identity and consciousness in an era dominated by technology.</p>\n<p>The anime&#8217;s gritty, cyberpunk environment deepens its exploration of these issues, making it a standout in the genre.</p>\n<p>By tackling such profound topics, Ghost in the Shell continues to be a thought-provoking and influential anime.</p>\n<h2>9) Legend of the Galactic Heroes</h2>\n<p>Legend of the Galactic Heroes is a sweeping space opera that centers on the dramatic struggle between two powerful factions: the authoritarian Galactic Empire and the democratic Free Planets Alliance.</p>\n<p>Known for its intricate political stories and profound themes, the anime delves deeply into its characters and the complex interplay of ideologies that drive the conflict.</p>\n<p>Set in a world where advanced space travel makes exploration of numerous terraformed planets possible, the series meticulously constructs its setting.</p>\n<figure id=\"attachment_1570887\" aria-describedby=\"caption-attachment-1570887\" style=\"width: 1280px\" class=\"wp-caption alignnone\"><img loading=\"lazy\" decoding=\"async\" class=\"size-full wp-image-1570887\" src=\"https://otakukart.com/wp-content/uploads/2024/07/Legend-of-the-Galactic-Heroes-.jpg\" alt=\"The 10 Must-Watch Futuristic Anime That Every Fan Should See\" width=\"1280\" height=\"720\" srcset=\"https://otakukart.com/wp-content/uploads/2024/07/Legend-of-the-Galactic-Heroes-.jpg 1280w, https://otakukart.com/wp-content/uploads/2024/07/Legend-of-the-Galactic-Heroes--770x433.jpg 770w, https://otakukart.com/wp-content/uploads/2024/07/Legend-of-the-Galactic-Heroes--150x84.jpg 150w, https://otakukart.com/wp-content/uploads/2024/07/Legend-of-the-Galactic-Heroes--750x422.jpg 750w, https://otakukart.com/wp-content/uploads/2024/07/Legend-of-the-Galactic-Heroes--1140x641.jpg 1140w\" sizes=\"(max-width: 1280px) 100vw, 1280px\" /><figcaption id=\"caption-attachment-1570887\" class=\"wp-caption-text\">Legend of the Galactic Heroes&nbsp;(Kitty Film)</figcaption></figure>\n<p>Legend of the Galactic Heroes includes diverse cultures and unique timekeeping systems to enhance the realism of its futuristic backdrop.</p>\n<p>This thorough approach to world-building creates a richly detailed and believable environment. The anime stands out for its comprehensive and thoughtful world-building.</p>\n<p>By focusing on every detail of its setting, Legend of the Galactic Heroes offers an immersive and engaging experience.</p>\n<p>Its dedication to creating a fully realized sci-fi world helps solidify its place as one of the most expansive stories in anime.</p>\n<h2>10) Planetes&nbsp;</h2>\n<p>Space exploration has always been a fascinating theme in science fiction, embodying humanity&#8217;s desire to venture into the unknown.</p>\n<p>As the reality of space travel becomes more attainable, stories set beyond Earth&#8217;s orbit have evolved to be more intricate and believable. Among these, Planetes stands out for its genuine portrayal of life in space.</p>\n<p>Planetes distinguishes itself by focusing on the daily lives of space janitors in the Debris Section of a major space corporation, rather than on grandiose interstellar adventures.</p>\n<figure id=\"attachment_1570886\" aria-describedby=\"caption-attachment-1570886\" style=\"width: 1280px\" class=\"wp-caption alignnone\"><img loading=\"lazy\" decoding=\"async\" class=\"size-full wp-image-1570886\" src=\"https://otakukart.com/wp-content/uploads/2024/07/Planetes-.jpg\" alt=\"The 10 Must-Watch Futuristic Anime That Every Fan Should See\" width=\"1280\" height=\"720\" srcset=\"https://otakukart.com/wp-content/uploads/2024/07/Planetes-.jpg 1280w, https://otakukart.com/wp-content/uploads/2024/07/Planetes--770x433.jpg 770w, https://otakukart.com/wp-content/uploads/2024/07/Planetes--150x84.jpg 150w, https://otakukart.com/wp-content/uploads/2024/07/Planetes--750x422.jpg 750w, https://otakukart.com/wp-content/uploads/2024/07/Planetes--1140x641.jpg 1140w\" sizes=\"(max-width: 1280px) 100vw, 1280px\" /><figcaption id=\"caption-attachment-1570886\" class=\"wp-caption-text\">Planetes&nbsp;(Sunrise)</figcaption></figure>\n<p>The characters explore common challenges such as workplace communication issues and job burnout, all while working in the expansive and unfamiliar environment of outer space.</p>\n<p>This down-to-earth approach makes the series feel both relatable and realistic, despite its futuristic premise.</p>\n<p>By highlighting the ordinary aspects of space life, Planetes offers a fresh and unique perspective on space exploration.</p>\n<p>The everyday struggles of its characters resonate with viewers, lending the anime a comforting and familiar vibe.</p>\n<p>Through its focus on the mundane within the extraordinary, Planetes delivers an engaging and distinctive story that amazes and connects with its audience.</p>\n<h3>Memes of the Day</h3>\n<h3>Pixiv 115638104</h3>\n<p><img loading=\"lazy\" decoding=\"async\" class=\"alignnone size-full wp-image-1571070\" src=\"https://otakukart.com/wp-content/uploads/2024/07/fdhgdfhdfhgfhgfghhhhhh-scaled.jpg\" alt=\"\" width=\"2560\" height=\"2177\" srcset=\"https://otakukart.com/wp-content/uploads/2024/07/fdhgdfhdfhgfhgfghhhhhh-scaled.jpg 2560w, https://otakukart.com/wp-content/uploads/2024/07/fdhgdfhdfhgfhgfghhhhhh-770x655.jpg 770w, https://otakukart.com/wp-content/uploads/2024/07/fdhgdfhdfhgfhgfghhhhhh-1536x1306.jpg 1536w, https://otakukart.com/wp-content/uploads/2024/07/fdhgdfhdfhgfhgfghhhhhh-2048x1742.jpg 2048w, https://otakukart.com/wp-content/uploads/2024/07/fdhgdfhdfhgfhgfghhhhhh-150x128.jpg 150w, https://otakukart.com/wp-content/uploads/2024/07/fdhgdfhdfhgfhgfghhhhhh-750x638.jpg 750w, https://otakukart.com/wp-content/uploads/2024/07/fdhgdfhdfhgfhgfghhhhhh-1140x969.jpg 1140w\" sizes=\"(max-width: 2560px) 100vw, 2560px\" /></p>\n<h3>Dokidoki Sentou Bandai | Dokidoki Public Bath Watch Stand [Minamida Usuke]</h3>\n<p><img loading=\"lazy\" decoding=\"async\" class=\"alignnone size-full wp-image-1571256\" src=\"https://otakukart.com/wp-content/uploads/2024/07/sfdgsgsdfgsfdgfdgsdfgsdfgxcvbxcvb.jpg\" alt=\"\" width=\"2048\" height=\"2117\" srcset=\"https://otakukart.com/wp-content/uploads/2024/07/sfdgsgsdfgsfdgfdgsdfgsdfgxcvbxcvb.jpg 2048w, https://otakukart.com/wp-content/uploads/2024/07/sfdgsgsdfgsfdgfdgsdfgsdfgxcvbxcvb-770x796.jpg 770w, https://otakukart.com/wp-content/uploads/2024/07/sfdgsgsdfgsfdgfdgsdfgsdfgxcvbxcvb-1486x1536.jpg 1486w, https://otakukart.com/wp-content/uploads/2024/07/sfdgsgsdfgsfdgfdgsdfgsdfgxcvbxcvb-1981x2048.jpg 1981w, https://otakukart.com/wp-content/uploads/2024/07/sfdgsgsdfgsfdgfdgsdfgsdfgxcvbxcvb-150x155.jpg 150w, https://otakukart.com/wp-content/uploads/2024/07/sfdgsgsdfgsfdgfdgsdfgsdfgxcvbxcvb-750x775.jpg 750w, https://otakukart.com/wp-content/uploads/2024/07/sfdgsgsdfgsfdgfdgsdfgsdfgxcvbxcvb-1140x1178.jpg 1140w\" sizes=\"(max-width: 2048px) 100vw, 2048px\" /></p>\n<p>Read the full post on <a rel=\"nofollow\" href=\"https://otakukart.com/the-10-must-watch-futuristic-anime-that-every-fan-should-see/\">The 10 Must-Watch Futuristic Anime That Every Fan Should See</a></p>\n<h2>Also Read:</h2><ul><li><a href=\"https://otakukart.com/13-true-event-based-korean-dramas-you-need-to-watch-that-bring-koreas-past-to-the-small-screen/\">13 True Event-Based Korean Dramas You Need to Watch That Bring Korea’s Past to the Small Screen</a></li><li><a href=\"https://otakukart.com/must-watch-15-chinese-dramas-on-netflix-offering-romance-adventure-and-historical-epics/\">Must-Watch 15 Chinese Dramas on Netflix Offering Romance, Adventure, and Historical Epics</a></li></ul>",
				},
				Attributes: map[string]*pb.CloudEventAttributeValue{
					"imageurl": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "https://otakukart.com/wp-content/uploads/2024/07/The-10-Must-Watch-Futuristic-Anime-That-Every-Fan-Should-See.jpg",
						},
					},
					"language": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "en",
						},
					},
					"object": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "https://otakukart.com/?p=1570851",
						},
					},
					"objecturl": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "https://otakukart.com/the-10-must-watch-futuristic-anime-that-every-fan-should-see/",
						},
					},
					"summary": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "<div><img width=\"1280\" height=\"720\" src=\"https://otakukart.com/wp-content/uploads/2024/07/The-10-Must-Watch-Futuristic-Anime-That-Every-Fan-Should-See.jpg\" class=\"attachment-medium size-medium wp-post-image\" alt=\"The 10 Must-Watch Futuristic Anime That Every Fan Should See\" decoding=\"async\" srcset=\"https://otakukart.com/wp-content/uploads/2024/07/The-10-Must-Watch-Futuristic-Anime-That-Every-Fan-Should-See.jpg 1280w, https://otakukart.com/wp-content/uploads/2024/07/The-10-Must-Watch-Futuristic-Anime-That-Every-Fan-Should-See-770x433.jpg 770w, https://otakukart.com/wp-content/uploads/2024/07/The-10-Must-Watch-Futuristic-Anime-That-Every-Fan-Should-See-150x84.jpg 150w, https://otakukart.com/wp-content/uploads/2024/07/The-10-Must-Watch-Futuristic-Anime-That-Every-Fan-Should-See-750x422.jpg 750w, https://otakukart.com/wp-content/uploads/2024/07/The-10-Must-Watch-Futuristic-Anime-That-Every-Fan-Should-See-1140x641.jpg 1140w\" sizes=\"(max-width: 1280px) 100vw, 1280px\" /></div><p><img src=\"https://otakukart.com/wp-content/uploads/2024/07/The-10-Must-Watch-Futuristic-Anime-That-Every-Fan-Should-See.jpg\" style=\"display: block; margin: 1em auto\"></p>\n<p>Anime is known for its wide range of stories, each reflecting the boundless creativity of its creators. Among the various genres, speculative fiction stands out for its focus on futuristic themes and ideas. This genre allows for an exploration of what the future might hold, resulting in innovative and impressive stories. By imagining different possibilities, [&#8230;]</p>\n<p>Read the full post on <a rel=\"nofollow\" href=\"https://otakukart.com/the-10-must-watch-futuristic-anime-that-every-fan-should-see/\">The 10 Must-Watch Futuristic Anime That Every Fan Should See</a></p>\n<h2>Also Read:</h2><ul><li><a href=\"https://otakukart.com/top-13-charming-korean-countryside-dramas-filled-with-self-discovery-romance-and-mystery-too/\">Top 13 Charming Korean Countryside Dramas Filled with Self-Discovery, Romance and Mystery Too!</a></li><li><a href=\"https://otakukart.com/12-must-watch-japanese-bl-dramas-featuring-heartwarming-romance-and-emotional-journeys/\">12 Must-Watch Japanese BL Dramas Featuring Heartwarming Romance and Emotional Journeys</a></li></ul>",
						},
					},
					"time": {
						Attr: &pb.CloudEventAttributeValue_CeTimestamp{
							CeTimestamp: timestamppb.New(ts),
						},
					},
					"title": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "The 10 Must-Watch Futuristic Anime That Every Fan Should See",
						},
					},
					"categories": {
						Attr: &pb.CloudEventAttributeValue_CeString{
							CeString: "anime otaku",
						},
					},
				},
			},
			interestId: "interest1",
			follower: &vocab.Actor{
				ID:   "https://mastodon.social/users/johndoe",
				URL:  vocab.IRI("https://mastodon.social/users/johndoe"),
				Name: vocab.DefaultNaturalLanguageValue("John Doe"),
			},
			dst: vocab.Activity{
				ID:      "https://base/2jrVcFeXfGNcExKHLCcrrXBYyLJ",
				URL:     vocab.IRI("https://reader/evt2jrVcFeXfGNcExKHLCcrrXBYyLJ"),
				Type:    "Create",
				Context: vocab.IRI("https://www.w3.org/ns/activitystreams"),
				Actor:   vocab.IRI("https://base/actor/interest1"),
				To: vocab.ItemCollection{
					vocab.IRI("https://mastodon.social/users/johndoe"),
					vocab.IRI("https://www.w3.org/ns/activitystreams#Public"),
				},
				Published: ts,
				Object: &vocab.Object{
					ID:           "https://base/2jrVcFeXfGNcExKHLCcrrXBYyLJ",
					Type:         "Note",
					Name:         vocab.NaturalLanguageValues{},
					AttributedTo: vocab.IRI("https://otakukart.com/feed/"),
					Attachment: vocab.ItemCollection{
						&vocab.Document{
							Type: vocab.ImageType,
							URL:  vocab.IRI("https://otakukart.com/wp-content/uploads/2024/07/The-10-Must-Watch-Futuristic-Anime-That-Every-Fan-Should-See.jpg"),
						},
					},
					Image: &vocab.Link{
						ID:   "https://otakukart.com/wp-content/uploads/2024/07/The-10-Must-Watch-Futuristic-Anime-That-Every-Fan-Should-See.jpg",
						Type: "Link",
					},
					Content: vocab.DefaultNaturalLanguageValue(
						`The 10 Must-Watch Futuristic Anime That Every Fan Should See       <br/> Anime is known for its w...<br/><br/>Original: <a href="https://otakukart.com/the-10-must-watch-futuristic-anime-that-every-fan-should-see/">https://otakukart.com/the-10-must-watch-futuristic-anime-that-every-fan-should-see/</a><br/><br/><a href="https://reader/evt2jrVcFeXfGNcExKHLCcrrXBYyLJ">Event Attributes</a>`),
					Published: ts,
					Replies: &vocab.Collection{
						ID:      "https://otakukart.com/the-10-must-watch-futuristic-anime-that-every-fan-should-see/replies",
						Type:    "Collection",
						Content: vocab.NaturalLanguageValues{},
						Name:    vocab.NaturalLanguageValues{},
						Summary: vocab.NaturalLanguageValues{},
						First: &vocab.CollectionPage{
							Type:    "CollectionPage",
							Content: vocab.NaturalLanguageValues{},
							Name:    vocab.NaturalLanguageValues{},
							Summary: vocab.NaturalLanguageValues{},
							PartOf:  vocab.IRI("https://otakukart.com/the-10-must-watch-futuristic-anime-that-every-fan-should-see/replies"),
							Next:    vocab.IRI("https://otakukart.com/the-10-must-watch-futuristic-anime-that-every-fan-should-see/replies"),
						},
					},
					Tag: vocab.ItemCollection{
						&vocab.Link{
							Type: "Hashtag",
							Name: vocab.DefaultNaturalLanguageValue("#anime"),
							Href: vocab.IRI("https://mastodon.social/tags/anime"),
						},
						&vocab.Link{
							Type: "Hashtag",
							Name: vocab.DefaultNaturalLanguageValue("#otaku"),
							Href: vocab.IRI("https://mastodon.social/tags/otaku"),
						},
						&vocab.Mention{
							Type: "Mention",
							Name: vocab.DefaultNaturalLanguageValue("@@mastodon.social"),
							Href: vocab.IRI("https://mastodon.social/users/johndoe"),
						},
					},
					To: vocab.ItemCollection{
						vocab.IRI("https://mastodon.social/users/johndoe"),
						vocab.IRI("https://www.w3.org/ns/activitystreams#Public"),
					},
					URL: vocab.IRI("https://otakukart.com/the-10-must-watch-futuristic-anime-that-every-fan-should-see/"),
				},
			},
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			a, err := svc.ConvertEventToActivity(context.TODO(), c.src, c.interestId, c.follower, &ts)
			assert.Equal(t, c.dst, a)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestService_ConvertEventToActorUpdate(t *testing.T) {
	svc := NewService("foo", "https://base", "https://reader/evt", vocab.ServiceType)
	svc = NewLogging(svc, slog.Default())
	ts := time.Date(2024, 7, 27, 1, 32, 21, 0, time.UTC)
	cases := map[string]struct {
		src        *pb.CloudEvent
		interestId string
		follower   *vocab.Actor
		dst        vocab.Activity
		err        error
	}{
		"1": {
			src: &pb.CloudEvent{
				Id:          "2jrVcFeXfGNcExKHLCcrrXBYyLJ",
				SpecVersion: CeSpecVersion,
				Source:      "https://awakari.com/reader",
				Type:        "interests-updated",
				Data: &pb.CloudEvent_TextData{
					TextData: "Interest has been updated by its owner.",
				},
			},
			interestId: "interest1",
			follower: &vocab.Actor{
				ID:   "https://mastodon.social/users/johndoe",
				URL:  vocab.IRI("https://mastodon.social/users/johndoe"),
				Name: vocab.DefaultNaturalLanguageValue("John Doe"),
			},
			dst: vocab.Activity{
				ID:      "https://base/2jrVcFeXfGNcExKHLCcrrXBYyLJ-update",
				URL:     vocab.IRI("https://reader/evt2jrVcFeXfGNcExKHLCcrrXBYyLJ"),
				Type:    "Update",
				Context: vocab.IRI("https://www.w3.org/ns/activitystreams"),
				Actor:   vocab.IRI("https://base/actor/interest1"),
				To: vocab.ItemCollection{
					vocab.IRI("https://mastodon.social/users/johndoe"),
					vocab.IRI("https://www.w3.org/ns/activitystreams#Public"),
				},
				Published: ts,
				Object:    vocab.IRI("https://base/actor/interest1"),
				Summary:   vocab.DefaultNaturalLanguageValue("Interest has been updated by its owner."),
			},
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			a, err := svc.ConvertEventToActorUpdate(context.TODO(), c.src, c.interestId, c.follower, &ts)
			assert.Equal(t, c.dst, a)
			assert.ErrorIs(t, err, c.err)
		})
	}
}
