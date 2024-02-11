FROM golang:1.22-alpine3.19 AS builder
WORKDIR /go/src/int-activitypub
COPY . .
RUN \
    apk add protoc protobuf-dev make git && \
    make build

FROM scratch
COPY --from=builder /go/src/int-activitypub/int-activitypub /bin/int-activitypub
ENTRYPOINT ["/bin/int-activitypub"]
