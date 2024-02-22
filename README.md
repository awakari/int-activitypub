ActivityPub source implementation for Awakari.

# Conversion Schema

Specific (non as is) attribute conversions:

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
| object.attachment.id        | attachment                       | only if attachment is link                                     |
| object.attachment.url       | attachment                       | only if attachment is object                                   |
| object.attachment.mediaType | attachmenttype                   | only if attachment is object                                   |
| object.content              | `<text data>`                    | Prepends the existing text data (if any) with a line separator |
| object.inReplyTo            | inreplyto                        |
| object.location             | latitude                         |
| object.location             | longitude                        |
| object.startTime            | starts                           |
| object.summary              | `<text data>`                    | Prepends the existing text data (if any) with a line separator |

Notes:

* All other attributes (not mentioned in the table above) are been converted as is, e.g. "duration" -> "duration"

* Activity attribute may be an "object" without an activity type (verb, e.g. "Create"). 
  Then it's also been converted as "object" in addition to the activity fields.

# Compatibility

| Software                                                      | Following                                                            | Delivery |
|---------------------------------------------------------------|----------------------------------------------------------------------|----------|
| [Mastodon](https://en.wikipedia.org/wiki/Mastodon_(software)) | ✅ OK                                                                 | ✅ OK     |
| [Lemmy](https://en.wikipedia.org/wiki/Lemmy_(software))       | ❌ status: 400                                                        | ❌ N/A    |
| [Kbin](https://kbin.socail)                                   | ✅ OK                                                                 | ?        |
| [Pixelfed](https://pixelfed.ru)                               | ✅ OK                                                                 | ?        |
| PeerTube                                                      | ❌ status: 400, "incorrect activity"                                  | ❌ N/A    |
| [Pleroma](https://stereophonic.space)                         | ❌ status 500, message: {"errors":{"detail":"Internal server error"}} | ❌ N/A    |         |
| [Misskey](https://den.raccoon.quest/)                         | ✅ OK                                                                 | ?        |
| [BookWyrm](https://bookwyrm.social)                           | ✅ OK                                                                 | ?        |
| [Friendica](https://venera.social)                            | ✅ OK                                                                 | ?        |
| [Hubzilla](https://libera.site)                               | ✅ OK                                                                 | ✅ OK    |
| [Funkwhale](https://funkwhale.our-space.xyz)                  | ❌ status 500                                                         | ❌ N/A   |          

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
