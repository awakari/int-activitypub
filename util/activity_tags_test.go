package util

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestActivityHasNoBotTag(t *testing.T) {
	cases := map[string]struct {
		in   string
		tags ActivityTags
	}{
		"nobot": {
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
			tags: ActivityTags{
				Tag: []ActivityTag{
					{
						Name: "#nobot",
					},
				},
			},
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			var tags ActivityTags
			err := json.Unmarshal([]byte(c.in), &tags)
			require.Nil(t, err)
			assert.Equal(t, c.tags, tags)
		})
	}
}
