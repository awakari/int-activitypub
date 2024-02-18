```shell
openssl genrsa 2048 | openssl pkcs8 -topk8 -nocrypt -out private.pem
openssl rsa -in private.pem -outform PEM -pubout -out public.pem
```

```shell
kubectl create secret generic int-activitypub-keys \
  --from-file=public=public.pem \
  --from-file=private=private.pem
```

Example request:
```shell
grpcurl \
  -plaintext \
  -proto api/grpc/service.proto \
  -d '{ "addr": "Mastodon@mastodon.social" }' \
  localhost:50051 \
  awakari.int.activitypub.Service/Create
```

TODO conversion schema
