Mastodon account: awakari@mastodon.social
Application: Awakari
Client key	2jipLz3y0-YK5VrT2WO_v7bfU94zQFwTa2d3VsWBl5k
Client secret	1pnnoiOgCIkr_mJO5-ieMHVihNhsnuPYxxLpv-qHbv4
Your access token	rrqIkvp6WfXIMzLZtRLOPGN66fsE0hD8uun-qtkLJ4A

```shell
openssl genrsa 2048 | openssl pkcs8 -topk8 -nocrypt -out private.pem
openssl rsa -in private.pem -outform PEM -pubout -out public.pem
```

```shell
kubectl create secret generic int-activitypub-keys \
  --from-file=public=public.pem \
  --from-file=private=private.pem
```
