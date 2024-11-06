package http

import (
	"github.com/bytedance/sonic"
	vocab "github.com/go-ap/activitypub"
	"hash/crc32"
)

const ContentTypeActivity = "application/activity+json; charset=utf-8"

var contextExtMastodon = map[string]any{
	"manuallyApprovesFollowers": "as:manuallyApprovesFollowers",
	"toot":                      "http://joinmastodon.org/ns#",
	"featured": map[string]any{
		"@id":   "toot:featured",
		"@type": "@id",
	},
	"featuredTags": map[string]any{
		"@id":   "toot:featuredTags",
		"@type": "@id",
	},
	"alsoKnownAs": map[string]any{
		"@id":   "as:alsoKnownAs",
		"@type": "@id",
	},
	"movedTo": map[string]any{
		"@id":   "as:movedTo",
		"@type": "@id",
	},
	"schema":           "http://schema.org#",
	"PropertyValue":    "schema:PropertyValue",
	"value":            "schema:value",
	"discoverable":     "toot:discoverable",
	"Device":           "toot:Device",
	"Ed25519Signature": "toot:Ed25519Signature",
	"Ed25519Key":       "toot:Ed25519Key",
	"Curve25519Key":    "toot:Curve25519Key",
	"EncryptedMessage": "toot:EncryptedMessage",
	"publicKeyBase64":  "toot:publicKeyBase64",
	"deviceId":         "toot:deviceId",
	"claim": map[string]any{
		"@type": "@id",
		"@id":   "toot:claim",
	},
	"fingerprintKey": map[string]any{
		"@type": "@id",
		"@id":   "toot:fingerprintKey",
	},
	"identityKey": map[string]any{
		"@type": "@id",
		"@id":   "toot:identityKey",
	},
	"devices": map[string]any{
		"@type": "@id",
		"@id":   "toot:devices",
	},
	"messageFranking": "toot:messageFranking",
	"messageType":     "toot:messageType",
	"cipherText":      "toot:cipherText",
	"suspended":       "toot:suspended",
	"memorial":        "toot:memorial",
	"indexable":       "toot:indexable",
	"focalPoint": map[string]any{
		"@container": "@list",
		"@id":        "toot:focalPoint",
	},
}

func FixContext(obj vocab.ActivityObject) (m map[string]any, checkSum uint32) {
	d, _ := sonic.Marshal(obj)
	checkSum = crc32.ChecksumIEEE(d)
	m = make(map[string]any)
	_ = sonic.Unmarshal(d, &m)
	c, ok := m["context"]
	switch obj.(type) {
	case vocab.Actor:
		c = append(c.([]any), contextExtMastodon)
	}
	if ok {
		m["@context"] = c
		delete(m, "context")
	}
	return
}
