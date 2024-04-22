# About

ActivityPub source implementation for Awakari. Actually, just another Activitypub server that follows specified publishers.

Awakari is a service consuming public updates from various sources and filters these for a user.
The purpose is only to notify user in real time and provide a link to the source (e.g. post). 
If you don't want Awakari to follow you, just find it in a list of your followers and block.

TODO sources blacklist

# Conversion Schema

Specific (non "as is") attribute conversions:

| Source Activity Attribute   | Destination CloudEvent Attribute | Notes                                                          |
|-----------------------------|----------------------------------|----------------------------------------------------------------|
| actor.id                    | source                           |                                                                |
| actor.name                  | subject                          | e.g. "John Doe"                                                |
| published                   | time                             |                                                                |
| content                     | `<text data>`                    |                                                                |
| summary                     | `<text data>`                    | Prepends the existing text data (if any) with a line separator |
| type                        | action                           | e.g. "Create"                                                  |
| object.id                   | objecturl                        | only if object is link                                         |
| object.type                 | object                           | e.g. "Note"                                                    |
| object.attachment.id        | attachmenturl                    | only if attachment is link                                     |
| object.attachment.url       | attachmenturl                    | only if attachment is object                                   |
| object.attachment.mediaType | attachmenttype                   | only if attachment is object                                   |
| object.content              | `<text data>`                    | Prepends the existing text data (if any) with a line separator |
| object.inReplyTo            | inreplyto                        |
| object.location             | latitude                         |
| object.location             | longitude                        |
| object.startTime            | starts                           |
| object.summary              | `<text data>`                    | Prepends the existing text data (if any) with a line separator |
| object.image                | imageurl                         |                                                                |

Notes:

* All other attributes (not mentioned in the table above) are been converted as is, e.g. "duration" -> "duration"

* Activity attribute may be an "object" without an activity type (verb, e.g. "Create"). 
  Then it's also been converted as "object" in addition to the activity fields.

# Compatibility

| Software                                                      | Following                             | Delivery |
|---------------------------------------------------------------|---------------------------------------|----------|
| [Mastodon](https://en.wikipedia.org/wiki/Mastodon_(software)) | ✅ OK                                  | ✅ OK     |
| [Lemmy](https://en.wikipedia.org/wiki/Lemmy_(software))       | ❌ status: 400                         | ❌ N/A    |
| [Kbin](https://kbin.socail)                                   | ✅ OK                                  | ?        |
| [Pixelfed](https://pixelfed.ru)                               | ✅ OK                                  | ?        |
| PeerTube                                                      | ❌ status: 400, "incorrect activity"   | ❌ N/A    |
| [Pleroma](https://stereophonic.space)                         | ❌ status 500, "Internal server error" | ❌ N/A    |         |
| [Misskey](https://den.raccoon.quest/)                         | ✅ OK                                  | ?        |
| [BookWyrm](https://bookwyrm.social)                           | ✅ OK                                  | ?        |
| [Friendica](https://venera.social)                            | ✅ OK                                  | ✅ OK     |
| [Hubzilla](https://libera.site)                               | ✅ OK                                  | ✅ OK     |
| [Funkwhale](https://funkwhale.our-space.xyz)                  | ❌ status 500                          | ❌ N/A    |          

# Other

## Public Key

```shell
openssl genrsa 2048 | openssl pkcs8 -topk8 -nocrypt -out private.pem
openssl rsa -in private.pem -outform PEM -pubout -out public.pem
```

```shell
kubectl create secret generic int-activitypub-keys \
  --from-file=public=public.pem \
  --from-file=private=private.pem
```

```shell
kubectl create secret generic int-activitypub-search-client-mastodon \
  --from-literal=key=key1 \
  --from-literal=secret=secret1 \
  --from-literal=token=token1
```

## Manual Testing

Example request:
```shell
grpcurl \
  -plaintext \
  -proto api/grpc/service.proto \
  -d '{ "addr": "Mastodon@mastodon.social" }' \
  localhost:50051 \
  awakari.int.activitypub.Service/Create
```
